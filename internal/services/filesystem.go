package services

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/euforicio/spec-kit/internal/models"
)

// FilesystemServiceInterface defines the interface for filesystem operations
type FilesystemServiceInterface interface {
	GetWorkingDirectory() (string, error)
	DirectoryExists(path string) (bool, error)
	IsDirectoryEmpty(path string) (bool, error)
	CreateDirectory(path string) error
	WriteFile(path, content string) error
	ReadFile(path string) (string, error)
	FileExists(path string) (bool, error)
	DownloadFile(url, destination string) error
	ExtractZip(src, dest string) error
	ListDirectory(path string) ([]string, error)
	CopyFile(src, dest string) error
}

// FilesystemService handles file and directory operations
type FilesystemService struct{}

// NewFilesystemService creates a new filesystem service instance
func NewFilesystemService() *FilesystemService {
	return &FilesystemService{}
}

// CreateDirectory creates a directory with the specified path
func (fs *FilesystemService) CreateDirectory(path string) error {
	if path == "" {
		return &models.ProjectError{
			Type:    models.ProjectPathInvalid,
			Path:    path,
			Message: "directory path cannot be empty",
		}
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		if os.IsPermission(err) {
			return &models.ProjectError{
				Type:    models.ProjectAccessDenied,
				Path:    path,
				Message: fmt.Sprintf("permission denied creating directory: %v", err),
				Cause:   err,
			}
		}
		return &models.ProjectError{
			Type:    models.ProjectPathInvalid,
			Path:    path,
			Message: fmt.Sprintf("failed to create directory: %v", err),
			Cause:   err,
		}
	}

	return nil
}

// DirectoryExists checks if a directory exists at the specified path
func (fs *FilesystemService) DirectoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check directory existence: %w", err)
	}
	return info.IsDir(), nil
}

// IsDirectoryEmpty checks if a directory is empty
func (fs *FilesystemService) IsDirectoryEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("failed to read directory: %w", err)
	}
	return len(entries) == 0, nil
}

// ListDirectoryContents returns the contents of a directory
func (fs *FilesystemService) ListDirectoryContents(path string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory contents: %w", err)
	}
	return entries, nil
}

// ExtractZIP extracts a ZIP file to the specified destination directory
func (fs *FilesystemService) ExtractZIP(zipPath, destDir string) error {
	// Open ZIP file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return &models.TemplateError{
			Type:    models.TemplateExtractionFailed,
			Message: fmt.Sprintf("failed to open ZIP file: %v", err),
			Cause:   err,
		}
	}
	defer reader.Close()

	// Create destination directory
	err = fs.CreateDirectory(destDir)
	if err != nil {
		return err
	}

	// Extract files
	for _, file := range reader.File {
		err := fs.extractZIPFile(file, destDir)
		if err != nil {
			return &models.TemplateError{
				Type:    models.TemplateExtractionFailed,
				Message: fmt.Sprintf("failed to extract file %s: %v", file.Name, err),
				Cause:   err,
			}
		}
	}

	return nil
}

// extractZIPFile extracts a single file from a ZIP archive
func (fs *FilesystemService) extractZIPFile(file *zip.File, destDir string) error {
	// Validate file path to prevent directory traversal
	if err := fs.validateZIPFilePath(file.Name); err != nil {
		return err
	}

	path := filepath.Join(destDir, file.Name)

	// Ensure the file path is within the destination directory
	if !strings.HasPrefix(path, filepath.Clean(destDir)+string(os.PathSeparator)) &&
	   path != filepath.Clean(destDir) {
		return fmt.Errorf("invalid file path: %s (potential directory traversal)", file.Name)
	}

	if file.FileInfo().IsDir() {
		return fs.CreateDirectory(path)
	}

	// Create parent directories if they don't exist
	if err := fs.CreateDirectory(filepath.Dir(path)); err != nil {
		return err
	}

	// Open file in ZIP archive
	fileReader, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in ZIP: %w", err)
	}
	defer fileReader.Close()

	// Create destination file
	targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer targetFile.Close()

	// Copy file contents
	_, err = io.Copy(targetFile, fileReader)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}

// validateZIPFilePath validates a file path from a ZIP archive
func (fs *FilesystemService) validateZIPFilePath(path string) error {
	// Check for directory traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("invalid file path: %s (contains '..')", path)
	}

	// Check for absolute paths
	if filepath.IsAbs(path) {
		return fmt.Errorf("invalid file path: %s (absolute path not allowed)", path)
	}

	return nil
}

// MergeDirectories copies files from source directory to destination directory
func (fs *FilesystemService) MergeDirectories(srcDir, destDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking source directory: %w", err)
		}

		// Calculate relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path: %w", err)
		}

		destPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return fs.CreateDirectory(destPath)
		}

		return fs.copyFile(path, destPath, info.Mode())
	})
}

// copyFile copies a single file from source to destination
func (fs *FilesystemService) copyFile(srcPath, destPath string, mode os.FileMode) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create parent directory if it doesn't exist
	if err := fs.CreateDirectory(filepath.Dir(destPath)); err != nil {
		return err
	}

	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}

// RemoveFile removes a file at the specified path
func (fs *FilesystemService) RemoveFile(path string) error {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove file: %w", err)
	}
	return nil
}

// RemoveDirectory removes a directory and all its contents
func (fs *FilesystemService) RemoveDirectory(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}
	return nil
}

// WriteFile writes content to a file
func (fs *FilesystemService) WriteFile(path, content string) error {
	// Create parent directory if it doesn't exist
	if err := fs.CreateDirectory(filepath.Dir(path)); err != nil {
		return err
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// ReadFile reads content from a file as string
func (fs *FilesystemService) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(data), nil
}

// GetWorkingDirectory returns the current working directory
func (fs *FilesystemService) GetWorkingDirectory() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return wd, nil
}

// IsWritable checks if a directory is writable
func (fs *FilesystemService) IsWritable(path string) error {
	// Try to create a temporary file in the directory
	testFile := filepath.Join(path, ".write_test_tmp")
	file, err := os.OpenFile(testFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if os.IsPermission(err) {
			return &models.ProjectError{
				Type:    models.ProjectAccessDenied,
				Path:    path,
				Message: "directory is not writable",
				Cause:   err,
			}
		}
		return fmt.Errorf("failed to test directory writability: %w", err)
	}
	file.Close()

	// Clean up test file
	os.Remove(testFile)
	return nil
}

// GetDirectorySize calculates the total size of a directory and its contents
func (fs *FilesystemService) GetDirectorySize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// CreateTempDirectory creates a temporary directory
func (fs *FilesystemService) CreateTempDirectory(prefix string) (string, error) {
	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	return tempDir, nil
}

// ExtractZIPWithFlatten extracts a ZIP file and flattens single root directories
func (fs *FilesystemService) ExtractZIPWithFlatten(zipPath, destDir string) error {
	// First extract to a temporary directory
	tempDir, err := fs.CreateTempDirectory("specify-extract-")
	if err != nil {
		return err
	}
	defer fs.RemoveDirectory(tempDir)

	// Extract ZIP to temp directory
	if err := fs.ExtractZIP(zipPath, tempDir); err != nil {
		return err
	}

	// Check if we have a single root directory (GitHub-style ZIP)
	entries, err := fs.ListDirectoryContents(tempDir)
	if err != nil {
		return fmt.Errorf("failed to list extracted contents: %w", err)
	}

	sourceDir := tempDir
	if len(entries) == 1 && entries[0].IsDir() {
		// Single root directory - use its contents
		sourceDir = filepath.Join(tempDir, entries[0].Name())
	}

	// Create destination directory
	if err := fs.CreateDirectory(destDir); err != nil {
		return err
	}

	// Copy/merge contents to final destination
	return fs.MergeDirectories(sourceDir, destDir)
}

// FileExists checks if a file exists
func (fs *FilesystemService) FileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return !info.IsDir(), nil
}

// DownloadFile downloads a file from a URL (stub for interface compatibility)
func (fs *FilesystemService) DownloadFile(url, destination string) error {
	return fmt.Errorf("download functionality not implemented in filesystem service")
}

// ExtractZip extracts a ZIP file (alias for ExtractZIP)
func (fs *FilesystemService) ExtractZip(src, dest string) error {
	return fs.ExtractZIP(src, dest)
}

// ListDirectory returns directory names as strings
func (fs *FilesystemService) ListDirectory(path string) ([]string, error) {
	entries, err := fs.ListDirectoryContents(path)
	if err != nil {
		return nil, err
	}
	
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names, nil
}

// CopyFile copies a file from source to destination
func (fs *FilesystemService) CopyFile(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}
	
	return fs.copyFile(src, dest, info.Mode())
}