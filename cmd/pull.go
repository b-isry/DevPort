package cmd

import (
	"errors"
	"context"
	"devport-lab/utils"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pulls the dependency cache for the current git commit from the cloud.",
	Long: `Checks the current git commit hash, fetches the corresponding manifest from
the S3 cache, and then reconstructs the 'node_modules' directory by downloading
the necessary file objects.`,
	Run: func(cmd *cobra.Command, args []string) {
		pullCache()
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

func pullCache() {
	lockFileHash, err := utils.GetLockFileHash()
	if err != nil {
		logger.Error("Failed to get lock file hash", "error", err)
		os.Exit(1)
	}
	logger.Info("Pulling cache for dependency signature", "hash", lockFileHash)

	commitHash, err := utils.GetCurrentCommitHash()
	if err != nil {
		logger.Error("Failed to get current commit hash", "error", err)
		os.Exit(1)
	}
	logger.Info("Current commit hash", "hash", commitHash)

	rootDir := viper.GetString("root_directory")
	bucket := viper.GetString("s3.bucket")
	
	logger.Info("Starting rebuild from cache...")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(viper.GetString("s3.region")),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			viper.GetString("s3.access_key_id"),
			viper.GetString("s3.secret_access_key"),
			"",
		)),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           viper.GetString("s3.endpoint"),
					SigningRegion: viper.GetString("s3.region"),
					Source:        aws.EndpointSourceCustom,
				}, nil
			},
		)),
		)
		if err != nil {
			logger.Error("Failed to load AWS configuration", "error", err)
			os.Exit(1)
		}
		s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		})


	manifestKey := "manifests/" + lockFileHash + ".json"

	logger.Info("Fetching remote manifest", "path", manifestKey)

	result, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(manifestKey),
	})
	if err != nil {
		var notFoundError *types.NoSuchKey
		if errors.As(err, &notFoundError) {
			logger.Error("Cache miss. No cache found for this dependency signature.", "hash", lockFileHash)
			logger.Info("Please run 'npm install' and then 'devport push' on the source machine to populate the cache.")
		} else {
			logger.Error("Failed to download manifest from S3", "path", manifestKey, "error", err)
		}
		os.Exit(1)
	}
	defer result.Body.Close()

	manifestData, err := io.ReadAll(result.Body)
	if err != nil {
		logger.Error("Failed to read manifest data from S3 response", "error", err)
		os.Exit(1)
	}

	var manifest map[string]string
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		logger.Error("Failed to parse manifest JSON", "error", err)
		os.Exit(1)
	}
	logger.Info("Manifest found. Rebuilding node_modules...")
	logger.Info("Removing old directory", "path", rootDir)
	if err := os.RemoveAll(rootDir); err != nil {
		logger.Error("Failed to remove old directory", "path", rootDir, "error", err)
		os.Exit(1)
	}

	for destinationPath, hashString := range manifest {
		logger.Debug("Rebuilding file", "path", destinationPath)

		result, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(hashString),
		})
		if err != nil {
			logger.Warn("Failed to retrieve object from S3, skipping", "bucket", bucket, "key", hashString, "error", err)
			continue
		}
		defer result.Body.Close()

		destinationDir := filepath.Dir(destinationPath)
		if err := os.MkdirAll(destinationDir, 0755); err != nil {
			logger.Error("Failed to create directory during rebuild", "path", destinationDir, "error", err)
			os.Exit(1) 
		}

		destFile, err := os.Create(destinationPath)
		if err != nil {
			logger.Warn("Failed to create destination file, skipping", "path", destinationPath, "error", err)
			continue
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, result.Body); err != nil {
			logger.Warn("Failed to write downloaded content to file", "path", destinationPath, "error", err)
		}
	}

	logger.Info("Rebuild complete.")
}