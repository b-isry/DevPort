// cmd/scan.go
package cmd

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)


// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scans node_modules, populates the cache, and creates a manifest.",
	Long: `Scans the 'node_modules' directory file by file. For each file, it calculates
a SHA-256 hash, stores a deduplicated copy in the local cache, and generates
a manifest.json file mapping all file paths to their content hash.`,
	Run: func(cmd *cobra.Command, args []string) {
		scanAndCache() // This is the function call to our logic
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func scanAndCache() {
	scanRootDir := viper.GetString("root_directory")
    scanCacheDir := viper.GetString("cache_directory")
    scanManifestFile := viper.GetString("manifest_file")

	manifest := make(map[string]string)
	fmt.Printf("Scanning directory: %s\n", scanRootDir)

	if err := os.MkdirAll(scanCacheDir, 0755); err != nil {
		log.Fatalf("Failed to create cache directory: %v", err)
	}

	err := filepath.Walk(scanRootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				log.Printf("Could not open file %s: %v", path, err)
				return nil
			}
			defer file.Close()

			hash := sha256.New()
			if _, err := io.Copy(hash, file); err != nil {
				log.Printf("Could not hash file %s: %v", path, err)
				return nil
			}
			hashString := fmt.Sprintf("%x", hash.Sum(nil))
			manifest[path] = hashString

			cachedFilePath := filepath.Join(scanCacheDir, hashString)
			if _, err := os.Stat(cachedFilePath); os.IsNotExist(err) {
				file.Seek(0, 0)
				newCacheFile, err := os.Create(cachedFilePath)
				if err != nil {
					log.Printf("Could not create cache file for %s: %v", path, err)
					return nil
				}
				defer newCacheFile.Close()
				io.Copy(newCacheFile, file)
				fmt.Printf("  [CACHED] %s\n", path)
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error walking the path %s: %v", scanRootDir, err)
	}

	fmt.Println("\nScan complete. Generating manifest file...")

	jsonData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		log.Fatalf("Error generating JSON: %v", err)
	}
	err = os.WriteFile(scanManifestFile, jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing manifest file: %v", err)
	}
	fmt.Printf("Successfully created %s\n", scanManifestFile)
}