package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"devport-lab/utils"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Pushes the dependency cache for the current git commit to the cloud.",
	Long: `Scans the 'node_modules' directory, uploads any new file objects to the S3
cache, and then uploads a manifest file tagged with the current git commit hash.`,
	Run: func(cmd *cobra.Command, args []string) {
		pushCache() 
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func pushCache() {
	commitHash, err := utils.GetCurrentCommitHash()
	if err != nil {
		logger.Error("Failed to get current commit hash", "error", err)
		os.Exit(1)
	}
	logger.Info("Current commit hash", "hash", commitHash)

	rootDir := viper.GetString("root_directory")
	s3Bucket := viper.GetString("s3.bucket") 

	s3Client, err := newS3Client()
	if err != nil {
		logger.Error("Failed to create S3 client", "error", err)
		os.Exit(1)
	}

	manifest := make(map[string]string)
	logger.Info("Starting scan", "directory", rootDir)

	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				logger.Warn("Could not open file, skipping", "path", path, "error", err)
				return nil
			}
			defer file.Close()

			hash := sha256.New()
			if _, err := io.Copy(hash, file); err != nil {
				logger.Warn("Could not hash file, skipping", "path", path, "error", err)
				return nil
			}
			hashString := fmt.Sprintf("%x", hash.Sum(nil))
			manifest[path] = hashString

			_, err = s3Client.HeadObject(context.TODO(), &s3.HeadObjectInput{
				Bucket: aws.String(s3Bucket),
				Key:    aws.String(hashString),
			})
			if err == nil {
				logger.Debug("Object already exists in cache, skipping upload", "hash", hashString)
				return nil
			}
			logger.Debug("Object not found in cache, uploading", "hash", hashString)
			file.Seek(0, 0) // Reset file pointer to the beginning
			_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
				Bucket:      aws.String(s3Bucket),
				Key: 	   aws.String(hashString),
				Body:        file,
			})
			if err != nil {
				logger.Warn("Failed to upload file to S3", "path", path, "error", err)
			} else {
				logger.Debug("Successfully uploaded new object", "hash", hashString)
			}
		}
		return nil
	})

	if err != nil {
		logger.Error("Error walking the path", "path", rootDir, "error", err)
		os.Exit(1)
	}

	logger.Info("Scan complete. Generating manifest file...")



	jsonData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		logger.Error("Error generating JSON", "error", err)
		os.Exit(1)
	}
	manifestKey := "manifests/" + commitHash + ".json"

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(manifestKey),
		Body:   bytes.NewReader(jsonData),
	})
	if err != nil {
		logger.Error("Error uploading manifest to S3", "bucket", s3Bucket, "key", manifestKey, "error", err)
		os.Exit(1)
	}
	logger.Info("Manifest uploaded successfully", "bucket", s3Bucket, "key", manifestKey)
	logger.Info("Push complete. Cache updated for commit", "commit", commitHash)
}