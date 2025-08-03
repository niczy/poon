package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
)

// VersionManager implements VersionStore interface
type VersionManager struct {
	backend StorageBackend
}

// NewVersionManager creates a new version manager
func NewVersionManager(backend StorageBackend) *VersionManager {
	return &VersionManager{
		backend: backend,
	}
}

// GetCurrentVersion returns the current version number
func (vm *VersionManager) GetCurrentVersion(ctx context.Context) (int64, error) {
	data, err := vm.backend.Get(ctx, "version/current")
	if err != nil {
		// No versions exist yet, start at 0
		return 0, nil
	}

	version, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse current version: %w", err)
	}

	return version, nil
}

// GetVersionInfo returns version information for a specific version
func (vm *VersionManager) GetVersionInfo(ctx context.Context, version int64) (*VersionInfo, error) {
	key := fmt.Sprintf("version/info/%d", version)
	data, err := vm.backend.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("version %d not found: %w", version, err)
	}

	var info VersionInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal version info: %w", err)
	}

	return &info, nil
}

// GetLatestVersionInfo returns the latest version information
func (vm *VersionManager) GetLatestVersionInfo(ctx context.Context) (*VersionInfo, error) {
	currentVersion, err := vm.GetCurrentVersion(ctx)
	if err != nil {
		return nil, err
	}

	if currentVersion == 0 {
		return nil, fmt.Errorf("no versions exist")
	}

	return vm.GetVersionInfo(ctx, currentVersion)
}

// CreateVersion creates a new version pointing to a commit
func (vm *VersionManager) CreateVersion(ctx context.Context, commitHash Hash, message string) (*VersionInfo, error) {
	// Get next version number
	currentVersion, err := vm.GetCurrentVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	newVersion := currentVersion + 1
	now := time.Now()

	// Create version info
	info := &VersionInfo{
		Version:    newVersion,
		CommitHash: commitHash,
		Timestamp:  now,
		Message:    message,
	}

	// Store version info
	infoData, err := json.Marshal(info)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal version info: %w", err)
	}

	infoKey := fmt.Sprintf("version/info/%d", newVersion)
	if err := vm.backend.Put(ctx, infoKey, infoData); err != nil {
		return nil, fmt.Errorf("failed to store version info: %w", err)
	}

	// Update current version
	currentData := []byte(strconv.FormatInt(newVersion, 10))
	if err := vm.backend.Put(ctx, "version/current", currentData); err != nil {
		return nil, fmt.Errorf("failed to update current version: %w", err)
	}

	// Store commit hash mapping for quick lookup
	hashKey := fmt.Sprintf("version/hash/%s", commitHash)
	versionData := []byte(strconv.FormatInt(newVersion, 10))
	if err := vm.backend.Put(ctx, hashKey, versionData); err != nil {
		return nil, fmt.Errorf("failed to store commit hash mapping: %w", err)
	}

	return info, nil
}

// ListVersions returns all versions in chronological order
func (vm *VersionManager) ListVersions(ctx context.Context, limit int) ([]*VersionInfo, error) {
	keys, err := vm.backend.List(ctx, "version/info/")
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}

	// Extract version numbers and sort them
	var versions []int64
	for _, key := range keys {
		// Extract version number from key "version/info/123"
		if len(key) > 13 {
			versionStr := key[13:]
			version, err := strconv.ParseInt(versionStr, 10, 64)
			if err != nil {
				continue // Skip invalid version keys
			}
			versions = append(versions, version)
		}
	}

	// Sort versions in descending order (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] > versions[j]
	})

	// Apply limit
	if limit > 0 && len(versions) > limit {
		versions = versions[:limit]
	}

	// Fetch version info for each version
	var result []*VersionInfo
	for _, version := range versions {
		info, err := vm.GetVersionInfo(ctx, version)
		if err != nil {
			continue // Skip corrupted version info
		}
		result = append(result, info)
	}

	return result, nil
}

// GetVersionByCommit returns the version number for a commit hash
func (vm *VersionManager) GetVersionByCommit(ctx context.Context, commitHash Hash) (int64, error) {
	key := fmt.Sprintf("version/hash/%s", commitHash)
	data, err := vm.backend.Get(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("commit hash not found: %w", err)
	}

	version, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse version: %w", err)
	}

	return version, nil
}

// DeleteVersion removes a version (for cleanup or rollback)
func (vm *VersionManager) DeleteVersion(ctx context.Context, version int64) error {
	// Get version info first to get commit hash
	info, err := vm.GetVersionInfo(ctx, version)
	if err != nil {
		return fmt.Errorf("version not found: %w", err)
	}

	// Delete version info
	infoKey := fmt.Sprintf("version/info/%d", version)
	if err := vm.backend.Delete(ctx, infoKey); err != nil {
		return fmt.Errorf("failed to delete version info: %w", err)
	}

	// Delete commit hash mapping
	hashKey := fmt.Sprintf("version/hash/%s", info.CommitHash)
	if err := vm.backend.Delete(ctx, hashKey); err != nil {
		return fmt.Errorf("failed to delete commit hash mapping: %w", err)
	}

	// Update current version if this was the latest
	currentVersion, err := vm.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if version == currentVersion {
		// Find the previous version
		newCurrent := version - 1
		if newCurrent < 1 {
			// No more versions, reset to 0
			newCurrent = 0
		}

		currentData := []byte(strconv.FormatInt(newCurrent, 10))
		if err := vm.backend.Put(ctx, "version/current", currentData); err != nil {
			return fmt.Errorf("failed to update current version: %w", err)
		}
	}

	return nil
}
