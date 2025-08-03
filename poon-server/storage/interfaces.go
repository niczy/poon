package storage

import (
	"context"
	"io"
)

// ObjectStore defines the interface for storing and retrieving objects
type ObjectStore interface {
	// Store stores an object and returns its hash
	Store(ctx context.Context, obj *Object) (Hash, error)

	// Get retrieves an object by its hash
	Get(ctx context.Context, hash Hash) (*Object, error)

	// Exists checks if an object exists
	Exists(ctx context.Context, hash Hash) (bool, error)

	// Delete removes an object (optional for some backends)
	Delete(ctx context.Context, hash Hash) error

	// List returns all object hashes (optional, for debugging)
	List(ctx context.Context) ([]Hash, error)
}

// VersionStore defines the interface for version management
type VersionStore interface {
	// GetCurrentVersion returns the current version number
	GetCurrentVersion(ctx context.Context) (int64, error)

	// GetVersionInfo returns version information for a specific version
	GetVersionInfo(ctx context.Context, version int64) (*VersionInfo, error)

	// GetLatestVersionInfo returns the latest version information
	GetLatestVersionInfo(ctx context.Context) (*VersionInfo, error)

	// CreateVersion creates a new version pointing to a commit
	CreateVersion(ctx context.Context, commitHash Hash, message string) (*VersionInfo, error)

	// ListVersions returns all versions in chronological order
	ListVersions(ctx context.Context, limit int) ([]*VersionInfo, error)
}

// ContentAddressable defines the interface for content-addressable operations
type ContentAddressable interface {
	// ComputeHash computes the hash for given content
	ComputeHash(content []byte) Hash

	// StoreBlob stores file content and returns its hash
	StoreBlob(ctx context.Context, content []byte) (Hash, error)

	// StoreTree stores directory structure and returns its hash
	StoreTree(ctx context.Context, tree *TreeObject) (Hash, error)

	// StoreCommit stores commit object and returns its hash
	StoreCommit(ctx context.Context, commit *CommitObject) (Hash, error)

	// GetBlob retrieves blob content
	GetBlob(ctx context.Context, hash Hash) (*BlobObject, error)

	// GetTree retrieves tree structure
	GetTree(ctx context.Context, hash Hash) (*TreeObject, error)

	// GetCommit retrieves commit object
	GetCommit(ctx context.Context, hash Hash) (*CommitObject, error)
}

// Repository combines all storage interfaces for high-level operations
type Repository interface {
	ObjectStore
	VersionStore
	ContentAddressable

	// ReadFile reads file content at a specific path in a version
	ReadFile(ctx context.Context, version int64, path string) ([]byte, error)

	// ReadDirectory lists directory contents at a specific path in a version
	ReadDirectory(ctx context.Context, version int64, path string) ([]*TreeEntry, error)

	// CreateCommitFromFileSystem creates a commit from current file system state
	CreateCommitFromFileSystem(ctx context.Context, rootPath string, author, message string) (*VersionInfo, error)

	// ApplyPatch applies a patch and creates a new version
	ApplyPatch(ctx context.Context, patch []byte, author, message string) (*VersionInfo, error)

	// Close closes the repository and any underlying resources
	Close() error
}

// StorageBackend defines the low-level storage interface that can be implemented
// by different backends (in-memory, S3, filesystem, etc.)
type StorageBackend interface {
	// Put stores data at the given key
	Put(ctx context.Context, key string, data []byte) error

	// Get retrieves data for the given key
	Get(ctx context.Context, key string) ([]byte, error)

	// Exists checks if a key exists
	Exists(ctx context.Context, key string) (bool, error)

	// Delete removes data for the given key
	Delete(ctx context.Context, key string) error

	// List returns all keys with the given prefix
	List(ctx context.Context, prefix string) ([]string, error)

	// Stream returns a reader for large objects (optional)
	Stream(ctx context.Context, key string) (io.ReadCloser, error)

	// Close closes the backend and releases resources
	Close() error
}
