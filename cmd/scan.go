// cmd/scan.go
package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scans node_modules, populates the cache, and creates a manifest.",
	Long: `Scans the 'node_modules' directory file by file. For each file, it calculates
a SHA-256 hash, stores a deduplicated copy in the local cache, and generates
a manifest.json file mapping all file paths to their content hash.`,
	Run: func(cmd *cobra.Command, args []string) {
		scanAndCache() 
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func scanAndCache() {
	rootDir := viper.GetString("root_directory")
	manifestFile := viper.GetString("manifest_file")
	s3Bucket := viper.GetString("s3.bucket") 

	s3Client, err := newS3Client()
	if err != nil {
		logger.Error("Failed to create S3 client", "error", err)
		os.Exit(1)
	}

	manifest := make(map[string]string)
	logger.Info("Starting scan", "directory", rootDir)

	// if err := os.MkdirAll(cacheDir, 0755); err != nil {
	// 	logger.Error("Failed to create cache directory", "path", cacheDir, "error", err)
	// 	os.Exit(1)
	// }

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
	err = os.WriteFile(manifestFile, jsonData, 0644)
	if err != nil {
		logger.Error("Error writing manifest file", "path", manifestFile, "error", err)
		os.Exit(1)
	}
	logger.Info("Successfully created manifest", "path", manifestFile)
}