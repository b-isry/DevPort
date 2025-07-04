// utils/git.go
package utils

import (
	"bytes" 
	"errors" 
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