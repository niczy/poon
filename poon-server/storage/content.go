package storage

import (
	"context"
	"encoding/json"
	"fmt"
)

// ContentStore implements ContentAddressable interface
type ContentStore struct {
	backend StorageBackend
	hasher  *Hasher
}

// NewContentStore creates a new content-addressable store
func NewContentStore(backend StorageBackend) *ContentStore {
	return &ContentStore{
		backend: backend,
		hasher:  NewHasher(),
	}
}

// ComputeHash computes the hash for given content
func (cs *ContentStore) ComputeHash(content []byte) Hash {
	return cs.hasher.ComputeHash(content)
}

// Store stores an object and returns its hash
func (cs *ContentStore) Store(ctx context.Context, obj *Object) (Hash, error) {
	// Verify object integrity
	if err := cs.hasher.VerifyObject(obj); err != nil {
		return "", fmt.Errorf("object verification failed: %w", err)
	}
	
	// Serialize object for storage
	data, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal object: %w", err)
	}
	
	// Store with hash as key
	key := "objects/" + string(obj.Hash)
	if err := cs.backend.Put(ctx, key, data); err != nil {
		return "", fmt.Errorf("failed to store object: %w", err)
	}
	
	return obj.Hash, nil
}

// Get retrieves an object by its hash
func (cs *ContentStore) Get(ctx context.Context, hash Hash) (*Object, error) {
	if err := cs.hasher.ValidateHash(hash); err != nil {
		return nil, fmt.Errorf("invalid hash: %w", err)
	}
	
	key := "objects/" + string(hash)
	data, err := cs.backend.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("object not found: %w", err)
	}
	
	var obj Object
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal object: %w", err)
	}
	
	// Verify object integrity
	if err := cs.hasher.VerifyObject(&obj); err != nil {
		return nil, fmt.Errorf("stored object verification failed: %w", err)
	}
	
	return &obj, nil
}

// Exists checks if an object exists
func (cs *ContentStore) Exists(ctx context.Context, hash Hash) (bool, error) {
	if err := cs.hasher.ValidateHash(hash); err != nil {
		return false, fmt.Errorf("invalid hash: %w", err)
	}
	
	key := "objects/" + string(hash)
	return cs.backend.Exists(ctx, key)
}

// Delete removes an object
func (cs *ContentStore) Delete(ctx context.Context, hash Hash) error {
	if err := cs.hasher.ValidateHash(hash); err != nil {
		return fmt.Errorf("invalid hash: %w", err)
	}
	
	key := "objects/" + string(hash)
	return cs.backend.Delete(ctx, key)
}

// List returns all object hashes
func (cs *ContentStore) List(ctx context.Context) ([]Hash, error) {
	keys, err := cs.backend.List(ctx, "objects/")
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}
	
	hashes := make([]Hash, 0, len(keys))
	for _, key := range keys {
		// Remove "objects/" prefix to get hash
		if len(key) > 8 {
			hash := key[8:]
			hashes = append(hashes, Hash(hash))
		}
	}
	
	return hashes, nil
}

// StoreBlob stores file content and returns its hash
func (cs *ContentStore) StoreBlob(ctx context.Context, content []byte) (Hash, error) {
	obj := cs.hasher.CreateBlobObject(content)
	return cs.Store(ctx, obj)
}

// StoreTree stores directory structure and returns its hash
func (cs *ContentStore) StoreTree(ctx context.Context, tree *TreeObject) (Hash, error) {
	obj, err := cs.hasher.CreateTreeObject(tree)
	if err != nil {
		return "", fmt.Errorf("failed to create tree object: %w", err)
	}
	return cs.Store(ctx, obj)
}

// StoreCommit stores commit object and returns its hash
func (cs *ContentStore) StoreCommit(ctx context.Context, commit *CommitObject) (Hash, error) {
	obj, err := cs.hasher.CreateCommitObject(commit)
	if err != nil {
		return "", fmt.Errorf("failed to create commit object: %w", err)
	}
	return cs.Store(ctx, obj)
}

// GetBlob retrieves blob content
func (cs *ContentStore) GetBlob(ctx context.Context, hash Hash) (*BlobObject, error) {
	obj, err := cs.Get(ctx, hash)
	if err != nil {
		return nil, err
	}
	
	if obj.Type != ObjectTypeBlob {
		return nil, fmt.Errorf("object is not a blob: %s", obj.Type)
	}
	
	return &BlobObject{Content: obj.Content}, nil
}

// GetTree retrieves tree structure
func (cs *ContentStore) GetTree(ctx context.Context, hash Hash) (*TreeObject, error) {
	obj, err := cs.Get(ctx, hash)
	if err != nil {
		return nil, err
	}
	
	if obj.Type != ObjectTypeTree {
		return nil, fmt.Errorf("object is not a tree: %s", obj.Type)
	}
	
	var tree TreeObject
	if err := json.Unmarshal(obj.Content, &tree); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tree: %w", err)
	}
	
	return &tree, nil
}

// GetCommit retrieves commit object
func (cs *ContentStore) GetCommit(ctx context.Context, hash Hash) (*CommitObject, error) {
	obj, err := cs.Get(ctx, hash)
	if err != nil {
		return nil, err
	}
	
	if obj.Type != ObjectTypeCommit {
		return nil, fmt.Errorf("object is not a commit: %s", obj.Type)
	}
	
	var commit CommitObject
	if err := json.Unmarshal(obj.Content, &commit); err != nil {
		return nil, fmt.Errorf("failed to unmarshal commit: %w", err)
	}
	
	return &commit, nil
}