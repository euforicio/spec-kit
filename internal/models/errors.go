package models

import "errors"

// Sentinel errors for template operations
var (
	ErrTemplateNotFound         = errors.New("template not found")
	ErrTemplateDownloadFailed   = errors.New("template download failed")
	ErrTemplateExtractionFailed = errors.New("template extraction failed")
	ErrTemplateCorrupted        = errors.New("template corrupted")
	ErrTemplateCacheFailed      = errors.New("template cache failed")
)

// Sentinel errors for project operations
var (
	ErrProjectAlreadyExists = errors.New("project already exists")
	ErrProjectPathInvalid   = errors.New("project path invalid")
	ErrProjectNameInvalid   = errors.New("project name invalid")
	ErrProjectAccessDenied  = errors.New("project access denied")
)

// Sentinel errors for environment operations
var (
	ErrToolNotFound           = errors.New("tool not found")
	ErrToolVersionUnsupported = errors.New("tool version unsupported")
	ErrInternetNotAvailable   = errors.New("internet not available")
	ErrGitConfigMissing       = errors.New("git config missing")
)
