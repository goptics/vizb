package template

import (
	"strings"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/version"
)

func TestGenerateHTMLBenchmarkUI(t *testing.T) {
	// Save original OsExit and restore after test
	originalOsExit := shared.OsExit
	defer func() { shared.OsExit = originalOsExit }()

	// Simple test template
	testTemplate := `<!DOCTYPE html><html><head><title>Test</title></head><body><script>window.VIZB_VERSION = {{ .Version }}; window.VIZB_DATA = {{ .Data }};</script></body></html>`

	t.Run("Happy Path - Valid JSON", func(t *testing.T) {
		exitCalled := false
		shared.OsExit = func(code int) {
			exitCalled = true
		}

		benchmarkJSON := []byte(`[{"name":"test","data":[]}]`)

		result := GenerateHTMLBenchmarkUI(benchmarkJSON, testTemplate)

		// Verify OsExit was not called
		if exitCalled {
			t.Error("Expected OsExit not to be called for valid input")
		}

		// Verify result is not empty
		if result == "" {
			t.Error("Expected non-empty HTML output")
		}

		// Verify window.VIZB_DATA is injected
		if !strings.Contains(result, "window.VIZB_DATA") {
			t.Error("Expected HTML to contain window.VIZB_DATA")
		}

		// Verify window.VIZB_VERSION is injected
		if !strings.Contains(result, "window.VIZB_VERSION") {
			t.Error("Expected HTML to contain window.VIZB_VERSION")
		}

		// Verify version value is present
		if !strings.Contains(result, version.Version) {
			t.Errorf("Expected HTML to contain version %s", version.Version)
		}

		// Verify the JSON data is present
		if !strings.Contains(result, string(benchmarkJSON)) {
			t.Error("Expected HTML to contain the benchmark JSON data")
		}

		// Verify basic HTML structure
		if !strings.Contains(result, "<!DOCTYPE html>") {
			t.Error("Expected HTML to contain DOCTYPE declaration")
		}

		if !strings.Contains(result, "<html") {
			t.Error("Expected HTML to contain html tag")
		}
	})

	t.Run("Empty JSON Array", func(t *testing.T) {
		exitCalled := false
		shared.OsExit = func(code int) {
			exitCalled = true
		}

		benchmarkJSON := []byte(`[]`)

		result := GenerateHTMLBenchmarkUI(benchmarkJSON, testTemplate)

		if exitCalled {
			t.Error("Expected OsExit not to be called for empty JSON array")
		}

		if result == "" {
			t.Error("Expected non-empty HTML output")
		}

		if !strings.Contains(result, "window.VIZB_DATA") {
			t.Error("Expected HTML to contain window.VIZB_DATA")
		}
	})

	t.Run("Invalid Template Execution", func(t *testing.T) {
		exitCalled := false
		shared.OsExit = func(code int) {
			exitCalled = true
			panic("exit called") // Panic to stop execution
		}

		benchmarkJSON := []byte(`[{"name":"test","data":[]}]`)
		invalidTemplate := `<!DOCTYPE html><html><body>{{ .InvalidField }}</body></html>`

		// Use WithSafe to handle the panic from our mocked OsExit
		err := shared.WithSafe("GenerateHTMLBenchmarkUI", func() {
			_ = GenerateHTMLBenchmarkUI(benchmarkJSON, invalidTemplate)
		})

		if !exitCalled {
			t.Error("Expected OsExit to be called for invalid template execution")
		}

		if err == nil {
			t.Error("Expected error from WithSafe when OsExit is called")
		}
	})

	t.Run("Malformed Template Syntax", func(t *testing.T) {
		exitCalled := false
		shared.OsExit = func(code int) {
			exitCalled = true
			panic("exit called") // Panic to stop execution
		}

		benchmarkJSON := []byte(`[{"name":"test","data":[]}]`)
		malformedTemplate := `<!DOCTYPE html><html><body>{{ .Version</body></html>`

		// Use WithSafe to handle the panic from our mocked OsExit
		err := shared.WithSafe("GenerateHTMLBenchmarkUI", func() {
			_ = GenerateHTMLBenchmarkUI(benchmarkJSON, malformedTemplate)
		})

		if !exitCalled {
			t.Error("Expected OsExit to be called for malformed template syntax")
		}

		if err == nil {
			t.Error("Expected error from WithSafe when OsExit is called")
		}
	})
}
