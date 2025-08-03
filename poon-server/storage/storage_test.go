package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryBackend(t *testing.T) {
	backend := NewMemoryBackend()
	defer backend.Close()
	
	ctx := context.Background()
	
	t.Run("Put and Get", func(t *testing.T) {
		key := "test-key"
		data := []byte("test data")
		
		err := backend.Put(ctx, key, data)
		require.NoError(t, err)
		
		retrieved, err := backend.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, data, retrieved)
	})
	
	t.Run("Exists", func(t *testing.T) {
		key := "exists-key"
		data := []byte("exists data")
		
		exists, err := backend.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists)
		
		err = backend.Put(ctx, key, data)
		require.NoError(t, err)
		
		exists, err = backend.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)
	})
	
	t.Run("Delete", func(t *testing.T) {
		key := "delete-key"
		data := []byte("delete data")
		
		err := backend.Put(ctx, key, data)
		require.NoError(t, err)
		
		err = backend.Delete(ctx, key)
		require.NoError(t, err)
		
		exists, err := backend.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists)
	})
	
	t.Run("List", func(t *testing.T) {
		prefix := "list-test/"
		keys := []string{
			prefix + "key1",
			prefix + "key2",
			prefix + "key3",
			"other-key",
		}
		
		for _, key := range keys {
			err := backend.Put(ctx, key, []byte("data"))
			require.NoError(t, err)
		}
		
		listed, err := backend.List(ctx, prefix)
		require.NoError(t, err)
		assert.Len(t, listed, 3)
		
		for _, key := range listed {
			assert.Contains(t, keys[:3], key)
		}
	})
}

func TestHasher(t *testing.T) {
	hasher := NewHasher()
	
	t.Run("ComputeHash", func(t *testing.T) {
		content := []byte("hello world")
		hash1 := hasher.ComputeHash(content)
		hash2 := hasher.ComputeHash(content)
		
		assert.Equal(t, hash1, hash2)
		assert.Len(t, string(hash1), 64) // SHA-256 hex length
		
		// Different content should produce different hash
		differentContent := []byte("hello world!")
		hash3 := hasher.ComputeHash(differentContent)
		assert.NotEqual(t, hash1, hash3)
	})
	
	t.Run("ValidateHash", func(t *testing.T) {
		validHash := Hash("b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9")
		err := hasher.ValidateHash(validHash)
		assert.NoError(t, err)
		
		invalidHash := Hash("not-a-valid-hash")
		err = hasher.ValidateHash(invalidHash)
		assert.Error(t, err)
		
		shortHash := Hash("abc123")
		err = hasher.ValidateHash(shortHash)
		assert.Error(t, err)
	})
	
	t.Run("CreateBlobObject", func(t *testing.T) {
		content := []byte("blob content")
		obj := hasher.CreateBlobObject(content)
		
		assert.Equal(t, ObjectTypeBlob, obj.Type)
		assert.Equal(t, content, obj.Content)
		assert.Equal(t, int64(len(content)), obj.Size)
		assert.NotEmpty(t, obj.Hash)
		
		err := hasher.VerifyObject(obj)
		assert.NoError(t, err)
	})
}

func TestContentStore(t *testing.T) {
	backend := NewMemoryBackend()
	defer backend.Close()
	
	store := NewContentStore(backend)
	ctx := context.Background()
	
	t.Run("StoreBlob", func(t *testing.T) {
		content := []byte("test blob content")
		hash, err := store.StoreBlob(ctx, content)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		
		blob, err := store.GetBlob(ctx, hash)
		require.NoError(t, err)
		assert.Equal(t, content, blob.Content)
	})
	
	t.Run("StoreTree", func(t *testing.T) {
		tree := &TreeObject{
			Entries: []TreeEntry{
				{
					Name: "file1.txt",
					Hash: Hash("abc123"),
					Type: ObjectTypeBlob,
					Mode: 0644,
					Size: 100,
				},
				{
					Name: "subdir",
					Hash: Hash("def456"),
					Type: ObjectTypeTree,
					Mode: 0755,
				},
			},
		}
		
		hash, err := store.StoreTree(ctx, tree)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		
		retrieved, err := store.GetTree(ctx, hash)
		require.NoError(t, err)
		assert.Len(t, retrieved.Entries, 2)
		assert.Equal(t, "file1.txt", retrieved.Entries[0].Name)
		assert.Equal(t, "subdir", retrieved.Entries[1].Name)
	})
	
	t.Run("StoreCommit", func(t *testing.T) {
		commit := &CommitObject{
			RootTree:  Hash("tree123"),
			Author:    "test@example.com",
			Message:   "Initial commit",
			Timestamp: time.Now(),
			Version:   1,
		}
		
		hash, err := store.StoreCommit(ctx, commit)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		
		retrieved, err := store.GetCommit(ctx, hash)
		require.NoError(t, err)
		assert.Equal(t, commit.RootTree, retrieved.RootTree)
		assert.Equal(t, commit.Author, retrieved.Author)
		assert.Equal(t, commit.Message, retrieved.Message)
		assert.Equal(t, commit.Version, retrieved.Version)
	})
}

func TestVersionManager(t *testing.T) {
	backend := NewMemoryBackend()
	defer backend.Close()
	
	vm := NewVersionManager(backend)
	ctx := context.Background()
	
	t.Run("InitialState", func(t *testing.T) {
		version, err := vm.GetCurrentVersion(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), version)
		
		_, err = vm.GetLatestVersionInfo(ctx)
		assert.Error(t, err) // No versions exist
	})
	
	t.Run("CreateVersions", func(t *testing.T) {
		// Create first version
		commitHash1 := Hash("commit1hash")
		info1, err := vm.CreateVersion(ctx, commitHash1, "First commit")
		require.NoError(t, err)
		assert.Equal(t, int64(1), info1.Version)
		assert.Equal(t, commitHash1, info1.CommitHash)
		
		// Create second version
		commitHash2 := Hash("commit2hash")
		info2, err := vm.CreateVersion(ctx, commitHash2, "Second commit")
		require.NoError(t, err)
		assert.Equal(t, int64(2), info2.Version)
		assert.Equal(t, commitHash2, info2.CommitHash)
		
		// Check current version
		current, err := vm.GetCurrentVersion(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), current)
		
		// Check latest version info
		latest, err := vm.GetLatestVersionInfo(ctx)
		require.NoError(t, err)
		assert.Equal(t, info2.Version, latest.Version)
		assert.Equal(t, info2.CommitHash, latest.CommitHash)
	})
	
	t.Run("ListVersions", func(t *testing.T) {
		versions, err := vm.ListVersions(ctx, 10)
		require.NoError(t, err)
		assert.Len(t, versions, 2)
		
		// Should be in descending order (newest first)
		assert.Equal(t, int64(2), versions[0].Version)
		assert.Equal(t, int64(1), versions[1].Version)
		
		// Test limit
		limited, err := vm.ListVersions(ctx, 1)
		require.NoError(t, err)
		assert.Len(t, limited, 1)
		assert.Equal(t, int64(2), limited[0].Version)
	})
	
	t.Run("GetVersionInfo", func(t *testing.T) {
		info, err := vm.GetVersionInfo(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), info.Version)
		assert.Equal(t, "First commit", info.Message)
		
		_, err = vm.GetVersionInfo(ctx, 999)
		assert.Error(t, err) // Version doesn't exist
	})
}

func TestRepository(t *testing.T) {
	backend := NewMemoryBackend()
	defer backend.Close()
	
	repo := NewRepository(backend)
	defer repo.Close()
	
	ctx := context.Background()
	
	t.Run("InitialState", func(t *testing.T) {
		version, err := repo.GetCurrentVersion(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), version)
	})
	
	t.Run("ContentOperations", func(t *testing.T) {
		// Store a blob
		content := []byte("repository test content")
		hash, err := repo.StoreBlob(ctx, content)
		require.NoError(t, err)
		
		// Retrieve the blob
		blob, err := repo.GetBlob(ctx, hash)
		require.NoError(t, err)
		assert.Equal(t, content, blob.Content)
		
		// Create a tree
		tree := &TreeObject{
			Entries: []TreeEntry{
				{
					Name: "test.txt",
					Hash: hash,
					Type: ObjectTypeBlob,
					Mode: 0644,
					Size: int64(len(content)),
				},
			},
		}
		
		treeHash, err := repo.StoreTree(ctx, tree)
		require.NoError(t, err)
		
		// Create a commit
		commit := &CommitObject{
			RootTree:  treeHash,
			Author:    "test@example.com",
			Message:   "Test commit",
			Timestamp: time.Now(),
			Version:   1,
		}
		
		commitHash, err := repo.StoreCommit(ctx, commit)
		require.NoError(t, err)
		
		// Create a version
		versionInfo, err := repo.CreateVersion(ctx, commitHash, "Test commit")
		require.NoError(t, err)
		assert.Equal(t, int64(1), versionInfo.Version)
		assert.Equal(t, commitHash, versionInfo.CommitHash)
	})
}