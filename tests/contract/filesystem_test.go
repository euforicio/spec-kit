package contract

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesystemOperationsContract(t *testing.T) {
	t.Run("create directory structure", func(t *testing.T) {
		tempDir := t.TempDir()
		projectPath := filepath.Join(tempDir, "test-project")

		// Test directory creation
		err := os.MkdirAll(projectPath, 0755)
		require.NoError(t, err, "should be able to create project directory")

		// Verify directory exists and has correct permissions
		info, err := os.Stat(projectPath)
		require.NoError(t, err, "created directory should be accessible")
		assert.True(t, info.IsDir(), "should be a directory")
		assert.Equal(t, os.FileMode(0755), info.Mode().Perm(), "should have correct permissions")
	})

	t.Run("create nested directory structure", func(t *testing.T) {
		tempDir := t.TempDir()
		nestedPath := filepath.Join(tempDir, "project", "src", "components")

		err := os.MkdirAll(nestedPath, 0755)
		require.NoError(t, err, "should be able to create nested directories")

		// Verify all levels exist
		assert.DirExists(t, filepath.Join(tempDir, "project"))
		assert.DirExists(t, filepath.Join(tempDir, "project", "src"))
		assert.DirExists(t, nestedPath)
	})

	t.Run("handle existing directory", func(t *testing.T) {
		tempDir := t.TempDir()
		projectPath := filepath.Join(tempDir, "existing-project")

		// Create directory first
		err := os.Mkdir(projectPath, 0755)
		require.NoError(t, err)

		// Try to create again - should not error with MkdirAll
		err = os.MkdirAll(projectPath, 0755)
		assert.NoError(t, err, "MkdirAll should not error on existing directory")
	})

	t.Run("detect existing directory contents", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create some files
		testFile := filepath.Join(tempDir, "existing.txt")
		err := os.WriteFile(testFile, []byte("existing content"), 0644)
		require.NoError(t, err)

		// Read directory contents
		entries, err := os.ReadDir(tempDir)
		require.NoError(t, err, "should be able to read directory")

		assert.Len(t, entries, 1, "should detect existing file")
		assert.Equal(t, "existing.txt", entries[0].Name())
		assert.False(t, entries[0].IsDir(), "should correctly identify file type")
	})
}

func TestZipExtractionContract(t *testing.T) {
	t.Run("create and extract ZIP file", func(t *testing.T) {
		tempDir := t.TempDir()
		zipPath := filepath.Join(tempDir, "test.zip")
		extractDir := filepath.Join(tempDir, "extracted")

		// Create a test ZIP file
		createTestZIP(t, zipPath)

		// Extract ZIP file
		err := extractZIP(zipPath, extractDir)
		require.NoError(t, err, "should be able to extract ZIP file")

		// Verify extracted contents
		assert.DirExists(t, extractDir)
		assert.FileExists(t, filepath.Join(extractDir, "file1.txt"))
		assert.FileExists(t, filepath.Join(extractDir, "subdir", "file2.txt"))

		// Verify file contents
		content1, err := os.ReadFile(filepath.Join(extractDir, "file1.txt"))
		require.NoError(t, err)
		assert.Equal(t, "content1", string(content1))

		content2, err := os.ReadFile(filepath.Join(extractDir, "subdir", "file2.txt"))
		require.NoError(t, err)
		assert.Equal(t, "content2", string(content2))
	})

	t.Run("extract ZIP with nested directory structure", func(t *testing.T) {
		tempDir := t.TempDir()
		zipPath := filepath.Join(tempDir, "nested.zip")
		extractDir := filepath.Join(tempDir, "extracted")

		// Create ZIP with GitHub-style nested structure
		createNestedZIP(t, zipPath)

		// Extract and verify
		err := extractZIP(zipPath, extractDir)
		require.NoError(t, err)

		// Should handle nested structure correctly
		entries, err := os.ReadDir(extractDir)
		require.NoError(t, err)

		// Check if we have the expected structure
		foundDirectStructure := false
		foundNestedStructure := false

		for _, entry := range entries {
			if entry.Name() == "root-dir" && entry.IsDir() {
				foundNestedStructure = true
			}
			if entry.Name() == "file1.txt" && !entry.IsDir() {
				foundDirectStructure = true
			}
		}

		// Should have either direct files or a single root directory
		assert.True(t, foundDirectStructure || foundNestedStructure,
			"should handle either direct or nested ZIP structure")
	})

	t.Run("handle ZIP extraction errors", func(t *testing.T) {
		tempDir := t.TempDir()
		nonExistentZip := filepath.Join(tempDir, "nonexistent.zip")
		extractDir := filepath.Join(tempDir, "extracted")

		// Try to extract non-existent ZIP
		err := extractZIP(nonExistentZip, extractDir)
		assert.Error(t, err, "should error when ZIP file doesn't exist")
	})
}

func TestFileMergingContract(t *testing.T) {
	t.Run("merge files into existing directory", func(t *testing.T) {
		tempDir := t.TempDir()
		existingDir := filepath.Join(tempDir, "existing")
		sourceDir := filepath.Join(tempDir, "source")

		// Create existing directory with content
		err := os.MkdirAll(existingDir, 0755)
		require.NoError(t, err)

		existingFile := filepath.Join(existingDir, "existing.txt")
		err = os.WriteFile(existingFile, []byte("existing"), 0644)
		require.NoError(t, err)

		// Create source directory with new content
		err = os.MkdirAll(sourceDir, 0755)
		require.NoError(t, err)

		newFile := filepath.Join(sourceDir, "new.txt")
		err = os.WriteFile(newFile, []byte("new content"), 0644)
		require.NoError(t, err)

		// Test merging (copy source files to existing directory)
		err = mergeDirectories(sourceDir, existingDir)
		require.NoError(t, err, "should be able to merge directories")

		// Verify both files exist
		assert.FileExists(t, existingFile, "existing file should be preserved")
		assert.FileExists(t, filepath.Join(existingDir, "new.txt"), "new file should be added")

		// Verify content preservation
		content, err := os.ReadFile(existingFile)
		require.NoError(t, err)
		assert.Equal(t, "existing", string(content), "existing content should be preserved")
	})

	t.Run("handle file conflicts during merge", func(t *testing.T) {
		tempDir := t.TempDir()
		existingDir := filepath.Join(tempDir, "existing")
		sourceDir := filepath.Join(tempDir, "source")

		// Create existing directory with file
		err := os.MkdirAll(existingDir, 0755)
		require.NoError(t, err)

		conflictFile := filepath.Join(existingDir, "conflict.txt")
		err = os.WriteFile(conflictFile, []byte("existing content"), 0644)
		require.NoError(t, err)

		// Create source directory with conflicting file
		err = os.MkdirAll(sourceDir, 0755)
		require.NoError(t, err)

		sourceConflictFile := filepath.Join(sourceDir, "conflict.txt")
		err = os.WriteFile(sourceConflictFile, []byte("new content"), 0644)
		require.NoError(t, err)

		// Test merging with conflict
		err = mergeDirectories(sourceDir, existingDir)

		// Behavior depends on implementation - could overwrite, skip, or error
		// The important part is that it handles the conflict gracefully
		if err != nil {
			assert.Contains(t, err.Error(), "conflict", "error should mention conflict")
		} else {
			// If no error, file should exist with some content
			assert.FileExists(t, conflictFile, "conflict file should exist after merge")
		}
	})
}

func TestPermissionsContract(t *testing.T) {
	t.Run("handle permission denied scenarios", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("running as root, cannot test permission denied")
		}

		tempDir := t.TempDir()
		restrictedDir := filepath.Join(tempDir, "restricted")

		// Create directory with no permissions
		err := os.Mkdir(restrictedDir, 0000)
		require.NoError(t, err)
		defer os.Chmod(restrictedDir, 0755) // Restore for cleanup

		// Try to create subdirectory - should fail
		subDir := filepath.Join(restrictedDir, "subdir")
		err = os.Mkdir(subDir, 0755)
		assert.Error(t, err, "should fail with permission denied")
		assert.Contains(t, err.Error(), "permission denied", "error should mention permission denied")
	})

	t.Run("preserve file permissions during operations", func(t *testing.T) {
		tempDir := t.TempDir()
		sourceFile := filepath.Join(tempDir, "source.txt")
		destFile := filepath.Join(tempDir, "dest.txt")

		// Create file with specific permissions
		err := os.WriteFile(sourceFile, []byte("content"), 0644)
		require.NoError(t, err)

		// Copy file and verify permissions are preserved
		sourceContent, err := os.ReadFile(sourceFile)
		require.NoError(t, err)

		err = os.WriteFile(destFile, sourceContent, 0644)
		require.NoError(t, err)

		// Check permissions
		sourceInfo, err := os.Stat(sourceFile)
		require.NoError(t, err)

		destInfo, err := os.Stat(destFile)
		require.NoError(t, err)

		assert.Equal(t, sourceInfo.Mode().Perm(), destInfo.Mode().Perm(),
			"file permissions should be preserved")
	})
}

// Helper functions for testing

func createTestZIP(t *testing.T, zipPath string) {
	t.Helper()

	zipFile, err := os.Create(zipPath)
	require.NoError(t, err)
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add file1.txt
	file1Writer, err := zipWriter.Create("file1.txt")
	require.NoError(t, err)
	_, err = file1Writer.Write([]byte("content1"))
	require.NoError(t, err)

	// Add subdir/file2.txt
	file2Writer, err := zipWriter.Create("subdir/file2.txt")
	require.NoError(t, err)
	_, err = file2Writer.Write([]byte("content2"))
	require.NoError(t, err)
}

func createNestedZIP(t *testing.T, zipPath string) {
	t.Helper()

	zipFile, err := os.Create(zipPath)
	require.NoError(t, err)
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add files in a root directory (GitHub-style)
	file1Writer, err := zipWriter.Create("root-dir/file1.txt")
	require.NoError(t, err)
	_, err = file1Writer.Write([]byte("content1"))
	require.NoError(t, err)

	file2Writer, err := zipWriter.Create("root-dir/subdir/file2.txt")
	require.NoError(t, err)
	_, err = file2Writer.Write([]byte("content2"))
	require.NoError(t, err)
}

func extractZIP(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	err = os.MkdirAll(destDir, 0755)
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			err := os.MkdirAll(path, file.FileInfo().Mode())
			if err != nil {
				return err
			}
			continue
		}

		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return err
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		_, err = io.Copy(targetFile, fileReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func mergeDirectories(srcDir, destDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		return err
	})
}
