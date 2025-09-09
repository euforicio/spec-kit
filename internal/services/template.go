package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/euforicio/spec-kit/internal/models"
)

// TemplateService handles template downloading and extraction
type TemplateService struct {
	github     *GitHubService
	filesystem *FilesystemService
}

// NewTemplateService creates a new template service instance
func NewTemplateService(github *GitHubService, filesystem *FilesystemService) *TemplateService {
	return &TemplateService{
		github:     github,
		filesystem: filesystem,
	}
}

// DownloadAndExtract downloads a template for the specified AI assistant and extracts it to the target path
func (t *TemplateService) DownloadAndExtract(aiAssistant, targetPath string, isHere bool) (*models.Template, error) {
	// Get template information
	template, err := t.github.GetTemplateForAI(aiAssistant)
	if err != nil {
		return nil, err
	}

	// Create temporary directory for download
	tempDir, err := t.filesystem.CreateTempDirectory("specify-download-")
	if err != nil {
		return nil, &models.TemplateError{
			Type:      models.TemplateDownloadFailed,
			Assistant: aiAssistant,
			Message:   fmt.Sprintf("failed to create temporary directory: %v", err),
			Cause:     err,
		}
	}
	defer t.filesystem.RemoveDirectory(tempDir)

	// Download template
	zipPath := filepath.Join(tempDir, template.FileName)
	if err := t.downloadTemplate(template, zipPath); err != nil {
		return nil, err
	}

	// Extract template
	if err := t.extractTemplate(zipPath, targetPath, isHere); err != nil {
		return nil, err
	}

	return template, nil
}

// downloadTemplate downloads the template ZIP file
func (t *TemplateService) downloadTemplate(template *models.Template, zipPath string) error {
	// Create ZIP file
	file, err := os.Create(zipPath)
	if err != nil {
		return &models.TemplateError{
			Type:      models.TemplateDownloadFailed,
			Assistant: template.AIAssistant,
			Version:   template.Version,
			Message:   fmt.Sprintf("failed to create download file: %v", err),
			Cause:     err,
		}
	}
	defer file.Close()

	// Get release information again to get asset details
	release, err := t.github.GetLatestRelease()
	if err != nil {
		return err
	}

	// Find the asset
	asset, err := t.github.FindTemplateAsset(release, template.AIAssistant)
	if err != nil {
		return err
	}

	// Download the asset
	if err := t.github.DownloadAsset(asset, file); err != nil {
		return err
	}

	return nil
}

// extractTemplate extracts the template to the target path
func (t *TemplateService) extractTemplate(zipPath, targetPath string, isHere bool) error {
	if isHere {
		// Extract directly to target path with flattening
		return t.filesystem.ExtractZIPWithFlatten(zipPath, targetPath)
	}

	// Create target directory first
	if err := t.filesystem.CreateDirectory(targetPath); err != nil {
		return err
	}

	// Extract with flattening to handle GitHub-style nested directories
	return t.filesystem.ExtractZIPWithFlatten(zipPath, targetPath)
}

// GetAvailableTemplates returns information about available templates
func (t *TemplateService) GetAvailableTemplates() (map[string]*models.Template, error) {
	templates := make(map[string]*models.Template)

	for _, aiAssistant := range models.ValidAIAssistants {
		template, err := t.github.GetTemplateForAI(aiAssistant)
		if err != nil {
			// Log error but continue with other templates
			continue
		}
		templates[aiAssistant] = template
	}

	if len(templates) == 0 {
		return nil, &models.TemplateError{
			Type:    models.TemplateNotFound,
			Message: "no templates available",
		}
	}

	return templates, nil
}

// ValidateTemplate checks if a template is valid and available
func (t *TemplateService) ValidateTemplate(aiAssistant string) error {
	_, err := t.github.GetTemplateForAI(aiAssistant)
	return err
}

// DownloadTemplateWithProgress downloads a template with progress tracking
func (t *TemplateService) DownloadTemplateWithProgress(aiAssistant, targetPath string, isHere bool, progressCallback func(downloaded, total int64)) (*models.Template, error) {
	// Get template information
	template, err := t.github.GetTemplateForAI(aiAssistant)
	if err != nil {
		return nil, err
	}

	// Create temporary directory for download
	tempDir, err := t.filesystem.CreateTempDirectory("specify-download-")
	if err != nil {
		return nil, &models.TemplateError{
			Type:      models.TemplateDownloadFailed,
			Assistant: aiAssistant,
			Message:   fmt.Sprintf("failed to create temporary directory: %v", err),
			Cause:     err,
		}
	}
	defer t.filesystem.RemoveDirectory(tempDir)

	// Download with progress
	zipPath := filepath.Join(tempDir, template.FileName)
	if err := t.downloadTemplateWithProgress(template, zipPath, progressCallback); err != nil {
		return nil, err
	}

	// Extract template
	if err := t.extractTemplate(zipPath, targetPath, isHere); err != nil {
		return nil, err
	}

	return template, nil
}

// downloadTemplateWithProgress downloads the template with progress tracking
func (t *TemplateService) downloadTemplateWithProgress(template *models.Template, zipPath string, progressCallback func(downloaded, total int64)) error {
	// Create ZIP file
	file, err := os.Create(zipPath)
	if err != nil {
		return &models.TemplateError{
			Type:      models.TemplateDownloadFailed,
			Assistant: template.AIAssistant,
			Version:   template.Version,
			Message:   fmt.Sprintf("failed to create download file: %v", err),
			Cause:     err,
		}
	}
	defer file.Close()

	// Get release information again to get asset details
	release, err := t.github.GetLatestRelease()
	if err != nil {
		return err
	}

	// Find the asset
	asset, err := t.github.FindTemplateAsset(release, template.AIAssistant)
	if err != nil {
		return err
	}

	// Create a progress writer
	progressWriter := &progressWriter{
		writer:   file,
		total:    asset.Size,
		callback: progressCallback,
	}

	// Download the asset with progress tracking
	if err := t.github.DownloadAsset(asset, progressWriter); err != nil {
		return err
	}

	return nil
}

// progressWriter wraps an io.Writer to track download progress
type progressWriter struct {
	writer     io.Writer
	total      int64
	downloaded int64
	callback   func(downloaded, total int64)
}

// Write implements io.Writer and tracks progress
func (pw *progressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.downloaded += int64(n)
	if pw.callback != nil {
		pw.callback(pw.downloaded, pw.total)
	}

	return n, nil
}

// GetTemplateInfo returns information about a specific template without downloading
func (t *TemplateService) GetTemplateInfo(aiAssistant string) (*models.Template, error) {
	return t.github.GetTemplateForAI(aiAssistant)
}

// CheckTemplateUpdates checks if there are newer versions of templates available
func (t *TemplateService) CheckTemplateUpdates(currentVersions map[string]string) (map[string]*models.Template, error) {
	updates := make(map[string]*models.Template)

	for aiAssistant, currentVersion := range currentVersions {
		template, err := t.github.GetTemplateForAI(aiAssistant)
		if err != nil {
			continue
		}

		if template.Version != currentVersion {
			updates[aiAssistant] = template
		}
	}

	return updates, nil
}

// CleanupTemporaryFiles removes any temporary files created during template operations
func (t *TemplateService) CleanupTemporaryFiles() error {
	// This would be implemented to clean up any leftover temporary files
	// For now, we rely on the defer cleanup in the methods above
	return nil
}

// EstimateDownloadTime estimates download time based on template size and connection speed
func (t *TemplateService) EstimateDownloadTime(template *models.Template, speedBytesPerSecond int64) (estimatedSeconds int64) {
	if speedBytesPerSecond <= 0 {
		// Assume a conservative 1 Mbps connection
		speedBytesPerSecond = 125000 // 1 Mbps = 125 KB/s
	}

	return template.Size / speedBytesPerSecond
}

// ValidateExtractedTemplate performs basic validation on an extracted template
func (t *TemplateService) ValidateExtractedTemplate(path string) error {
	// Check if the directory exists
	exists, err := t.filesystem.DirectoryExists(path)
	if err != nil {
		return fmt.Errorf("failed to check template directory: %w", err)
	}

	if !exists {
		return &models.TemplateError{
			Type:    models.TemplateExtractionFailed,
			Message: "template directory does not exist after extraction",
		}
	}

	// Check if directory is not empty
	isEmpty, err := t.filesystem.IsDirectoryEmpty(path)
	if err != nil {
		return fmt.Errorf("failed to check if template directory is empty: %w", err)
	}

	if isEmpty {
		return &models.TemplateError{
			Type:    models.TemplateExtractionFailed,
			Message: "template directory is empty after extraction",
		}
	}

	// Additional validation could be added here:
	// - Check for required files (e.g., CONSTITUTION.md)
	// - Validate file permissions
	// - Check for malicious files

	return nil
}