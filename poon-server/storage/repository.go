package storage

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nic/poon/poon-server/merge"
)

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	*ContentStore
	*VersionManager
	hasher *Hasher
}

// NewRepository creates a new repository with the given backend
func NewRepository(backend StorageBackend) Repository {
	contentStore := NewContentStore(backend)
	versionManager := NewVersionManager(backend)

	return &RepositoryImpl{
		ContentStore:   contentStore,
		VersionManager: versionManager,
		hasher:         NewHasher(),
	}
}

// ReadFile reads file content at a specific path in a version
func (r *RepositoryImpl) ReadFile(ctx context.Context, version int64, path string) ([]byte, error) {
	// Get version info
	versionInfo, err := r.GetVersionInfo(ctx, version)
	if err != nil {
		return nil, fmt.Errorf("version %d not found: %w", version, err)
	}

	// Get commit object
	commit, err := r.GetCommit(ctx, versionInfo.CommitHash)
	if err != nil {
		return nil, fmt.Errorf("commit not found: %w", err)
	}

	// Navigate to file through tree structure
	blobHash, err := r.findFileInTree(ctx, commit.RootTree, path)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// Get blob content
	blob, err := r.GetBlob(ctx, blobHash)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob: %w", err)
	}

	return blob.Content, nil
}

// ReadDirectory lists directory contents at a specific path in a version
func (r *RepositoryImpl) ReadDirectory(ctx context.Context, version int64, path string) ([]*TreeEntry, error) {
	// Get version info
	versionInfo, err := r.GetVersionInfo(ctx, version)
	if err != nil {
		return nil, fmt.Errorf("version %d not found: %w", version, err)
	}

	// Get commit object
	commit, err := r.GetCommit(ctx, versionInfo.CommitHash)
	if err != nil {
		return nil, fmt.Errorf("commit not found: %w", err)
	}

	// Navigate to directory through tree structure
	treeHash, err := r.findDirectoryInTree(ctx, commit.RootTree, path)
	if err != nil {
		return nil, fmt.Errorf("directory not found: %w", err)
	}

	// Get tree object
	tree, err := r.GetTree(ctx, treeHash)
	if err != nil {
		return nil, fmt.Errorf("failed to read tree: %w", err)
	}

	// Convert []TreeEntry to []*TreeEntry
	result := make([]*TreeEntry, len(tree.Entries))
	for i := range tree.Entries {
		result[i] = &tree.Entries[i]
	}
	return result, nil
}

// CreateCommitFromFileSystem creates a commit from current file system state
func (r *RepositoryImpl) CreateCommitFromFileSystem(ctx context.Context, rootPath string, author, message string) (*VersionInfo, error) {
	// Get current version for parent reference
	currentVersion, err := r.GetCurrentVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	var parentHash *Hash
	if currentVersion > 0 {
		parentInfo, err := r.GetVersionInfo(ctx, currentVersion)
		if err == nil {
			parentHash = &parentInfo.CommitHash
		}
	}

	// Create tree from file system
	rootTreeHash, err := r.createTreeFromFileSystem(ctx, rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create tree from filesystem: %w", err)
	}

	// Create commit object
	commit := &CommitObject{
		RootTree:  rootTreeHash,
		Parent:    parentHash,
		Author:    author,
		Message:   message,
		Timestamp: time.Now(),
		Version:   currentVersion + 1,
	}

	// Store commit
	commitHash, err := r.StoreCommit(ctx, commit)
	if err != nil {
		return nil, fmt.Errorf("failed to store commit: %w", err)
	}

	// Create new version
	return r.CreateVersion(ctx, commitHash, message)
}

// ApplyPatch applies a patch and creates a new version
func (r *RepositoryImpl) ApplyPatch(ctx context.Context, patchData []byte, author, message string) (*VersionInfo, error) {
	// Parse patch
	parsed, err := merge.ParsePatch(patchData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse patch: %w", err)
	}

	// Get current version
	currentVersion, err := r.GetCurrentVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion == 0 {
		return nil, fmt.Errorf("cannot apply patch to empty repository")
	}

	// Get current commit
	currentInfo, err := r.GetVersionInfo(ctx, currentVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get current version info: %w", err)
	}

	currentCommit, err := r.GetCommit(ctx, currentInfo.CommitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit: %w", err)
	}

	// Apply patch to tree structure
	newRootHash, err := r.applyPatchToTree(ctx, currentCommit.RootTree, parsed)
	if err != nil {
		return nil, fmt.Errorf("failed to apply patch: %w", err)
	}

	// Create new commit
	newCommit := &CommitObject{
		RootTree:  newRootHash,
		Parent:    &currentInfo.CommitHash,
		Author:    author,
		Message:   message,
		Timestamp: time.Now(),
		Version:   currentVersion + 1,
	}

	// Store new commit
	commitHash, err := r.StoreCommit(ctx, newCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to store commit: %w", err)
	}

	// Create new version
	return r.CreateVersion(ctx, commitHash, message)
}

// Close closes the repository and any underlying resources
func (r *RepositoryImpl) Close() error {
	return r.ContentStore.backend.Close()
}

// Helper methods

func (r *RepositoryImpl) findFileInTree(ctx context.Context, treeHash Hash, path string) (Hash, error) {
	if path == "" {
		return "", fmt.Errorf("empty path")
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	currentTreeHash := treeHash

	// Navigate through directories
	for i, part := range parts[:len(parts)-1] {
		tree, err := r.GetTree(ctx, currentTreeHash)
		if err != nil {
			return "", fmt.Errorf("failed to get tree at level %d: %w", i, err)
		}

		found := false
		for _, entry := range tree.Entries {
			if entry.Name == part && entry.Type == ObjectTypeTree {
				currentTreeHash = entry.Hash
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("directory '%s' not found", part)
		}
	}

	// Find file in final directory
	tree, err := r.GetTree(ctx, currentTreeHash)
	if err != nil {
		return "", fmt.Errorf("failed to get final tree: %w", err)
	}

	fileName := parts[len(parts)-1]
	for _, entry := range tree.Entries {
		if entry.Name == fileName && entry.Type == ObjectTypeBlob {
			return entry.Hash, nil
		}
	}

	return "", fmt.Errorf("file '%s' not found", fileName)
}

func (r *RepositoryImpl) findDirectoryInTree(ctx context.Context, treeHash Hash, path string) (Hash, error) {
	if path == "" {
		return treeHash, nil // Root directory
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	currentTreeHash := treeHash

	// Navigate through all directories
	for i, part := range parts {
		tree, err := r.GetTree(ctx, currentTreeHash)
		if err != nil {
			return "", fmt.Errorf("failed to get tree at level %d: %w", i, err)
		}

		found := false
		for _, entry := range tree.Entries {
			if entry.Name == part && entry.Type == ObjectTypeTree {
				currentTreeHash = entry.Hash
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("directory '%s' not found", part)
		}
	}

	return currentTreeHash, nil
}

func (r *RepositoryImpl) createTreeFromFileSystem(ctx context.Context, dirPath string) (Hash, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var treeEntries []TreeEntry

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			// Recursively create tree for subdirectory
			subTreeHash, err := r.createTreeFromFileSystem(ctx, fullPath)
			if err != nil {
				return "", fmt.Errorf("failed to create subtree for %s: %w", entry.Name(), err)
			}

			treeEntries = append(treeEntries, TreeEntry{
				Name: entry.Name(),
				Hash: subTreeHash,
				Type: ObjectTypeTree,
				Mode: int32(entry.Type() & fs.ModePerm),
			})
		} else {
			// Read file content and create blob
			content, err := os.ReadFile(fullPath)
			if err != nil {
				return "", fmt.Errorf("failed to read file %s: %w", entry.Name(), err)
			}

			blobHash, err := r.StoreBlob(ctx, content)
			if err != nil {
				return "", fmt.Errorf("failed to store blob for %s: %w", entry.Name(), err)
			}

			info, err := entry.Info()
			if err != nil {
				return "", fmt.Errorf("failed to get file info for %s: %w", entry.Name(), err)
			}

			treeEntries = append(treeEntries, TreeEntry{
				Name: entry.Name(),
				Hash: blobHash,
				Type: ObjectTypeBlob,
				Mode: int32(info.Mode()),
				Size: info.Size(),
			})
		}
	}

	// Create and store tree object
	tree := &TreeObject{Entries: treeEntries}
	return r.StoreTree(ctx, tree)
}

func (r *RepositoryImpl) applyPatchToTree(ctx context.Context, rootTreeHash Hash, patch *merge.ParsedPatch) (Hash, error) {
	// For now, this is a simplified implementation that creates a new version from file system
	// In a full implementation, this would apply the patch directly to the tree structure

	// TODO: Implement proper patch application to tree structure
	// This requires:
	// 1. Navigate to the target file in the tree
	// 2. Apply the patch to the file content using merge.ApplyPatch
	// 3. Create new blob with patched content
	// 4. Update tree structures with new blob hash
	// 5. Return new root tree hash

	return "", fmt.Errorf("patch application to tree structure not yet implemented - please use CreateCommitFromFileSystem for now")
}
