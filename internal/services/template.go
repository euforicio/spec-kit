package services

import (
	"crypto/sha256"
	"encoding/json"
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
// Uses cache-first approach: checks cache validity, falls back to local templates if needed
func (t *TemplateService) DownloadAndExtract(aiAssistant, targetPath string, isHere bool) (*models.Template, error) {
	// First, try to use cached templates
	if err := t.extractFromCache(aiAssistant, targetPath, isHere); err == nil {
		// Cache extraction successful
		template := &models.Template{
			AIAssistant: aiAssistant,
			Version:     t.GetSpecKitVersion(),
			FileName:    "cached-template",
		}
		return template, nil
	}

	// Cache failed, fall back to GitHub download (original behavior)
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

	for _, aiAssistant := range models.ListAgents() {
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

// ResolveRoot returns the cache root directory path
func (t *TemplateService) ResolveRoot() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".spec-kit", "templates"), nil
}

// ReadManifest reads and parses the cache manifest file
func (t *TemplateService) ReadManifest() (*models.CacheManifest, error) {
	cacheRoot, err := t.ResolveRoot()
	if err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(cacheRoot, ".manifest.json")
	exists, err := t.filesystem.FileExists(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check manifest existence: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("manifest file does not exist")
	}

	content, err := t.filesystem.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest models.CacheManifest
	if err := json.Unmarshal([]byte(content), &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	return &manifest, nil
}

// WriteManifest writes the cache manifest to the cache directory
func (t *TemplateService) WriteManifest(manifest *models.CacheManifest) error {
	cacheRoot, err := t.ResolveRoot()
	if err != nil {
		return err
	}

	// Ensure cache directory exists
	if err := t.filesystem.CreateDirectory(cacheRoot); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest to JSON: %w", err)
	}

	manifestPath := filepath.Join(cacheRoot, ".manifest.json")
	if err := t.filesystem.WriteFile(manifestPath, string(manifestData)); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}

// CalculateFileHash calculates SHA256 hash of a file
func (t *TemplateService) CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for hashing: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to calculate file hash: %w", err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// ValidateCache validates the cache integrity by checking file hashes
func (t *TemplateService) ValidateCache(manifest *models.CacheManifest) error {
	cacheRoot, err := t.ResolveRoot()
	if err != nil {
		return err
	}

	for relativePath, expectedHash := range manifest.Templates {
		fullPath := filepath.Join(cacheRoot, relativePath)

		exists, err := t.filesystem.FileExists(fullPath)
		if err != nil {
			return fmt.Errorf("failed to check file existence for %s: %w", relativePath, err)
		}

		if !exists {
			return fmt.Errorf("cached file missing: %s", relativePath)
		}

		actualHash, err := t.CalculateFileHash(fullPath)
		if err != nil {
			return fmt.Errorf("failed to calculate hash for %s: %w", relativePath, err)
		}

		if actualHash != expectedHash {
			return fmt.Errorf("hash mismatch for %s: expected %s, got %s", relativePath, expectedHash, actualHash)
		}
	}

	return nil
}

// GetSpecKitVersion returns the current spec-kit version
func (t *TemplateService) GetSpecKitVersion() string {
	// This should be injected from build info, for now use a placeholder
	// In a real implementation, this would come from ldflags during build
	return "dev"
}

// extractFromCache extracts templates from cache to target path
func (t *TemplateService) extractFromCache(aiAssistant, targetPath string, isHere bool) error {
	// Read and validate cache
	manifest, err := t.ReadManifest()
	if err != nil {
		return fmt.Errorf("failed to read cache manifest: %w", err)
	}

	// Check version match
	currentVersion := t.GetSpecKitVersion()
	if !manifest.IsVersionMatch(currentVersion) {
		return fmt.Errorf("cache version mismatch: cache has %s, current is %s", manifest.SpecKitVersion, currentVersion)
	}

	// Validate cache integrity
	if err := t.ValidateCache(manifest); err != nil {
		return fmt.Errorf("cache validation failed: %w", err)
	}

	// Get cache root
	cacheRoot, err := t.ResolveRoot()
	if err != nil {
		return fmt.Errorf("failed to resolve cache root: %w", err)
	}

	// Create target directory if needed
	if !isHere {
		if err := t.filesystem.CreateDirectory(targetPath); err != nil {
			return fmt.Errorf("failed to create target directory: %w", err)
		}
	}

	// Copy shared files (memory and commands)
	if err := t.copySharedFiles(cacheRoot, targetPath); err != nil {
		return fmt.Errorf("failed to copy shared files: %w", err)
	}

	// Copy agent-specific template
	if err := t.copyAgentTemplate(cacheRoot, targetPath, aiAssistant); err != nil {
		return fmt.Errorf("failed to copy agent template: %w", err)
	}

	// Copy general templates
	if err := t.copyGeneralTemplates(cacheRoot, targetPath); err != nil {
		return fmt.Errorf("failed to copy general templates: %w", err)
	}

	return nil
}

// copySharedFiles copies memory and command files from cache to target
func (t *TemplateService) copySharedFiles(cacheRoot, targetPath string) error {
	// Copy memory files
	sourceMemoryDir := filepath.Join(cacheRoot, "memory")
	targetMemoryDir := filepath.Join(targetPath, "memory")

	memoryExists, err := t.filesystem.DirectoryExists(sourceMemoryDir)
	if err != nil {
		return fmt.Errorf("failed to check memory directory: %w", err)
	}

	if memoryExists {
		if err := t.filesystem.MergeDirectories(sourceMemoryDir, targetMemoryDir); err != nil {
			return fmt.Errorf("failed to copy memory files: %w", err)
		}
	}

	// Copy command files
	sourceCommandsDir := filepath.Join(cacheRoot, ".claude", "commands")
	targetCommandsDir := filepath.Join(targetPath, ".claude", "commands")

	commandsExists, err := t.filesystem.DirectoryExists(sourceCommandsDir)
	if err != nil {
		return fmt.Errorf("failed to check commands directory: %w", err)
	}

	if commandsExists {
		if err := t.filesystem.MergeDirectories(sourceCommandsDir, targetCommandsDir); err != nil {
			return fmt.Errorf("failed to copy command files: %w", err)
		}
	}

	return nil
}

// copyAgentTemplate copies the agent-specific template file
func (t *TemplateService) copyAgentTemplate(cacheRoot, targetPath, aiAssistant string) error {
	targetFile := filepath.Join(targetPath, "AGENT.md")

	// Try agent-specific template first
	sourceFile := filepath.Join(cacheRoot, "templates", fmt.Sprintf("agent_%s.template.md", aiAssistant))
	exists, err := t.filesystem.FileExists(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to check agent template: %w", err)
	}

	if !exists {
		// Fall back to generic agent-file-template.md
		sourceFile = filepath.Join(cacheRoot, "templates", "agent-file-template.md")
		exists, err = t.filesystem.FileExists(sourceFile)
		if err != nil {
			return fmt.Errorf("failed to check generic agent template: %w", err)
		}
		if !exists {
			return fmt.Errorf("no agent template found for %s", aiAssistant)
		}
	}

	// Copy agent template
	if err := t.filesystem.CopyFile(sourceFile, targetFile); err != nil {
		return fmt.Errorf("failed to copy agent template: %w", err)
	}

	return nil
}

// copyGeneralTemplates copies general template files
func (t *TemplateService) copyGeneralTemplates(cacheRoot, targetPath string) error {
	templateMappings := map[string]string{
		"tasks-template.md": "TASKS.md",
		"plan-template.md":  "PLAN.md",
		"spec-template.md":  "SPEC.md",
	}

	for sourceFilename, targetFilename := range templateMappings {
		sourceFile := filepath.Join(cacheRoot, "templates", sourceFilename)
		targetFile := filepath.Join(targetPath, targetFilename)

		// Check if template exists
		exists, err := t.filesystem.FileExists(sourceFile)
		if err != nil {
			return fmt.Errorf("failed to check template %s: %w", sourceFilename, err)
		}

		if !exists {
			// Skip missing templates (they're optional)
			continue
		}

		// Copy template
		if err := t.filesystem.CopyFile(sourceFile, targetFile); err != nil {
			return fmt.Errorf("failed to copy template %s: %w", sourceFilename, err)
		}
	}

	return nil
}
