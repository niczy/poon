package storage

import (
	"time"
)

// Hash represents a content-addressable hash (SHA-256)
type Hash string

// ObjectType represents the type of object stored
type ObjectType string

const (
	ObjectTypeBlob   ObjectType = "blob"
	ObjectTypeTree   ObjectType = "tree"
	ObjectTypeCommit ObjectType = "commit"
)

// Object represents a stored object with its metadata
type Object struct {
	Hash    Hash       `json:"hash"`
	Type    ObjectType `json:"type"`
	Size    int64      `json:"size"`
	Content []byte     `json:"content"`
}

// BlobObject represents file content
type BlobObject struct {
	Content []byte `json:"content"`
}

// TreeEntry represents an entry in a tree (file or directory)
type TreeEntry struct {
	Name string     `json:"name"`
	Hash Hash       `json:"hash"`
	Type ObjectType `json:"type"`
	Mode int32      `json:"mode"` // File permissions
	Size int64      `json:"size,omitempty"`
}

// TreeObject represents directory structure
type TreeObject struct {
	Entries []TreeEntry `json:"entries"`
}

// CommitObject represents a version snapshot
type CommitObject struct {
	RootTree   Hash      `json:"root_tree"`
	Parent     *Hash     `json:"parent,omitempty"`
	Author     string    `json:"author"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
	Version    int64     `json:"version"`
}

// VersionInfo maps version numbers to commit hashes
type VersionInfo struct {
	Version    int64     `json:"version"`
	CommitHash Hash      `json:"commit_hash"`
	Timestamp  time.Time `json:"timestamp"`
	Message    string    `json:"message"`
}