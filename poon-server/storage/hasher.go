package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Hasher provides content-addressable hashing functionality
type Hasher struct{}

// NewHasher creates a new hasher instance
func NewHasher() *Hasher {
	return &Hasher{}
}

// ComputeHash computes SHA-256 hash for raw content
func (h *Hasher) ComputeHash(content []byte) Hash {
	hash := sha256.Sum256(content)
	return Hash(hex.EncodeToString(hash[:]))
}

// ComputeObjectHash computes hash for a typed object
func (h *Hasher) ComputeObjectHash(objType ObjectType, content []byte) Hash {
	// Create a canonical representation: type + size + content
	header := fmt.Sprintf("%s %d\x00", objType, len(content))
	fullContent := append([]byte(header), content...)
	return h.ComputeHash(fullContent)
}

// ComputeBlobHash computes hash for blob content
func (h *Hasher) ComputeBlobHash(content []byte) Hash {
	return h.ComputeObjectHash(ObjectTypeBlob, content)
}

// ComputeTreeHash computes hash for tree object
func (h *Hasher) ComputeTreeHash(tree *TreeObject) (Hash, error) {
	// Serialize tree to canonical JSON format
	data, err := json.Marshal(tree)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tree: %w", err)
	}
	return h.ComputeObjectHash(ObjectTypeTree, data), nil
}

// ComputeCommitHash computes hash for commit object
func (h *Hasher) ComputeCommitHash(commit *CommitObject) (Hash, error) {
	// Serialize commit to canonical JSON format
	data, err := json.Marshal(commit)
	if err != nil {
		return "", fmt.Errorf("failed to marshal commit: %w", err)
	}
	return h.ComputeObjectHash(ObjectTypeCommit, data), nil
}

// ValidateHash checks if a hash string is valid SHA-256
func (h *Hasher) ValidateHash(hash Hash) error {
	if len(hash) != 64 {
		return fmt.Errorf("invalid hash length: expected 64 characters, got %d", len(hash))
	}
	
	_, err := hex.DecodeString(string(hash))
	if err != nil {
		return fmt.Errorf("invalid hash format: %w", err)
	}
	
	return nil
}

// VerifyObject verifies that an object's content matches its hash
func (h *Hasher) VerifyObject(obj *Object) error {
	if err := h.ValidateHash(obj.Hash); err != nil {
		return fmt.Errorf("invalid object hash: %w", err)
	}
	
	expectedHash := h.ComputeObjectHash(obj.Type, obj.Content)
	if expectedHash != obj.Hash {
		return fmt.Errorf("object hash mismatch: expected %s, got %s", expectedHash, obj.Hash)
	}
	
	return nil
}

// CreateObject creates an object with computed hash
func (h *Hasher) CreateObject(objType ObjectType, content []byte) *Object {
	hash := h.ComputeObjectHash(objType, content)
	return &Object{
		Hash:    hash,
		Type:    objType,
		Size:    int64(len(content)),
		Content: content,
	}
}

// CreateBlobObject creates a blob object from content
func (h *Hasher) CreateBlobObject(content []byte) *Object {
	return h.CreateObject(ObjectTypeBlob, content)
}

// CreateTreeObject creates a tree object from tree structure
func (h *Hasher) CreateTreeObject(tree *TreeObject) (*Object, error) {
	data, err := json.Marshal(tree)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tree: %w", err)
	}
	return h.CreateObject(ObjectTypeTree, data), nil
}

// CreateCommitObject creates a commit object from commit structure
func (h *Hasher) CreateCommitObject(commit *CommitObject) (*Object, error) {
	data, err := json.Marshal(commit)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal commit: %w", err)
	}
	return h.CreateObject(ObjectTypeCommit, data), nil
}