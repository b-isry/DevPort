// cmd/rebuild.go
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)


// rebuildCmd represents the rebuild command
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

// This is our original rebuildFromCache function.
func rebuildFromCache() {
	rebuildRootDir := viper.GetString("root_directory")
	rebuildCacheDir := viper.GetString("cache_directory")	
	rebuildManifestFile := viper.GetString("manifest_file")

	fmt.Println("Starting rebuild from cache...")

	manifestData, err := os.ReadFile(rebuildManifestFile)
	if err != nil {
		log.Fatalf("Failed to read manifest file %s: %v. Please run 'scan' first.", rebuildManifestFile, err)
	}

	var manifest map[string]string
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		log.Fatalf("Failed to parse manifest JSON: %v", err)
	}

	fmt.Printf("Removing old directory: %s\n", rebuildRootDir)
	if err := os.RemoveAll(rebuildRootDir); err != nil {
		log.Fatalf("Failed to remove old %s: %v", rebuildRootDir, err)
	}

	for destinationPath, hashString := range manifest {
		sourcePath := filepath.Join(rebuildCacheDir, hashString)
		destinationDir := filepath.Dir(destinationPath)
		if err := os.MkdirAll(destinationDir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", destinationDir, err)
		}

		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			log.Printf("FATAL: Cache is corrupt. Could not open cache file %s: %v", sourcePath, err)
			continue
		}
		defer sourceFile.Close()

		destFile, err := os.Create(destinationPath)
		if err != nil {
			log.Printf("Failed to create destination file %s: %v", destinationPath, err)
			continue
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, sourceFile); err != nil {
			log.Printf("Failed to copy content to %s: %v", destinationPath, err)
		}
	}

	fmt.Println("\nRebuild complete.")
}