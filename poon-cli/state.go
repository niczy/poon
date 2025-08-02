package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type WorkspaceState struct {
	TrackedPaths map[string]*TrackedPathState `json:"trackedPaths"`
	LastSync     time.Time                    `json:"lastSync"`
}

type TrackedPathState struct {
	Path         string            `json:"path"`
	Files        map[string]string `json:"files"` // filename -> hash
	LastSyncHash string            `json:"lastSyncHash"`
	AddedAt      time.Time         `json:"addedAt"`
	LastSyncAt   time.Time         `json:"lastSyncAt"`
}

func loadWorkspaceState() (*WorkspaceState, error) {
	statePath := ".poon/state.json"
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return &WorkspaceState{
			TrackedPaths: make(map[string]*TrackedPathState),
			LastSync:     time.Time{},
		}, nil
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %v", err)
	}

	var state WorkspaceState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %v", err)
	}

	if state.TrackedPaths == nil {
		state.TrackedPaths = make(map[string]*TrackedPathState)
	}

	return &state, nil
}

func saveWorkspaceState(state *WorkspaceState) error {
	if err := os.MkdirAll(".poon", 0755); err != nil {
		return fmt.Errorf("failed to create .poon directory: %v", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %v", err)
	}

	statePath := ".poon/state.json"
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %v", err)
	}

	return nil
}

func calculateDirectoryHash(dirPath string) (map[string]string, error) {
	files := make(map[string]string)

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Skip .git and other hidden directories
		if strings.Contains(path, "/.") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", path, err)
		}

		hash := sha256.Sum256(content)
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		files[relPath] = fmt.Sprintf("%x", hash)
		return nil
	})

	return files, err
}

func generatePatchForPath(pathState *TrackedPathState, currentPath string) (string, error) {
	// Calculate current file hashes
	currentFiles, err := calculateDirectoryHash(currentPath)
	if err != nil {
		return "", fmt.Errorf("failed to calculate current directory hash: %v", err)
	}

	// TODO: Generate unified diff patch by comparing:
	// - pathState.Files (original state from monorepo)
	// - currentFiles (current local state)

	// For now, return a placeholder
	patch := fmt.Sprintf("# Patch for %s\n# Original files: %d\n# Current files: %d\n",
		pathState.Path, len(pathState.Files), len(currentFiles))

	return patch, nil
}

func updateTrackedPathState(state *WorkspaceState, path string) error {
	pathState, exists := state.TrackedPaths[path]
	if !exists {
		return fmt.Errorf("path %s is not tracked", path)
	}

	// Update file hashes to current state
	currentFiles, err := calculateDirectoryHash(path)
	if err != nil {
		return fmt.Errorf("failed to calculate directory hash: %v", err)
	}

	pathState.Files = currentFiles
	pathState.LastSyncAt = time.Now()

	return nil
}
