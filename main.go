package main

import (
	"crypto/sha256" 
	"encoding/json"
	"fmt"            
	"io"             
	"log"           
	"os"             
	"path/filepath"  
)

const(
		rootDir = "node_modules"
		cacheDir = ".devport_cache/objects"
		manifestFile = "manifest.json"
	)

func main() {	


	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run main.go [scan|rebuild]")
	}

	command := os.Args[1]

	switch command {
	case "scan":
		scanAndCache()
	case "rebuild":
		rebuildFromCach()
	default:
		log.Fatalf("Unknown command: %s. Use 'scan' or 'rebuild'.", command)
	}

}

func scanAndCache() {
	manifest := make(map[string]string)
	fmt.Printf("Scanning directory: %s\n\n", rootDir)

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Fatalf("Failed to create cache directory: %v", err)
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
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

			hashInBytes := hash.Sum(nil)

			hashString := fmt.Sprintf("%x", hashInBytes)

			manifest[path] = hashString

			cacheFilePath := filepath.Join(cacheDir, hashString)

			if _, err := os.Stat(cacheFilePath); os.IsNotExist(err) {
				file.Seek(0, 0) 
				newCacheFil, err := os.Create(cacheFilePath)
				if err != nil {
					log.Printf("Could not create cache file %s: %v", cacheFilePath, err)
					return nil
				}
				defer newCacheFil.Close()

				io.Copy(newCacheFil, file)
				fmt.Printf(" [CACHED] %s\n", path)
			}

		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error walking the path %s: %v", rootDir, err)
	}

	fmt.Println("\nScan complete.")


	jsonData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		log.Fatalf("Error generating JSON: %v", err)
	}

	err = os.WriteFile("manifest.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}

	fmt.Println("Successfully created manifest.json")
}

func rebuildFromCach() {
	fmt.Println("starting to rebuild from cache...")

	manifestData, err := os.ReadFile(manifestFile)
	if err != nil {
		log.Fatalf("Failed to read manifest file %s: %v. Please run 'scan' first.", manifestFile, err)
	}
	var manifest map[string]string
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		log.Fatalf("Failed to parse manifest file %s: %v", manifestFile, err)
	}

	fmt.Printf("removing old directory: %s\n", rootDir)
	if err := os.RemoveAll(rootDir); err != nil {
		log.Fatalf("Failed to remove old directory %s: %v", rootDir, err)
	}

	for destinationPath, hashString := range manifest {
		sourcePath := filepath.Join(cacheDir, hashString)
		destinationDir := filepath.Dir(destinationPath)
		if err := os.MkdirAll(destinationDir, 0755); err != nil {
			log.Printf("Failed to create directory %s: %v", destinationDir, err)
		}
		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			log.Printf("Failed to open source file %s: %v", sourcePath, err)
			continue //continue will be removed 
		}
		defer sourceFile.Close()

		destFile, err := os.Create(destinationPath)
		if err != nil {
			log.Printf("Failed to create destination file %s: %v", destinationPath, err)
			continue //continue will be removed
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, sourceFile); err != nil {
			log.Printf("Failed to copy file from %s to %s: %v", sourcePath, destinationPath, err)
		}
	}
	fmt.Println("Rebuild from cache complete.")
}