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

func main() {
	// For now, we'll hardcode it. Later, this will come from a command-line argument.
	rootDir := "node_modules"
	manifest := make(map[string]string)

	fmt.Printf("Scanning directory: %s\n\n", rootDir)

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