// cmd/rebuild.go
package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

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
	cacheDir := viper.GetString("cache_directory")
	manifestFile := viper.GetString("manifest_file")
	
	logger.Info("Starting rebuild from cache...")

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
		logger.Debug("Rebuilding file", "path", destinationPath) // A good debug message
		sourcePath := filepath.Join(cacheDir, hashString)
		destinationDir := filepath.Dir(destinationPath)
		if err := os.MkdirAll(destinationDir, 0755); err != nil {
			logger.Error("Failed to create directory during rebuild", "path", destinationDir, "error", err)
			os.Exit(1) // This is a fatal error
		}

		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			logger.Error("Cache is corrupt. Could not open cache file.", "path", sourcePath, "error", err)
			os.Exit(1) // Also fatal, the cache is broken
		}
		defer sourceFile.Close()

		destFile, err := os.Create(destinationPath)
		if err != nil {
			logger.Warn("Failed to create destination file, skipping", "path", destinationPath, "error", err)
			continue
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, sourceFile); err != nil {
			logger.Warn("Failed to copy content, file may be corrupt", "path", destinationPath, "error", err)
		}
	}

	logger.Info("Rebuild complete.")
}