package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuilds node_modules from the local cache using a manifest.",
	Long: `Deletes the current node_modules directory, then reconstructs it
by reading the manifest.json file and copying each required file
from the local cache to its correct destination.`,
	Run: func(cmd *cobra.Command, args []string) {
		rebuildFromCache()
	},
}

func init() {
	rootCmd.AddCommand(rebuildCmd)
}

func rebuildFromCache() {
	rootDir := viper.GetString("root_directory")
	manifestFile := viper.GetString("manifest_file")
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

	manifestData, err := os.ReadFile(manifestFile)
	if err != nil {
		logger.Error("Failed to read manifest file. Please run 'scan' first.", "path", manifestFile, "error", err)
		os.Exit(1)
	}

	var manifest map[string]string
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		logger.Error("Failed to parse manifest JSON", "error", err)
		os.Exit(1)
	}

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
			continue // Skip this file if it cannot be retrieved
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