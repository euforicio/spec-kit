package models

import "fmt"

// ProjectError represents errors related to project operations
type ProjectError struct {
	Type    ProjectErrorType `json:"type"`
	Path    string           `json:"path"`
	Message string           `json:"message"`
	Cause   error            `json:"cause,omitempty"`
}

// ProjectErrorType represents the type of project error
type ProjectErrorType int

const (
	ProjectAlreadyExists ProjectErrorType = iota
	ProjectPathInvalid
	ProjectNameInvalid
	ProjectAccessDenied
)

// String returns a string representation of the project error type
func (pet ProjectErrorType) String() string {
	switch pet {
	case ProjectAlreadyExists:
		return "already_exists"
	case ProjectPathInvalid:
		return "path_invalid"
	case ProjectNameInvalid:
		return "name_invalid"
	case ProjectAccessDenied:
		return "access_denied"
	default:
		return "unknown"
	}
}

// Error implements the error interface for ProjectError
func (pe *ProjectError) Error() string {
	if pe.Cause != nil {
		return fmt.Sprintf("project error (%s): %s - caused by: %v", pe.Type.String(), pe.Message, pe.Cause)
	}
	return fmt.Sprintf("project error (%s): %s", pe.Type.String(), pe.Message)
}

// Unwrap returns the underlying cause of the error
func (pe *ProjectError) Unwrap() error {
	return pe.Cause
}

// TemplateError represents errors related to template operations
type TemplateError struct {
	Type      TemplateErrorType `json:"type"`
	Assistant string            `json:"assistant"`
	Version   string            `json:"version,omitempty"`
	Message   string            `json:"message"`
	Cause     error             `json:"cause,omitempty"`
}

// TemplateErrorType represents the type of template error
type TemplateErrorType int

const (
	TemplateNotFound TemplateErrorType = iota
	TemplateDownloadFailed
	TemplateExtractionFailed
	TemplateCorrupted
)

// String returns a string representation of the template error type
func (tet TemplateErrorType) String() string {
	switch tet {
	case TemplateNotFound:
		return "not_found"
	case TemplateDownloadFailed:
		return "download_failed"
	case TemplateExtractionFailed:
		return "extraction_failed"
	case TemplateCorrupted:
		return "corrupted"
	default:
		return "unknown"
	}
}

// Error implements the error interface for TemplateError
func (te *TemplateError) Error() string {
	base := fmt.Sprintf("template error (%s) for assistant '%s': %s", te.Type.String(), te.Assistant, te.Message)
	if te.Version != "" {
		base = fmt.Sprintf("template error (%s) for assistant '%s' version '%s': %s", 
			te.Type.String(), te.Assistant, te.Version, te.Message)
	}
	if te.Cause != nil {
		base += fmt.Sprintf(" - caused by: %v", te.Cause)
	}
	return base
}

// Unwrap returns the underlying cause of the error
func (te *TemplateError) Unwrap() error {
	return te.Cause
}

// EnvironmentError represents errors related to environment operations
type EnvironmentError struct {
	Type    EnvironmentErrorType `json:"type"`
	Tool    string               `json:"tool,omitempty"`
	Message string               `json:"message"`
	Hint    string               `json:"hint,omitempty"`
}

// EnvironmentErrorType represents the type of environment error
type EnvironmentErrorType int

const (
	ToolNotFound EnvironmentErrorType = iota
	ToolVersionUnsupported
	InternetNotAvailable
	GitConfigMissing
)

// String returns a string representation of the environment error type
func (eet EnvironmentErrorType) String() string {
	switch eet {
	case ToolNotFound:
		return "tool_not_found"
	case ToolVersionUnsupported:
		return "tool_version_unsupported"
	case InternetNotAvailable:
		return "internet_not_available"
	case GitConfigMissing:
		return "git_config_missing"
	default:
		return "unknown"
	}
}

// Error implements the error interface for EnvironmentError
func (ee *EnvironmentError) Error() string {
	base := fmt.Sprintf("environment error (%s): %s", ee.Type.String(), ee.Message)
	if ee.Tool != "" {
		base = fmt.Sprintf("environment error (%s) for tool '%s': %s", ee.Type.String(), ee.Tool, ee.Message)
	}
	if ee.Hint != "" {
		base += fmt.Sprintf(" (hint: %s)", ee.Hint)
	}
	return base
}

// IsProjectError checks if an error is a ProjectError
func IsProjectError(err error) bool {
	_, ok := err.(*ProjectError)
	return ok
}

// IsTemplateError checks if an error is a TemplateError
func IsTemplateError(err error) bool {
	_, ok := err.(*TemplateError)
	return ok
}

// IsEnvironmentError checks if an error is an EnvironmentError
func IsEnvironmentError(err error) bool {
	_, ok := err.(*EnvironmentError)
	return ok
}

// AsProjectError returns the error as a ProjectError if possible
func AsProjectError(err error) (*ProjectError, bool) {
	pe, ok := err.(*ProjectError)
	return pe, ok
}

// AsTemplateError returns the error as a TemplateError if possible
func AsTemplateError(err error) (*TemplateError, bool) {
	te, ok := err.(*TemplateError)
	return te, ok
}

// AsEnvironmentError returns the error as an EnvironmentError if possible
func AsEnvironmentError(err error) (*EnvironmentError, bool) {
	ee, ok := err.(*EnvironmentError)
	return ee, ok
}