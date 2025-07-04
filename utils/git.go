package utils

import (
	"bytes" 
	"errors"
	"fmt"
	"io"
	"os"
	"crypto/sha256"
	"os/exec"
	"strings"
)

func GetCurrentCommitHash() (string, error) {

	cmd := exec.Command("git", "rev-parse", "HEAD")

	var out bytes.Buffer
	var stderr bytes.Buffer
	
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	
	err := cmd.Run()

	if err != nil {
		if strings.Contains(stderr.String(), "not a git repository") {
			return "", errors.New("current directory is not a Git repository")
		}
		return "", errors.New("failed to execute git command: " + stderr.String())
	}

	commitHash := strings.TrimSpace(out.String())
	
	return commitHash, nil
}

func GetLockFileHash() (string, error) {
	lockFiles := []string{"pnpm-lock.yaml", "yarn.lock", "package-lock.json"}
	var lockFilePath string
	
	for _, file := range lockFiles {
		if _, err := os.Stat(file); err == nil {
			lockFilePath = file
			break
		}
	}

	if lockFilePath == "" {
		return "", errors.New("no lock file found (pnpm-lock.yaml, yarn.lock, or package-lock.json)")
	}

	file, err := os.Open(lockFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open lock file %s: %w", lockFilePath, err)
	}
	defer file.Close()

	hash := sha256.New()
	
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to hash lock file %s: %w", lockFilePath, err)
	}
	
	hashString := fmt.Sprintf("%x", hash.Sum(nil))

	return hashString, nil
}