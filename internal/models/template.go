package models

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Pre-compiled regex for hex validation to improve performance
var hexRegex = regexp.MustCompile(`^[a-fA-F0-9]+$`)

// Template represents a downloadable project template from GitHub releases.
type Template struct {
	AIAssistant string    `json:"ai_assistant"` // AI assistant type (claude, gemini, copilot)
	Version     string    `json:"version"`      // GitHub release version (tag_name)
	DownloadURL string    `json:"download_url"` // Direct download URL for ZIP file
	FileName    string    `json:"file_name"`    // ZIP file name
	Size        int64     `json:"size"`         // File size in bytes
	ReleaseDate time.Time `json:"release_date"` // When release was published
	AssetName   string    `json:"asset_name"`   // GitHub asset name pattern
}

// TemplateState represents the current state of template processing
type TemplateState int

const (
	TemplateStateRequested TemplateState = iota
	TemplateStateLocated
	TemplateStateDownloading
	TemplateStateDownloaded
	TemplateStateExtracted
	TemplateStateCleaned
)

// String returns a string representation of the template state
func (ts TemplateState) String() string {
	switch ts {
	case TemplateStateRequested:
		return "requested"
	case TemplateStateLocated:
		return "located"
	case TemplateStateDownloading:
		return "downloading"
	case TemplateStateDownloaded:
		return "downloaded"
	case TemplateStateExtracted:
		return "extracted"
	case TemplateStateCleaned:
		return "cleaned"
	default:
		return "unknown"
	}
}

// NewTemplate creates a new Template instance with validation
func NewTemplate(aiAssistant, version, downloadURL, fileName string, size int64, releaseDate time.Time, assetName string) (*Template, error) {
	template := &Template{
		AIAssistant: aiAssistant,
		Version:     version,
		DownloadURL: downloadURL,
		FileName:    fileName,
		Size:        size,
		ReleaseDate: releaseDate,
		AssetName:   assetName,
	}

	if err := template.Validate(); err != nil {
		return nil, err
	}

	return template, nil
}

// Validate checks if the template configuration is valid
func (t *Template) Validate() error {
	// Validate AI assistant
	if err := t.validateAIAssistant(); err != nil {
		return err
	}

	// Validate download URL
	if err := t.validateDownloadURL(); err != nil {
		return err
	}

	// Validate size
	if err := t.validateSize(); err != nil {
		return err
	}

	// Validate version
	if err := t.validateVersion(); err != nil {
		return err
	}

	return nil
}

// validateAIAssistant validates the AI assistant type
func (t *Template) validateAIAssistant() error {
	if t.AIAssistant == "" {
		return &TemplateError{
			Type:      TemplateNotFound,
			Assistant: t.AIAssistant,
			Message:   "AI assistant cannot be empty",
		}
	}

	// Check if AI assistant is valid
	if IsValidAgent(t.AIAssistant) {
		return nil
	}

	return &TemplateError{
		Type:      TemplateNotFound,
		Assistant: t.AIAssistant,
		Message: fmt.Sprintf("invalid AI assistant '%s', must be one of: %s",
			t.AIAssistant, strings.Join(ListAgents(), ", ")),
	}
}

// validateDownloadURL validates the download URL
func (t *Template) validateDownloadURL() error {
	if t.DownloadURL == "" {
		return &TemplateError{
			Type:      TemplateDownloadFailed,
			Assistant: t.AIAssistant,
			Message:   "download URL cannot be empty",
		}
	}

	// Parse URL to ensure it's valid
	parsedURL, err := url.Parse(t.DownloadURL)
	if err != nil {
		return &TemplateError{
			Type:      TemplateDownloadFailed,
			Assistant: t.AIAssistant,
			Message:   fmt.Sprintf("invalid download URL: %v", err),
			Cause:     err,
		}
	}

	// Ensure HTTPS
	if parsedURL.Scheme != "https" {
		return &TemplateError{
			Type:      TemplateDownloadFailed,
			Assistant: t.AIAssistant,
			Message:   "download URL must use HTTPS",
		}
	}

	// Ensure it's a GitHub URL
	if !strings.Contains(parsedURL.Host, "github") {
		return &TemplateError{
			Type:      TemplateDownloadFailed,
			Assistant: t.AIAssistant,
			Message:   "download URL must be from GitHub",
		}
	}

	return nil
}

// validateSize validates the file size
func (t *Template) validateSize() error {
	if t.Size <= 0 {
		return &TemplateError{
			Type:      TemplateCorrupted,
			Assistant: t.AIAssistant,
			Message:   "file size must be positive",
		}
	}

	// Check for reasonable size limits (e.g., 100MB max)
	const maxSize = 100 * 1024 * 1024 // 100MB
	if t.Size > maxSize {
		return &TemplateError{
			Type:      TemplateCorrupted,
			Assistant: t.AIAssistant,
			Message:   fmt.Sprintf("file size %d bytes exceeds maximum allowed size of %d bytes", t.Size, maxSize),
		}
	}

	return nil
}

// validateVersion validates the version string
func (t *Template) validateVersion() error {
	if t.Version == "" {
		return &TemplateError{
			Type:      TemplateNotFound,
			Assistant: t.AIAssistant,
			Message:   "version cannot be empty",
		}
	}

	// Basic semantic version check (should start with 'v' or be a number)
	if !strings.HasPrefix(t.Version, "v") && !isNumeric(t.Version[0:1]) {
		return &TemplateError{
			Type:      TemplateNotFound,
			Assistant: t.AIAssistant,
			Message:   fmt.Sprintf("invalid version format: %s", t.Version),
		}
	}

	return nil
}

// GetExpectedAssetName returns the expected asset name pattern for this AI assistant
func (t *Template) GetExpectedAssetName() string {
	return fmt.Sprintf("spec-kit-template-%s", t.AIAssistant)
}

// IsAssetNameMatch checks if the given asset name matches this template
func (t *Template) IsAssetNameMatch(assetName string) bool {
	expectedPattern := t.GetExpectedAssetName()
	return strings.Contains(assetName, expectedPattern) && strings.HasSuffix(assetName, ".zip")
}

// GetSizeFormatted returns the file size in a human-readable format
func (t *Template) GetSizeFormatted() string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	size := float64(t.Size)
	switch {
	case t.Size >= GB:
		return fmt.Sprintf("%.1f GB", size/GB)
	case t.Size >= MB:
		return fmt.Sprintf("%.1f MB", size/MB)
	case t.Size >= KB:
		return fmt.Sprintf("%.1f KB", size/KB)
	default:
		return fmt.Sprintf("%d bytes", t.Size)
	}
}

// GetDisplayInfo returns a formatted string with template information
func (t *Template) GetDisplayInfo() string {
	return fmt.Sprintf("%s (%s, %s)", t.FileName, t.Version, t.GetSizeFormatted())
}

// CacheManifest represents the structure of the template cache manifest
type CacheManifest struct {
	SpecKitVersion string            `json:"spec_kit_version"`
	LastSync       time.Time         `json:"last_sync"`
	Templates      map[string]string `json:"templates"` // filepath -> sha256 hash
}

// NewCacheManifest creates a new cache manifest with the given version
func NewCacheManifest(specKitVersion string) *CacheManifest {
	return &CacheManifest{
		SpecKitVersion: specKitVersion,
		LastSync:       time.Now().UTC(),
		Templates:      make(map[string]string),
	}
}

// AddTemplate adds a template file with its hash to the manifest
func (cm *CacheManifest) AddTemplate(relativePath, hash string) error {
	if relativePath == "" {
		return fmt.Errorf("relative path cannot be empty")
	}
	if len(hash) != 64 { // SHA256 is 64 hex chars
		return fmt.Errorf("invalid hash length: expected 64, got %d", len(hash))
	}
	if !hexRegex.MatchString(hash) {
		return fmt.Errorf("invalid hash format: must be hexadecimal")
	}
	cm.Templates[relativePath] = hash
	cm.LastSync = time.Now().UTC()
	return nil
}

// GetTemplateHash returns the hash for a given template path
func (cm *CacheManifest) GetTemplateHash(relativePath string) (string, bool) {
	hash, exists := cm.Templates[relativePath]
	return hash, exists
}

// IsVersionMatch checks if the manifest version matches the given version
func (cm *CacheManifest) IsVersionMatch(version string) bool {
	return cm.SpecKitVersion == version
}

// GetTemplateCount returns the number of templates in the manifest
func (cm *CacheManifest) GetTemplateCount() int {
	return len(cm.Templates)
}

// UpdateLastSync updates the last sync timestamp to current UTC time
func (cm *CacheManifest) UpdateLastSync() {
	cm.LastSync = time.Now().UTC()
}

// Helper function to check if a string starts with a number
func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	return s[0] >= '0' && s[0] <= '9'
}
