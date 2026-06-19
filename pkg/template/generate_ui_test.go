package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/version"
)

func TestGenerateUI(t *testing.T) {
	originalOsExit := shared.OsExit
	defer func() { shared.OsExit = originalOsExit }()

	testTemplate := `<!DOCTYPE html><html><head><title>Test</title></head><body><script>window.VIZB_VERSION = [[VIZB .Version VIZB]]; window.VIZB_DATA = [[VIZB .Data VIZB]];</script></body></html>`

	t.Run("Happy Path - Valid JSON", func(t *testing.T) {
		exitCalled := false
		shared.OsExit = func(code int) {
			exitCalled = true
		}

		benchmarkJSON := []byte(`[{"name":"test","data":[]}]`)

		result := GenerateUI(benchmarkJSON, []string{"bar"}, false, false, testTemplate)

		assert.False(t, exitCalled, "Expected OsExit not to be called for valid input")
		require.NotEmpty(t, result, "Expected non-empty HTML output")
		assert.Contains(t, result, "window.VIZB_DATA")
		assert.Contains(t, result, "window.VIZB_VERSION")
		assert.Contains(t, result, version.Version)
		assert.Contains(t, result, string(benchmarkJSON))
		assert.Contains(t, result, "<!DOCTYPE html>")
		assert.Contains(t, result, "<html")
	})

	t.Run("Empty JSON Array", func(t *testing.T) {
		exitCalled := false
		shared.OsExit = func(code int) {
			exitCalled = true
		}

		benchmarkJSON := []byte(`[]`)

		result := GenerateUI(benchmarkJSON, []string{"bar"}, false, false, testTemplate)

		assert.False(t, exitCalled, "Expected OsExit not to be called for empty JSON array")
		require.NotEmpty(t, result, "Expected non-empty HTML output")
		assert.Contains(t, result, "window.VIZB_DATA")
	})

	t.Run("Invalid Template Execution", func(t *testing.T) {
		exitCalled := false
		shared.OsExit = func(code int) {
			exitCalled = true
			panic("exit called")
		}

		benchmarkJSON := []byte(`[{"name":"test","data":[]}]`)
		invalidTemplate := `<!DOCTYPE html><html><body>[[VIZB .InvalidField VIZB]]</body></html>`

		err := shared.WithSafe("GenerateUI", func() {
			_ = GenerateUI(benchmarkJSON, []string{"bar"}, false, false, invalidTemplate)
		})

		assert.True(t, exitCalled, "Expected OsExit to be called for invalid template execution")
		assert.NotNil(t, err, "Expected error from WithSafe when OsExit is called")
	})

	t.Run("Malformed Template Syntax", func(t *testing.T) {
		exitCalled := false
		shared.OsExit = func(code int) {
			exitCalled = true
			panic("exit called")
		}

		benchmarkJSON := []byte(`[{"name":"test","data":[]}]`)
		malformedTemplate := `<!DOCTYPE html><html><body>[[VIZB .Version</body></html>`

		err := shared.WithSafe("GenerateUI", func() {
			_ = GenerateUI(benchmarkJSON, []string{"bar"}, false, false, malformedTemplate)
		})

		assert.True(t, exitCalled, "Expected OsExit to be called for malformed template syntax")
		assert.NotNil(t, err, "Expected error from WithSafe when OsExit is called")
	})
}

func TestGenerateRemoteUI(t *testing.T) {
	testTemplate := `<!DOCTYPE html><html><body>` +
		`<script>window.VIZB_DATA = [[VIZB .Data VIZB]];` +
		`window.VIZB_DATA_URL = [[VIZB .DataURL VIZB]];` +
		`window.VIZB_CHARTS = [[VIZB .ChartList VIZB]];</script></body></html>`

	result := GenerateRemoteUI("https://example.com/data.json", []string{"bar"}, false, false, testTemplate)

	require.NotEmpty(t, result)
	assert.Contains(t, result, "https://example.com/data.json")
	assert.Contains(t, result, "null")
	assert.Contains(t, result, `"bar"`)
}

func TestChartListJSEmptyDefaults(t *testing.T) {
	testTemplate := `<!DOCTYPE html><html><body><script>window.VIZB_CHARTS = [[VIZB .ChartList VIZB]];</script></body></html>`
	result := GenerateUI([]byte(`[]`), nil, false, false, testTemplate)

	assert.Contains(t, result, `"bar"`)
	assert.Contains(t, result, `"line"`)
	assert.Contains(t, result, `"pie"`)
}
