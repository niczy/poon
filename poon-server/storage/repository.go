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

	// Clean the path to handle "." and ".." properly
	cleanPath := filepath.Clean(path)
	if cleanPath == "." {
		return "", fmt.Errorf("cannot read directory as file")
	}

	parts := strings.Split(strings.Trim(cleanPath, "/"), "/")
	// Filter out empty parts and "." parts
	var filteredParts []string
	for _, part := range parts {
		if part != "" && part != "." {
			filteredParts = append(filteredParts, part)
		}
	}
	parts = filteredParts

	if len(parts) == 0 {
		return "", fmt.Errorf("invalid path")
	}

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

	// Clean the path to handle "." and ".." properly
	cleanPath := filepath.Clean(path)
	if cleanPath == "." {
		return treeHash, nil // Current directory is root
	}

	parts := strings.Split(strings.Trim(cleanPath, "/"), "/")
	// Filter out empty parts and "." parts
	var filteredParts []string
	for _, part := range parts {
		if part != "" && part != "." {
			filteredParts = append(filteredParts, part)
		}
	}
	parts = filteredParts

	if len(parts) == 0 {
		return treeHash, nil // Resolved to root directory
	}

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

			info, err := entry.Info()
			if err != nil {
				return "", fmt.Errorf("failed to get directory info for %s: %w", entry.Name(), err)
			}

			treeEntries = append(treeEntries, TreeEntry{
				Name:    entry.Name(),
				Hash:    subTreeHash,
				Type:    ObjectTypeTree,
				Mode:    int32(entry.Type() & fs.ModePerm),
				ModTime: info.ModTime().Unix(),
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
				Name:    entry.Name(),
				Hash:    blobHash,
				Type:    ObjectTypeBlob,
				Mode:    int32(info.Mode()),
				Size:    info.Size(),
				ModTime: info.ModTime().Unix(),
			})
		}
	}

	// Create and store tree object
	tree := &TreeObject{Entries: treeEntries}
	return r.StoreTree(ctx, tree)
}

func (r *RepositoryImpl) applyPatchToTree(ctx context.Context, rootTreeHash Hash, patch *merge.ParsedPatch) (Hash, error) {
	// Get the target file path from the patch
	targetPath := patch.Header.NewFile
	if targetPath == "" {
		targetPath = patch.Header.OldFile
	}

	if targetPath == "" {
		return "", fmt.Errorf("patch does not specify a target file")
	}

	// Validate the target path for security (same validation as in validatePath)
	if strings.Contains(targetPath, "..") {
		return "", fmt.Errorf("path traversal not allowed in patch target: path contains '..'")
	}

	cleanPath := filepath.Clean(targetPath)
	if strings.HasPrefix(cleanPath, "..") || strings.HasPrefix(cleanPath, "/") {
		return "", fmt.Errorf("invalid patch target path: path must be relative and within repository")
	}

	// Get the current file content
	var originalContent []byte
	var err error

	// Try to read the existing file content
	originalContent, err = r.readFileFromTree(ctx, rootTreeHash, targetPath)
	if err != nil {
		// File might not exist (new file), start with empty content
		originalContent = []byte{}
	}

	// Apply the patch to the content
	patchedContent, err := r.applyPatchToContent(originalContent, patch)
	if err != nil {
		return "", fmt.Errorf("failed to apply patch to content: %w", err)
	}

	// Store the new blob
	newBlobHash, err := r.StoreBlob(ctx, patchedContent)
	if err != nil {
		return "", fmt.Errorf("failed to store patched blob: %w", err)
	}

	// Update the tree structure with the new blob
	newRootTreeHash, err := r.updateTreeWithBlob(ctx, rootTreeHash, targetPath, newBlobHash, int64(len(patchedContent)))
	if err != nil {
		return "", fmt.Errorf("failed to update tree structure: %w", err)
	}

	return newRootTreeHash, nil
}

// Helper function to read file content from tree structure
func (r *RepositoryImpl) readFileFromTree(ctx context.Context, rootTreeHash Hash, path string) ([]byte, error) {
	blobHash, err := r.findFileInTree(ctx, rootTreeHash, path)
	if err != nil {
		return nil, err
	}

	blob, err := r.GetBlob(ctx, blobHash)
	if err != nil {
		return nil, err
	}

	return blob.Content, nil
}

// Helper function to apply patch to content without filesystem
func (r *RepositoryImpl) applyPatchToContent(originalContent []byte, patch *merge.ParsedPatch) ([]byte, error) {
	var originalLines []string

	if len(originalContent) > 0 {
		originalContentStr := string(originalContent)
		if originalContentStr != "" {
			originalLines = strings.Split(originalContentStr, "\n")
			// Remove empty last line if present
			if len(originalLines) > 0 && originalLines[len(originalLines)-1] == "" {
				originalLines = originalLines[:len(originalLines)-1]
			}
		}
	}

	result := make([]string, 0, len(originalLines)+100)
	originalIndex := 0

	for _, hunk := range patch.Hunks {
		// Copy context lines before hunk
		for originalIndex < hunk.OldStart-1 && originalIndex < len(originalLines) {
			result = append(result, originalLines[originalIndex])
			originalIndex++
		}

		// Apply hunk changes
		for _, patchLine := range hunk.Lines {
			switch patchLine.Type {
			case " ": // Context line
				if originalIndex < len(originalLines) {
					result = append(result, originalLines[originalIndex])
					originalIndex++
				}
			case "-": // Deletion
				if originalIndex < len(originalLines) {
					originalIndex++
				}
			case "+": // Addition
				result = append(result, patchLine.Content)
			}
		}
	}

	// Copy remaining lines
	for originalIndex < len(originalLines) {
		result = append(result, originalLines[originalIndex])
		originalIndex++
	}

	newContent := strings.Join(result, "\n")
	if len(result) > 0 {
		newContent += "\n"
	}

	return []byte(newContent), nil
}

// Helper function to update tree structure with new blob
func (r *RepositoryImpl) updateTreeWithBlob(ctx context.Context, rootTreeHash Hash, path string, blobHash Hash, size int64) (Hash, error) {
	if path == "" {
		return "", fmt.Errorf("empty path")
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")

	// If it's a single file in root directory
	if len(parts) == 1 {
		return r.updateTreeEntryWithBlob(ctx, rootTreeHash, parts[0], blobHash, size)
	}

	// Navigate through directory structure and update trees recursively
	return r.updateNestedTreeWithBlob(ctx, rootTreeHash, parts, blobHash, size)
}

// Helper function to update a single tree entry with new blob
func (r *RepositoryImpl) updateTreeEntryWithBlob(ctx context.Context, treeHash Hash, fileName string, blobHash Hash, size int64) (Hash, error) {
	tree, err := r.GetTree(ctx, treeHash)
	if err != nil {
		return "", fmt.Errorf("failed to get tree: %w", err)
	}

	// Create new tree entries, updating or adding the target file
	var newEntries []TreeEntry
	fileUpdated := false

	for _, entry := range tree.Entries {
		if entry.Name == fileName && entry.Type == ObjectTypeBlob {
			// Update existing file
			newEntries = append(newEntries, TreeEntry{
				Name:    fileName,
				Hash:    blobHash,
				Type:    ObjectTypeBlob,
				Mode:    entry.Mode, // Keep original mode
				Size:    size,
				ModTime: time.Now().Unix(), // Current time for updated file
			})
			fileUpdated = true
		} else {
			// Keep other entries unchanged
			newEntries = append(newEntries, entry)
		}
	}

	// If file wasn't found, add it as a new entry
	if !fileUpdated {
		newEntries = append(newEntries, TreeEntry{
			Name:    fileName,
			Hash:    blobHash,
			Type:    ObjectTypeBlob,
			Mode:    0644, // Default file mode
			Size:    size,
			ModTime: time.Now().Unix(), // Current time for new file
		})
	}

	// Create and store new tree
	newTree := &TreeObject{Entries: newEntries}
	return r.StoreTree(ctx, newTree)
}

// Helper function to update nested tree structure
func (r *RepositoryImpl) updateNestedTreeWithBlob(ctx context.Context, rootTreeHash Hash, pathParts []string, blobHash Hash, size int64) (Hash, error) {
	if len(pathParts) == 0 {
		return "", fmt.Errorf("empty path parts")
	}

	tree, err := r.GetTree(ctx, rootTreeHash)
	if err != nil {
		return "", fmt.Errorf("failed to get tree: %w", err)
	}

	var newEntries []TreeEntry
	dirName := pathParts[0]
	dirUpdated := false

	for _, entry := range tree.Entries {
		if entry.Name == dirName && entry.Type == ObjectTypeTree {
			// This is the directory we need to update
			var newSubTreeHash Hash
			if len(pathParts) == 2 {
				// We're updating a file in this directory
				newSubTreeHash, err = r.updateTreeEntryWithBlob(ctx, entry.Hash, pathParts[1], blobHash, size)
			} else {
				// We need to go deeper
				newSubTreeHash, err = r.updateNestedTreeWithBlob(ctx, entry.Hash, pathParts[1:], blobHash, size)
			}

			if err != nil {
				return "", fmt.Errorf("failed to update subtree: %w", err)
			}

			newEntries = append(newEntries, TreeEntry{
				Name:    dirName,
				Hash:    newSubTreeHash,
				Type:    ObjectTypeTree,
				Mode:    entry.Mode,
				Size:    entry.Size,
				ModTime: entry.ModTime, // Keep original ModTime for existing directories
			})
			dirUpdated = true
		} else {
			// Keep other entries unchanged
			newEntries = append(newEntries, entry)
		}
	}

	// If directory wasn't found, create it
	if !dirUpdated {
		var newSubTreeHash Hash
		if len(pathParts) == 2 {
			// Create new directory with the file
			newTree := &TreeObject{
				Entries: []TreeEntry{
					{
						Name:    pathParts[1],
						Hash:    blobHash,
						Type:    ObjectTypeBlob,
						Mode:    0644,
						Size:    size,
						ModTime: time.Now().Unix(),
					},
				},
			}
			newSubTreeHash, err = r.StoreTree(ctx, newTree)
		} else {
			// Create nested directory structure
			newSubTreeHash, err = r.createNestedTreeWithBlob(ctx, pathParts[1:], blobHash, size)
		}

		if err != nil {
			return "", fmt.Errorf("failed to create new subtree: %w", err)
		}

		newEntries = append(newEntries, TreeEntry{
			Name:    dirName,
			Hash:    newSubTreeHash,
			Type:    ObjectTypeTree,
			Mode:    0755,              // Default directory mode
			ModTime: time.Now().Unix(), // Current time for new directory
		})
	}

	// Create and store new tree
	newTree := &TreeObject{Entries: newEntries}
	return r.StoreTree(ctx, newTree)
}

// Helper function to create nested directory structure with blob
func (r *RepositoryImpl) createNestedTreeWithBlob(ctx context.Context, pathParts []string, blobHash Hash, size int64) (Hash, error) {
	if len(pathParts) == 0 {
		return "", fmt.Errorf("empty path parts")
	}

	if len(pathParts) == 1 {
		// Create tree with single file
		tree := &TreeObject{
			Entries: []TreeEntry{
				{
					Name:    pathParts[0],
					Hash:    blobHash,
					Type:    ObjectTypeBlob,
					Mode:    0644,
					Size:    size,
					ModTime: time.Now().Unix(),
				},
			},
		}
		return r.StoreTree(ctx, tree)
	}

	// Create nested structure recursively
	subTreeHash, err := r.createNestedTreeWithBlob(ctx, pathParts[1:], blobHash, size)
	if err != nil {
		return "", err
	}

	tree := &TreeObject{
		Entries: []TreeEntry{
			{
				Name:    pathParts[0],
				Hash:    subTreeHash,
				Type:    ObjectTypeTree,
				Mode:    0755,
				ModTime: time.Now().Unix(),
			},
		},
	}
	return r.StoreTree(ctx, tree)
}
