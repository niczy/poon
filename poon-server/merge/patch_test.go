package merge

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatchParsing(t *testing.T) {
	t.Run("Valid Simple Patch", func(t *testing.T) {
		patchData := `--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,3 @@
 line 1
-line 2
+modified line 2
 line 3
`

		patch, err := ParsePatch([]byte(patchData))
		require.NoError(t, err)
		assert.Equal(t, "test.txt", patch.Header.OldFile)
		assert.Equal(t, "test.txt", patch.Header.NewFile)
		assert.Len(t, patch.Hunks, 1)

		hunk := patch.Hunks[0]
		assert.Equal(t, 1, hunk.OldStart)
		assert.Equal(t, 3, hunk.OldCount)
		assert.Equal(t, 1, hunk.NewStart)
		assert.Equal(t, 3, hunk.NewCount)
		assert.Len(t, hunk.Lines, 4)
	})

	t.Run("Invalid Patch Format", func(t *testing.T) {
		patchData := `not a valid patch`

		_, err := ParsePatch([]byte(patchData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not contain valid unified diff headers")
	})

	t.Run("Empty Patch", func(t *testing.T) {
		_, err := ParsePatch([]byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "patch data is empty")
	})

	t.Run("Multi-hunk Patch", func(t *testing.T) {
		patchData := `--- a/test.txt
+++ b/test.txt
@@ -1,2 +1,3 @@
 line 1
+new line
 line 2
@@ -10,1 +11,2 @@
 line 10
+another new line
`

		patch, err := ParsePatch([]byte(patchData))
		require.NoError(t, err)
		assert.Len(t, patch.Hunks, 2)

		assert.Equal(t, 1, patch.Hunks[0].OldStart)
		assert.Equal(t, 10, patch.Hunks[1].OldStart)
	})
}

func TestPatchValidation(t *testing.T) {
	t.Run("Valid Patch", func(t *testing.T) {
		patchData := `--- a/test.txt
+++ b/test.txt
@@ -1,1 +1,1 @@
-old
+new
`
		err := ValidatePatch([]byte(patchData))
		assert.NoError(t, err)
	})

	t.Run("Missing Headers", func(t *testing.T) {
		patchData := `@@ -1,1 +1,1 @@
-old
+new
`
		err := ValidatePatch([]byte(patchData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "patch has hunk without proper file headers")
	})

	t.Run("No Hunks", func(t *testing.T) {
		patchData := `--- a/test.txt
+++ b/test.txt
`
		err := ValidatePatch([]byte(patchData))
		assert.NoError(t, err) // Valid headers, no hunks is OK
	})
}

func TestPatchApplication(t *testing.T) {
	t.Run("Apply Simple Patch", func(t *testing.T) {
		// Create temporary file
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		originalContent := "line 1\nline 2\nline 3\n"
		err := os.WriteFile(testFile, []byte(originalContent), 0644)
		require.NoError(t, err)

		// Create patch
		patchData := `--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,3 @@
 line 1
-line 2
+modified line 2
 line 3
`

		patch, err := ParsePatch([]byte(patchData))
		require.NoError(t, err)

		// Apply patch
		err = ApplyPatch(testFile, patch)
		require.NoError(t, err)

		// Verify result
		result, err := os.ReadFile(testFile)
		require.NoError(t, err)
		expectedContent := "line 1\nmodified line 2\nline 3\n"
		assert.Equal(t, expectedContent, string(result))
	})

	t.Run("Create New File", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "new_file.txt")

		// Create patch for new file
		patchData := `--- /dev/null
+++ b/new_file.txt
@@ -0,0 +1,3 @@
+# New File
+
+This file was created by a patch.
`

		patch, err := ParsePatch([]byte(patchData))
		require.NoError(t, err)

		// Apply patch
		err = ApplyPatch(testFile, patch)
		require.NoError(t, err)

		// Verify result
		result, err := os.ReadFile(testFile)
		require.NoError(t, err)
		expectedContent := "# New File\n\nThis file was created by a patch.\n"
		assert.Equal(t, expectedContent, string(result))
	})

	t.Run("Multi-hunk Patch", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "config.yaml")
		originalContent := "environment: test\nservices:\n  frontend:\n    port: 3000\n  backend:\n    port: 8080\n"
		err := os.WriteFile(testFile, []byte(originalContent), 0644)
		require.NoError(t, err)

		// Create multi-hunk patch
		patchData := `--- a/config.yaml
+++ b/config.yaml
@@ -1,2 +1,3 @@
 environment: test
+version: 1.0
 services:
@@ -4,2 +5,3 @@
   backend:
     port: 8080
+    timeout: 30s
`

		patch, err := ParsePatch([]byte(patchData))
		require.NoError(t, err)

		// Apply patch
		err = ApplyPatch(testFile, patch)
		require.NoError(t, err)

		// Verify result
		result, err := os.ReadFile(testFile)
		require.NoError(t, err)
		resultStr := string(result)
		assert.Contains(t, resultStr, "version: 1.0")
		assert.Contains(t, resultStr, "timeout: 30s")
	})
}

func TestBackupFile(t *testing.T) {
	t.Run("Backup Existing File", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		originalContent := "original content"
		err := os.WriteFile(testFile, []byte(originalContent), 0644)
		require.NoError(t, err)

		// Create backup
		backupPath, err := BackupFile(testFile)
		require.NoError(t, err)
		assert.NotEmpty(t, backupPath)

		// Verify backup exists and has same content
		backupContent, err := os.ReadFile(backupPath)
		require.NoError(t, err)
		assert.Equal(t, originalContent, string(backupContent))

		// Clean up
		os.Remove(backupPath)
	})

	t.Run("Backup Nonexistent File", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "nonexistent.txt")

		// Try to backup nonexistent file
		backupPath, err := BackupFile(testFile)
		require.NoError(t, err)
		assert.Empty(t, backupPath) // Should return empty string for nonexistent files
	})
}
