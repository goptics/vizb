package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goptics/vizb/version"
	"github.com/stretchr/testify/suite"
)

type GenerateUISuite struct {
	suite.Suite
}

func (s *GenerateUISuite) TestGenerateUI() {
	t := s.T()

	testTemplate := `<!DOCTYPE html><html><head><title>Test</title></head><body><script>window.VIZB_VERSION = [[VIZB .Version VIZB]]; window.VIZB_DATA = [[VIZB .Data VIZB]];</script></body></html>`

	t.Run("Happy Path - Valid JSON", func(t *testing.T) {
		benchmarkJSON := []byte(`[{"name":"test","data":[]}]`)

		result, err := GenerateUI(benchmarkJSON, []string{"bar"}, false, false, testTemplate)

		require.NoError(t, err)
		require.NotEmpty(t, result, "Expected non-empty HTML output")
		assert.Contains(t, result, "window.VIZB_DATA")
		assert.Contains(t, result, "window.VIZB_VERSION")
		assert.Contains(t, result, version.Version)
		assert.Contains(t, result, string(benchmarkJSON))
		assert.Contains(t, result, "<!DOCTYPE html>")
		assert.Contains(t, result, "<html")
	})

	t.Run("Empty JSON Array", func(t *testing.T) {
		benchmarkJSON := []byte(`[]`)

		result, err := GenerateUI(benchmarkJSON, []string{"bar"}, false, false, testTemplate)

		require.NoError(t, err)
		require.NotEmpty(t, result, "Expected non-empty HTML output")
		assert.Contains(t, result, "window.VIZB_DATA")
	})

	t.Run("Invalid Template Execution", func(t *testing.T) {
		benchmarkJSON := []byte(`[{"name":"test","data":[]}]`)
		invalidTemplate := `<!DOCTYPE html><html><body>[[VIZB .InvalidField VIZB]]</body></html>`

		_, err := GenerateUI(benchmarkJSON, []string{"bar"}, false, false, invalidTemplate)
		require.ErrorContains(t, err, "execute HTML template")
	})

	t.Run("Malformed Template Syntax", func(t *testing.T) {
		benchmarkJSON := []byte(`[{"name":"test","data":[]}]`)
		malformedTemplate := `<!DOCTYPE html><html><body>[[VIZB .Version</body></html>`

		_, err := GenerateUI(benchmarkJSON, []string{"bar"}, false, false, malformedTemplate)
		require.ErrorContains(t, err, "parse HTML template")
	})
}

func (s *GenerateUISuite) TestGenerateRemoteUI() {
	t := s.T()
	testTemplate := `<!DOCTYPE html><html><body>` +
		`<script>window.VIZB_DATA = [[VIZB .Data VIZB]];` +
		`window.VIZB_DATA_URL = [[VIZB .DataURL VIZB]];` +
		`window.VIZB_CHARTS = [[VIZB .ChartList VIZB]];</script></body></html>`

	result, err := GenerateRemoteUI("https://example.com/data.json", []string{"bar"}, false, false, testTemplate)

	require.NoError(t, err)
	require.NotEmpty(t, result)
	assert.Contains(t, result, "https://example.com/data.json")
	assert.Contains(t, result, "null")
	assert.Contains(t, result, `"bar"`)
}

func (s *GenerateUISuite) TestGeneratorsReturnErrors() {
	invalidTemplate := `[[VIZB .Version`
	_, err := GenerateUI([]byte(`[]`), []string{"bar"}, false, false, invalidTemplate)
	s.ErrorContains(err, "parse HTML template")

	_, err = GenerateRemoteUI("https://example.com/data.json", []string{"bar"}, false, false, invalidTemplate)
	s.ErrorContains(err, "parse HTML template")
}

func (s *GenerateUISuite) TestChartListJSEmptyDefaults() {
	t := s.T()
	testTemplate := `<!DOCTYPE html><html><body><script>window.VIZB_CHARTS = [[VIZB .ChartList VIZB]];</script></body></html>`
	result, err := GenerateUI([]byte(`[]`), nil, false, false, testTemplate)

	require.NoError(t, err)
	assert.Contains(t, result, `"bar"`)
	assert.Contains(t, result, `"line"`)
	assert.Contains(t, result, `"pie"`)
}

func TestGenerateUISuite(t *testing.T) {
	suite.Run(t, new(GenerateUISuite))
}
