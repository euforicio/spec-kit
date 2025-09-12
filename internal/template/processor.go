package template

import (
	"bytes"
	"fmt"
	"sync"
	"text/template"
)

// Data represents the data available to templates
type Data struct {
	AIAssistant string
}

// ProcessorError represents errors from template processing
type ProcessorError struct {
	Type    ErrorType
	Message string
	Cause   error
}

func (e *ProcessorError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *ProcessorError) Unwrap() error {
	return e.Cause
}

// ErrorType represents different types of template processor errors
type ErrorType int

const (
	ParseError ErrorType = iota
	ExecuteError
)

// Processor handles template processing with caching and thread safety
type Processor struct {
	cache sync.Map // map[string]*template.Template
	mutex sync.RWMutex
}

// NewProcessor creates a new template processor instance
func NewProcessor() *Processor {
	return &Processor{}
}

// Process processes a template string with the given data
// Uses caching to avoid recompiling the same template multiple times
func (p *Processor) Process(content string, data Data) (string, error) {
	// Generate cache key from template content hash
	cacheKey := p.generateCacheKey(content)

	// Try to get cached template
	if cached, ok := p.cache.Load(cacheKey); ok {
		if tmpl, ok := cached.(*template.Template); ok {
			return p.executeTemplate(tmpl, data)
		}
	}

	// Parse and cache new template
	tmpl, err := template.New("template").Parse(content)
	if err != nil {
		return "", &ProcessorError{
			Type:    ParseError,
			Message: "failed to parse template",
			Cause:   err,
		}
	}

	// Store in cache
	p.cache.Store(cacheKey, tmpl)

	// Execute template
	return p.executeTemplate(tmpl, data)
}

// ProcessWithName processes a template string with a custom name for better error reporting
func (p *Processor) ProcessWithName(name, content string, data Data) (string, error) {
	// Generate cache key that includes the name
	cacheKey := p.generateNamedCacheKey(name, content)

	// Try to get cached template
	if cached, ok := p.cache.Load(cacheKey); ok {
		if tmpl, ok := cached.(*template.Template); ok {
			return p.executeTemplate(tmpl, data)
		}
	}

	// Parse and cache new template with custom name
	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return "", &ProcessorError{
			Type:    ParseError,
			Message: fmt.Sprintf("failed to parse template %q", name),
			Cause:   err,
		}
	}

	// Store in cache
	p.cache.Store(cacheKey, tmpl)

	// Execute template
	return p.executeTemplate(tmpl, data)
}

// executeTemplate executes a parsed template with thread safety
func (p *Processor) executeTemplate(tmpl *template.Template, data Data) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", &ProcessorError{
			Type:    ExecuteError,
			Message: "failed to execute template",
			Cause:   err,
		}
	}

	return buf.String(), nil
}

// generateCacheKey creates a cache key from template content
func (p *Processor) generateCacheKey(content string) string {
	// Simple hash-like key generation using content length and first/last chars
	// For production, consider using a proper hash function like SHA256
	if len(content) == 0 {
		return "empty"
	}

	// Create a simple but effective cache key
	first := content[0]
	last := content[len(content)-1]
	return fmt.Sprintf("tmpl_%d_%c_%c", len(content), first, last)
}

// generateNamedCacheKey creates a cache key from template name and content
func (p *Processor) generateNamedCacheKey(name, content string) string {
	return fmt.Sprintf("%s_%s", name, p.generateCacheKey(content))
}

// ClearCache clears all cached templates
func (p *Processor) ClearCache() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.cache = sync.Map{}
}

// CacheStats returns basic cache statistics
func (p *Processor) CacheStats() (count int) {
	p.cache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
