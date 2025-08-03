package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunCommand executes a command and returns any error
func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunCommandWithOutput executes a command and returns its output
func RunCommandWithOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// MoveDirectoryContents moves all files and directories from src to dst
func MoveDirectoryContents(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %v", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if err := os.Rename(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to move %s to %s: %v", srcPath, dstPath, err)
		}
	}

	return nil
}

// ExtractTarContent extracts tar content to the specified destination
func ExtractTarContent(tarContent []byte, destDir string) error {
	tempFile, err := os.CreateTemp("", "poon-download-*.tar")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := tempFile.Write(tarContent); err != nil {
		return fmt.Errorf("failed to write tar content: %v", err)
	}

	cmd := exec.Command("tar", "-xf", tempFile.Name(), "-C", destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract tar: %v", err)
	}

	return nil
}

// SyncFromRemote pulls the latest changes from the remote git repository
func SyncFromRemote() error {
	fmt.Printf("Syncing with remote repository...\n")

	if err := RunCommand("git", "fetch", "origin"); err != nil {
		return fmt.Errorf("failed to fetch from remote: %v", err)
	}

	if err := RunCommand("git", "merge", "origin/main", "--no-edit"); err != nil {
		fmt.Printf("Merge failed, attempting rebase...\n")
		if err := RunCommand("git", "reset", "--hard", "HEAD"); err != nil {
			return fmt.Errorf("failed to reset: %v", err)
		}
		if err := RunCommand("git", "rebase", "origin/main"); err != nil {
			return fmt.Errorf("failed to rebase: %v", err)
		}
	}

	return nil
}

// GitPull pulls the latest changes from the specified branch
func GitPull(remote, branch string) error {
	return RunCommand("git", "pull", remote, branch)
}

// GitPush pushes local changes to the remote repository
func GitPush(remote, branch string) error {
	return RunCommand("git", "push", remote, branch)
}
