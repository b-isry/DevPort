// cmd/scan.go
package cmd

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	cacheDir := viper.GetString("cache_directory")
	manifestFile := viper.GetString("manifest_file")

	manifest := make(map[string]string)
	// MODIFIED: Use our new logger
	logger.Info("Starting scan", "directory", rootDir)

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		// MODIFIED: Structured error logging
		logger.Error("Failed to create cache directory", "path", cacheDir, "error", err)
		os.Exit(1)
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				// MODIFIED: Use WARN for skippable errors
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

			cachedFilePath := filepath.Join(cacheDir, hashString)
			if _, err := os.Stat(cachedFilePath); os.IsNotExist(err) {
				file.Seek(0, 0)
				newCacheFile, err := os.Create(cachedFilePath)
				if err != nil {
					logger.Warn("Could not create cache file", "path", path, "error", err)
					return nil
				}
				defer newCacheFile.Close()
				io.Copy(newCacheFile, file)
				// MODIFIED: Use DEBUG for verbose messages
				logger.Debug("Cached new object", "path", path)
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