package template

import (
	"strings"
	"testing"
)

func TestProcessor(t *testing.T) {
	f := func(name string, content string, data Data, expected string, expectError bool) {
		t.Helper()
		processor := NewProcessor()

		result, err := processor.Process(content, data)

		if expectError {
			if err == nil {
				t.Fatalf("%s: expected error but got none", name)
			}
			return
		}

		if err != nil {
			t.Fatalf("%s: unexpected error: %v", name, err)
		}

		if result != expected {
			t.Fatalf("%s: got %q, expected %q", name, result, expected)
		}
	}

	// Test basic template processing
	data := Data{AIAssistant: "claude"}
	f("simple template", "Hello {{.AIAssistant}}", data, "Hello claude", false)

	// Test template with no variables
	f("static template", "Hello World", data, "Hello World", false)

	// Test invalid template
	f("invalid template", "Hello {{.Invalid", data, "", true)

	// Test empty template
	f("empty template", "", data, "", false)
}

func TestProcessorWithName(t *testing.T) {
	f := func(testName, templateName, content string, data Data, expected string, expectError bool) {
		t.Helper()
		processor := NewProcessor()

		result, err := processor.ProcessWithName(templateName, content, data)

		if expectError {
			if err == nil {
				t.Fatalf("%s: expected error but got none", testName)
			}
			// Check if error contains template name
			if !strings.Contains(err.Error(), templateName) {
				t.Fatalf("%s: error should contain template name %q: %v", testName, templateName, err)
			}
			return
		}

		if err != nil {
			t.Fatalf("%s: unexpected error: %v", testName, err)
		}

		if result != expected {
			t.Fatalf("%s: got %q, expected %q", testName, result, expected)
		}
	}

	data := Data{AIAssistant: "codex"}
	f("named template", "test.md", "Agent: {{.AIAssistant}}", data, "Agent: codex", false)
	f("named invalid template", "bad.md", "Agent: {{.Invalid", data, "", true)
}

func TestProcessorCaching(t *testing.T) {
	processor := NewProcessor()
	data := Data{AIAssistant: "gemini"}
	template := "Hello {{.AIAssistant}}"

	// Initial cache should be empty
	if count := processor.CacheStats(); count != 0 {
		t.Fatalf("expected empty cache, got %d entries", count)
	}

	// First processing should add to cache
	result1, err := processor.Process(template, data)
	if err != nil {
		t.Fatalf("first processing failed: %v", err)
	}

	if count := processor.CacheStats(); count != 1 {
		t.Fatalf("expected 1 cache entry after first processing, got %d", count)
	}

	// Second processing should use cache
	result2, err := processor.Process(template, data)
	if err != nil {
		t.Fatalf("second processing failed: %v", err)
	}

	if count := processor.CacheStats(); count != 1 {
		t.Fatalf("expected 1 cache entry after second processing, got %d", count)
	}

	if result1 != result2 {
		t.Fatalf("cached result differs: %q vs %q", result1, result2)
	}

	expected := "Hello gemini"
	if result1 != expected {
		t.Fatalf("got %q, expected %q", result1, expected)
	}
}

func TestProcessorClearCache(t *testing.T) {
	processor := NewProcessor()
	data := Data{AIAssistant: "claude"}

	// Add something to cache
	_, err := processor.Process("test {{.AIAssistant}}", data)
	if err != nil {
		t.Fatalf("processing failed: %v", err)
	}

	if count := processor.CacheStats(); count != 1 {
		t.Fatalf("expected 1 cache entry, got %d", count)
	}

	// Clear cache
	processor.ClearCache()

	if count := processor.CacheStats(); count != 0 {
		t.Fatalf("expected empty cache after clear, got %d entries", count)
	}
}

func TestProcessorError(t *testing.T) {
	processor := NewProcessor()
	data := Data{AIAssistant: "test"}

	// Test parse error
	_, err := processor.Process("{{.Invalid", data)
	if err == nil {
		t.Fatal("expected parse error")
	}

	if processorErr, ok := err.(*ProcessorError); ok {
		if processorErr.Type != ParseError {
			t.Fatalf("expected ParseError, got %v", processorErr.Type)
		}
		if processorErr.Unwrap() == nil {
			t.Fatal("expected wrapped error")
		}
	} else {
		t.Fatalf("expected ProcessorError, got %T", err)
	}

	// Test execute error
	_, err = processor.Process("{{.NonExistentField.InvalidAccess}}", data)
	if err == nil {
		t.Fatal("expected execute error")
	}

	if processorErr, ok := err.(*ProcessorError); ok {
		if processorErr.Type != ExecuteError {
			t.Fatalf("expected ExecuteError, got %v", processorErr.Type)
		}
	} else {
		t.Fatalf("expected ProcessorError, got %T", err)
	}
}
