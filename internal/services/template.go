package services

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/euforicio/spec-kit/internal/models"
	"github.com/euforicio/spec-kit/internal/template"
)


// TemplateService handles template downloading and extraction
type TemplateService struct {
	github     *GitHubService
	filesystem *FilesystemService
	processor  *template.Processor
}

// NewTemplateService creates a new template service instance
func NewTemplateService(github *GitHubService, filesystem *FilesystemService) *TemplateService {
	return &TemplateService{
		github:     github,
		filesystem: filesystem,
		processor:  template.NewProcessor(),
	}
}

// processTemplate processes a template string with the given data
func (t *TemplateService) processTemplate(content string, data template.Data) (string, error) {
	return t.processor.Process(content, data)
}

// processTemplateFile processes a template file and returns the processed content
func (t *TemplateService) processTemplateFile(filePath string, data template.Data) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", filePath, err)
	}

	return t.processor.ProcessWithName(filePath, string(content), data)
}

// isContentTemplate checks if a file path is in the content/ directory (special processing)
func (t *TemplateService) isContentTemplate(relPath string) bool {
	return strings.HasPrefix(relPath, "content/")
}

// processTemplateDirectory walks a directory and processes templates while copying files
func (t *TemplateService) processTemplateDirectory(sourceDir, targetDir, aiAssistant string) error {
	data := template.Data{
		AIAssistant: aiAssistant,
	}

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking directory: %w", err)
		}

		// Calculate relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path: %w", err)
		}

		// Skip if this is the source directory itself
		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(targetDir, relPath)

		// Create directory if it's a directory
		if info.IsDir() {
			return t.filesystem.CreateDirectory(targetPath)
		}

		// Handle file processing
		return t.processAndCopyFile(path, targetPath, relPath, data)
	})
}

// processAndCopyFile processes a template file or copies it directly based on its type
func (t *TemplateService) processAndCopyFile(sourcePath, targetPath, relPath string, data template.Data) error {
	// Check if this is a template file that should be processed
	if t.shouldProcessAsTemplate(sourcePath) {
		// Process template
		processedContent, err := t.processTemplateFile(sourcePath, data)
		if err != nil {
			return fmt.Errorf("failed to process template %s: %w", sourcePath, err)
		}

		// Write processed content
		return t.filesystem.WriteFile(targetPath, processedContent)
	}

	// Copy file directly (binary files, etc.)
	return t.filesystem.CopyFile(sourcePath, targetPath)
}

// shouldProcessAsTemplate determines if a file should be processed as a template
func (t *TemplateService) shouldProcessAsTemplate(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Skip binary file extensions that should never be processed as templates
	binaryExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".ico": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mov": true, ".wav": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
		".zip": true, ".tar": true, ".gz": true, ".7z": true, ".rar": true,
		".bin": true, ".dat": true, ".db": true, ".sqlite": true,
	}
	
	if binaryExts[ext] {
		return false
	}
	
	// Process all other files as potential templates (they may contain {{.xxx}} syntax)
	// This includes .go, .js, .ts, .py, .sh, .md, .txt, .yaml, .yml, .json, etc.
	return true
}

// DownloadAndExtract downloads a template for the specified AI assistant and extracts it to the target path
// Uses cache-first approach: checks cache validity, falls back to local templates if needed
func (t *TemplateService) DownloadAndExtract(aiAssistant, targetPath string, isHere bool) (*models.Template, error) {
	// First, check if cache exists and try to use it
	cacheRoot, err := t.ResolveRoot()
	if err == nil {
		if isEmpty, _ := t.isCacheEmpty(cacheRoot); !isEmpty {
			if err := t.extractFromCache(aiAssistant, targetPath, isHere); err == nil {
				// Cache extraction successful
				template := &models.Template{
					Version:  t.GetSpecKitVersion(),
					FileName: "cached-template",
				}
				return template, nil
			}
		}
	}

	// Cache is empty or invalid - automatically sync templates
	if err := t.autoSyncTemplates(); err != nil {
		return nil, fmt.Errorf("%w: failed to sync templates automatically. Please run 'specify templates sync' manually: %v", models.ErrTemplateNotFound, err)
	}

	// Now try to extract from newly synced cache
	if err := t.extractFromCache(aiAssistant, targetPath, isHere); err != nil {
		return nil, fmt.Errorf("%w: failed to extract from synced cache: %v", models.ErrTemplateExtractionFailed, err)
	}

	template := &models.Template{
		Version:  t.GetSpecKitVersion(),
		FileName: "cached-template",
	}
	return template, nil
}

// isCacheEmpty checks if the cache directory is empty or doesn't exist
func (t *TemplateService) isCacheEmpty(cacheRoot string) (bool, error) {
	// Check if cache directory exists
	exists, err := t.filesystem.DirectoryExists(cacheRoot)
	if err != nil {
		return true, err
	}
	if !exists {
		return true, nil
	}

	// Check if manifest exists
	manifestPath := filepath.Join(cacheRoot, ".manifest.json")
	manifestExists, err := t.filesystem.FileExists(manifestPath)
	if err != nil {
		return true, err
	}
	if !manifestExists {
		return true, nil
	}

	// Check if there are any template files
	entries, err := t.filesystem.ListDirectoryContents(cacheRoot)
	if err != nil {
		return true, err
	}

	// Count non-manifest files
	fileCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != ".manifest.json" {
			fileCount++
		}
	}

	return fileCount == 0, nil
}

// autoSyncTemplates automatically downloads and syncs templates from GitHub
func (t *TemplateService) autoSyncTemplates() error {
	// Use the existing sync logic from templates CLI
	// Create temporary directory for download
	tempDir, err := t.filesystem.CreateTempDirectory("specify-auto-sync-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer t.filesystem.RemoveDirectory(tempDir)

	// Get latest release and download template asset
	release, err := t.github.GetLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	// Look for cache template asset
	var cacheAsset *GitHubAsset
	for _, asset := range release.Assets {
		if asset.Name == "spec-kit-cache-template.zip" {
			cacheAsset = &asset
			break
		}
	}

	if cacheAsset == nil {
		return fmt.Errorf("cache template asset not found in release %s", release.TagName)
	}

	// Download the asset
	zipPath := filepath.Join(tempDir, cacheAsset.Name)
	file, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create download file: %w", err)
	}
	defer file.Close()

	if err := t.github.DownloadAsset(cacheAsset, file); err != nil {
		return fmt.Errorf("failed to download template asset: %w", err)
	}
	file.Close() // Close before extraction

	// Extract ZIP to temporary directory
	extractDir := filepath.Join(tempDir, "extracted")
	if err := t.filesystem.ExtractZIPWithFlatten(zipPath, extractDir); err != nil {
		return fmt.Errorf("failed to extract template ZIP: %w", err)
	}

	// Get cache root and copy extracted content
	cacheRoot, err := t.ResolveRoot()
	if err != nil {
		return fmt.Errorf("failed to resolve cache root: %w", err)
	}

	// Ensure cache directory exists
	if err := t.filesystem.CreateDirectory(cacheRoot); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Copy extracted templates to cache
	if err := t.filesystem.MergeDirectories(extractDir, cacheRoot); err != nil {
		return fmt.Errorf("failed to copy templates to cache: %w", err)
	}

	return nil
}

// downloadTemplate downloads the template ZIP file
func (t *TemplateService) downloadTemplate(template *models.Template, zipPath string) error {
	// Create ZIP file
	file, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("%w: failed to create download file for version %s: %v", models.ErrTemplateDownloadFailed, template.Version, err)
	}
	defer file.Close()

	// Get release information again to get asset details
	release, err := t.github.GetLatestRelease()
	if err != nil {
		return err
	}

	// Find the asset
	asset, err := t.github.FindTemplateAsset(release)
	if err != nil {
		return err
	}

	// Download the asset
	if err := t.github.DownloadAsset(asset, file); err != nil {
		return err
	}

	return nil
}

// extractTemplate extracts the template to the target path using unified template copying logic
func (t *TemplateService) extractTemplate(zipPath, targetPath string, isHere bool, aiAssistant string) error {
	// Create temporary directory for extraction
	tempDir, err := t.filesystem.CreateTempDirectory("specify-extract-")
	if err != nil {
		return fmt.Errorf("failed to create temporary extraction directory: %w", err)
	}
	defer t.filesystem.RemoveDirectory(tempDir)

	// Extract ZIP to temporary directory first
	if err := t.filesystem.ExtractZIPWithFlatten(zipPath, tempDir); err != nil {
		return fmt.Errorf("failed to extract ZIP: %w", err)
	}

	// Create target directory if needed
	if !isHere {
		if err := t.filesystem.CreateDirectory(targetPath); err != nil {
			return fmt.Errorf("failed to create target directory: %w", err)
		}
	}

	// Apply unified template copying logic to extracted files
	return t.applyUnifiedTemplateExtraction(tempDir, targetPath, isHere, aiAssistant)
}

// applyUnifiedTemplateExtraction applies the unified template copying logic to extracted template files
func (t *TemplateService) applyUnifiedTemplateExtraction(extractedPath, targetPath string, isHere bool, aiAssistant string) error {
	// List all entries in the extracted directory
	entries, err := t.filesystem.ListDirectoryContents(extractedPath)
	if err != nil {
		return fmt.Errorf("failed to list extracted directory contents: %w", err)
	}

	// Check what we have in the extracted template
	hasUnifiedStructure := false
	hasAgentFolders := false
	
	for _, entry := range entries {
		if entry.IsDir() {
			switch entry.Name() {
			case "commands", "templates", "tools":
				hasUnifiedStructure = true
			case ".claude", ".codex", ".gemini":
				hasAgentFolders = true
			}
		}
	}

	// If template has agent folders, apply reorganization for mixed structures
	if hasAgentFolders {
		return t.extractFromMixedStructure(extractedPath, targetPath, aiAssistant, hasUnifiedStructure)
	}
	
	// If the template has unified structure only (future format), treat it like a cache
	if hasUnifiedStructure {
		return t.extractFromUnifiedStructure(extractedPath, targetPath, aiAssistant)
	}

	// Otherwise, fall back to direct extraction (legacy format)
	return t.filesystem.MergeDirectories(extractedPath, targetPath)
}

// extractFromUnifiedStructure extracts from a unified template structure (like our cache)
func (t *TemplateService) extractFromUnifiedStructure(unifiedPath, targetPath, aiAssistant string) error {
	// List directories to see what's available
	entries, err := t.filesystem.ListDirectoryContents(unifiedPath)
	if err != nil {
		return fmt.Errorf("failed to list unified structure contents: %w", err)
	}

	// Copy non-memory directories to agent's hidden folder
	agentHiddenFolder := "." + aiAssistant
	agentTargetPath := filepath.Join(targetPath, agentHiddenFolder)
	
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "memory" {
			sourcePath := filepath.Join(unifiedPath, entry.Name())
			targetDirPath := filepath.Join(agentTargetPath, entry.Name())
			
			if err := t.filesystem.MergeDirectories(sourcePath, targetDirPath); err != nil {
				return fmt.Errorf("failed to copy directory %s to agent folder: %w", entry.Name(), err)
			}
		}
	}

	// Copy memory to project root
	memorySourcePath := filepath.Join(unifiedPath, "memory")
	memoryTargetPath := filepath.Join(targetPath, "memory")
	
	if exists, err := t.filesystem.DirectoryExists(memorySourcePath); err != nil {
		return fmt.Errorf("failed to check memory directory: %w", err)
	} else if exists {
		if err := t.filesystem.MergeDirectories(memorySourcePath, memoryTargetPath); err != nil {
			return fmt.Errorf("failed to copy memory directory: %w", err)
		}
	}

	return nil
}

// extractFromMixedStructure handles templates with both agent folders and unified structure
func (t *TemplateService) extractFromMixedStructure(extractedPath, targetPath, aiAssistant string, hasUnifiedStructure bool) error {
	// List all entries in the extracted directory
	entries, err := t.filesystem.ListDirectoryContents(extractedPath)
	if err != nil {
		return fmt.Errorf("failed to list mixed structure contents: %w", err)
	}

	agentHiddenFolder := "." + aiAssistant
	agentFolderInTemplate := filepath.Join(extractedPath, agentHiddenFolder)

	// First, copy agent-specific folder if it exists and matches the requested agent
	if exists, err := t.filesystem.DirectoryExists(agentFolderInTemplate); err != nil {
		return fmt.Errorf("failed to check agent folder existence: %w", err)
	} else if exists {
		// Process agent folder templates to target
		agentTargetPath := filepath.Join(targetPath, agentHiddenFolder)
		if err := t.processTemplateDirectory(agentFolderInTemplate, agentTargetPath, aiAssistant); err != nil {
			return fmt.Errorf("failed to process agent folder templates: %w", err)
		}
	}

	// If the template has unified structure files (like templates/), move them to agent folder
	if hasUnifiedStructure {
		agentTargetPath := filepath.Join(targetPath, agentHiddenFolder)
		
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") && entry.Name() != "memory" {
				// This is a unified structure directory (commands, templates, tools, etc.)
				sourcePath := filepath.Join(extractedPath, entry.Name())
				targetDirPath := filepath.Join(agentTargetPath, entry.Name())
				
				if err := t.processTemplateDirectory(sourcePath, targetDirPath, aiAssistant); err != nil {
					return fmt.Errorf("failed to process unified directory %s templates to agent folder: %w", entry.Name(), err)
				}
			}
		}
	}

	// Copy memory to project root
	memorySourcePath := filepath.Join(extractedPath, "memory")
	memoryTargetPath := filepath.Join(targetPath, "memory")
	
	if exists, err := t.filesystem.DirectoryExists(memorySourcePath); err != nil {
		return fmt.Errorf("failed to check memory directory: %w", err)
	} else if exists {
		if err := t.filesystem.MergeDirectories(memorySourcePath, memoryTargetPath); err != nil {
			return fmt.Errorf("failed to copy memory directory: %w", err)
		}
	}

	// Copy any other files that are not agent folders, unified structure, or memory
	for _, entry := range entries {
		if !entry.IsDir() || (entry.Name() != agentHiddenFolder && 
			entry.Name() != "memory" && 
			entry.Name() != "commands" && 
			entry.Name() != "templates" && 
			entry.Name() != "tools") {
			
			sourcePath := filepath.Join(extractedPath, entry.Name())
			targetPath := filepath.Join(targetPath, entry.Name())
			
			if entry.IsDir() {
				if err := t.filesystem.MergeDirectories(sourcePath, targetPath); err != nil {
					return fmt.Errorf("failed to copy directory %s: %w", entry.Name(), err)
				}
			} else {
				// Copy individual files
				if err := t.filesystem.CopyFile(sourcePath, targetPath); err != nil {
					return fmt.Errorf("failed to copy file %s: %w", entry.Name(), err)
				}
			}
		}
	}

	return nil
}

// GetAvailableTemplates returns information about available templates
func (t *TemplateService) GetAvailableTemplates() (map[string]*models.Template, error) {
	templates := make(map[string]*models.Template)

	for _, aiAssistant := range models.ListAgents() {
		template, err := t.github.GetTemplates()
		if err != nil {
			// Log error but continue with other templates
			continue
		}
		templates[aiAssistant] = template
	}

	if len(templates) == 0 {
		return nil, fmt.Errorf("%w: no templates available", models.ErrTemplateNotFound)
	}

	return templates, nil
}

// ValidateTemplate checks if a template is valid and available
func (t *TemplateService) ValidateTemplate(aiAssistant string) error {
	_, err := t.github.GetTemplates()
	return err
}

// DownloadTemplateWithProgress downloads a template with progress tracking
func (t *TemplateService) DownloadTemplateWithProgress(aiAssistant, targetPath string, isHere bool, progressCallback func(downloaded, total int64)) (*models.Template, error) {
	// Get template information
	template, err := t.github.GetTemplates()
	if err != nil {
		return nil, err
	}

	// Create temporary directory for download
	tempDir, err := t.filesystem.CreateTempDirectory("specify-download-")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create temporary directory: %v", models.ErrTemplateDownloadFailed, err)
	}
	defer t.filesystem.RemoveDirectory(tempDir)

	// Download with progress
	zipPath := filepath.Join(tempDir, template.FileName)
	if err := t.downloadTemplateWithProgress(template, zipPath, progressCallback); err != nil {
		return nil, err
	}

	// Extract template with AI assistant context
	if err := t.extractTemplate(zipPath, targetPath, isHere, aiAssistant); err != nil {
		return nil, err
	}

	return template, nil
}

// downloadTemplateWithProgress downloads the template with progress tracking
func (t *TemplateService) downloadTemplateWithProgress(template *models.Template, zipPath string, progressCallback func(downloaded, total int64)) error {
	// Create ZIP file
	file, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("%w: failed to create download file for version %s: %v", models.ErrTemplateDownloadFailed, template.Version, err)
	}
	defer file.Close()

	// Get release information again to get asset details
	release, err := t.github.GetLatestRelease()
	if err != nil {
		return err
	}

	// Find the asset
	asset, err := t.github.FindTemplateAsset(release)
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
	return t.github.GetTemplates()
}

// CheckTemplateUpdates checks if there are newer versions of templates available
func (t *TemplateService) CheckTemplateUpdates(currentVersions map[string]string) (map[string]*models.Template, error) {
	updates := make(map[string]*models.Template)

	for aiAssistant, currentVersion := range currentVersions {
		template, err := t.github.GetTemplates()
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
		return fmt.Errorf("%w: template directory does not exist after extraction", models.ErrTemplateExtractionFailed)
	}

	// Check if directory is not empty
	isEmpty, err := t.filesystem.IsDirectoryEmpty(path)
	if err != nil {
		return fmt.Errorf("failed to check if template directory is empty: %w", err)
	}

	if isEmpty {
		return fmt.Errorf("%w: template directory is empty after extraction", models.ErrTemplateExtractionFailed)
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

// ReadManifest reads and parses the cache manifest file from default cache location
func (t *TemplateService) ReadManifest() (*models.CacheManifest, error) {
	cacheRoot, err := t.ResolveRoot()
	if err != nil {
		return nil, err
	}
	return t.readManifestFromPath(cacheRoot)
}

// readManifestFromPath reads and parses the cache manifest file from specified path
func (t *TemplateService) readManifestFromPath(cacheRoot string) (*models.CacheManifest, error) {
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
	// This should be injected from build info using ldflags during build
	// For development, return "dev" and use flexible version checking
	return "dev"
}

// isVersionCompatible checks if cache version is compatible with current version
func (t *TemplateService) isVersionCompatible(cacheVersion, currentVersion string) bool {
	// If current version is "dev", accept any cache version
	if currentVersion == "dev" {
		return true
	}
	
	// If cache version is "dev", accept any current version
	if cacheVersion == "dev" {
		return true
	}
	
	// For release versions, require exact match
	return cacheVersion == currentVersion
}

// extractFromCache extracts templates from cache to target path
// New optimized approach: copy cache directories directly to hidden folders, memory to project
func (t *TemplateService) extractFromCache(aiAssistant, targetPath string, isHere bool) error {
	// Read and validate cache
	manifest, err := t.ReadManifest()
	if err != nil {
		return fmt.Errorf("failed to read cache manifest: %w", err)
	}

	// Check version match (flexible for development)
	currentVersion := t.GetSpecKitVersion()
	if !t.isVersionCompatible(manifest.SpecKitVersion, currentVersion) {
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

	// OPTIMIZED: Direct cache directory copying to hidden folders
	if err := t.CopyHiddenFolders(cacheRoot, targetPath, aiAssistant); err != nil {
		return fmt.Errorf("failed to copy hidden folders: %w", err)
	}

	// OPTIMIZED: Copy only memory folder to project directory
	if err := t.CopyMemoryToProject(cacheRoot, targetPath, aiAssistant); err != nil {
		return fmt.Errorf("failed to copy memory folder: %w", err)
	}

	return nil
}

// CopyHiddenFolders copies cache directories to agent's hidden folder
// Unified logic works for all agents (claude, codex, gemini, etc.)
func (t *TemplateService) CopyHiddenFolders(cacheRoot, targetPath, aiAssistant string) error {
	return t.copyAgentDirectories(cacheRoot, targetPath, aiAssistant)
}

// copyAgentDirectories is the unified logic for copying cache content to any agent's hidden folder
func (t *TemplateService) copyAgentDirectories(cacheRoot, targetPath, aiAssistant string) error {
	agentHiddenFolder := "." + aiAssistant
	agentTargetPath := filepath.Join(targetPath, agentHiddenFolder)

	// Read manifest from the specified cache root, fallback to directory scanning
	manifest, err := t.readManifestFromPath(cacheRoot)
	if err != nil {
		return t.copyAgentDirectoriesWithScan(cacheRoot, agentTargetPath, aiAssistant)
	}

	// Copy agent-specific directories from cache to agent's hidden folder
	copiedDirs := make(map[string]bool)
	
	for relativePath := range manifest.Templates {
		dirPath := filepath.Dir(relativePath)
		
		// Skip if already processed, current directory, memory, or content (handled separately)
		if dirPath == "." || dirPath == "memory" || dirPath == "content" || copiedDirs[dirPath] {
			continue
		}
		
		// Copy agent-specific directories: templates/, commands/, tools/, etc.
		// But NOT memory/ (that goes to project root via CopyMemoryToProject)
		sourcePath := filepath.Join(cacheRoot, dirPath)
		targetDirPath := filepath.Join(agentTargetPath, dirPath)
		
		if exists, err := t.filesystem.DirectoryExists(sourcePath); err != nil {
			return fmt.Errorf("failed to check directory %s: %w", dirPath, err)
		} else if exists {
			if err := t.processTemplateDirectory(sourcePath, targetDirPath, aiAssistant); err != nil {
				return fmt.Errorf("failed to process template directory %s: %w", dirPath, err)
			}
			copiedDirs[dirPath] = true
		}
	}

	return nil
}

// CopyHiddenFoldersWithScan is the public fallback method
func (t *TemplateService) CopyHiddenFoldersWithScan(cacheRoot, targetPath, aiAssistant string) error {
	agentHiddenFolder := "." + aiAssistant
	agentTargetPath := filepath.Join(targetPath, agentHiddenFolder)
	return t.copyAgentDirectoriesWithScan(cacheRoot, agentTargetPath, aiAssistant)
}

// copyAgentDirectoriesWithScan is the unified fallback when manifest is not available
func (t *TemplateService) copyAgentDirectoriesWithScan(cacheRoot, agentTargetPath, aiAssistant string) error {
	// Get list of all directories in cache root
	entries, err := t.filesystem.ListDirectoryContents(cacheRoot)
	if err != nil {
		return fmt.Errorf("failed to list cache directories: %w", err)
	}

	// Copy agent-specific directories from cache to agent's hidden folder
	// This handles: commands/, tools/, etc. - but NOT memory/ or content/
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "memory" && entry.Name() != "content" {
			sourcePath := filepath.Join(cacheRoot, entry.Name())
			targetDirPath := filepath.Join(agentTargetPath, entry.Name())

			if err := t.processTemplateDirectory(sourcePath, targetDirPath, aiAssistant); err != nil {
				return fmt.Errorf("failed to process template directory %s: %w", entry.Name(), err)
			}
		}
	}


	return nil
}

// copyMemoryToProject copies only the memory folder from cache to project directory
// Uses CacheManifest for dynamic discovery and validation of memory content
func (t *TemplateService) CopyMemoryToProject(cacheRoot, targetPath, aiAssistant string) error {
	// Read manifest from specified cache root for validation of memory files
	manifest, err := t.readManifestFromPath(cacheRoot)
	if err != nil {
		// Fallback to direct copy if manifest is not available
		return t.CopyMemoryDirectly(cacheRoot, targetPath, aiAssistant)
	}

	// Check if manifest contains any memory files
	hasMemoryFiles := false
	for relativePath := range manifest.Templates {
		if strings.HasPrefix(relativePath, "memory/") {
			hasMemoryFiles = true
			break
		}
	}

	if !hasMemoryFiles {
		// No memory files in cache manifest, skip copying
		return nil
	}

	// Copy memory directory to project using validated manifest
	memorySourcePath := filepath.Join(cacheRoot, "memory")
	memoryTargetPath := filepath.Join(targetPath, "memory")

	// Verify memory directory exists
	if exists, err := t.filesystem.DirectoryExists(memorySourcePath); err != nil {
		return fmt.Errorf("failed to check memory directory: %w", err)
	} else if !exists {
		return fmt.Errorf("manifest indicates memory files exist but directory is missing")
	}

	// Process memory directory templates to project
	if err := t.processTemplateDirectory(memorySourcePath, memoryTargetPath, aiAssistant); err != nil {
		return fmt.Errorf("failed to process memory directory templates: %w", err)
	}

	return nil
}

// copyMemoryDirectly is a fallback method when manifest is not available
func (t *TemplateService) CopyMemoryDirectly(cacheRoot, targetPath, aiAssistant string) error {
	memorySourcePath := filepath.Join(cacheRoot, "memory")
	memoryTargetPath := filepath.Join(targetPath, "memory")

	// Check if memory directory exists in cache
	exists, err := t.filesystem.DirectoryExists(memorySourcePath)
	if err != nil {
		return fmt.Errorf("failed to check memory directory: %w", err)
	}

	if !exists {
		// Memory folder is optional, don't error if missing
		return nil
	}

	// Process memory directory templates to project
	if err := t.processTemplateDirectory(memorySourcePath, memoryTargetPath, aiAssistant); err != nil {
		return fmt.Errorf("failed to process memory directory templates: %w", err)
	}

	return nil
}
