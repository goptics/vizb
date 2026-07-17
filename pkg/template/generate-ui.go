package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	htmlTemplate "html/template"

	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/version"
)

type PageData struct {
	Version   string
	Data      htmlTemplate.JS
	DataURL   string
	Chunks    htmlTemplate.JS
	ChartList htmlTemplate.JS
}

// chartListJS marshals the bundled chart selection for window.VIZB_CHARTS so the
// UI only surfaces tabs whose renderer chunks were actually shipped.
func chartListJS(charts []string) htmlTemplate.JS {
	if len(charts) == 0 {
		charts = defaultCharts
	}
	encoded, err := json.Marshal(charts)
	if err != nil {
		return htmlTemplate.JS("[]")
	}
	return htmlTemplate.JS(encoded)
}

func renderPage(pd PageData, HTMLtemplate string) (string, error) {
	// Custom delimiters avoid clashing with echarts-gl's clay.gl GLSL shaders,
	// which embed literal {{ }} sequences for shader loop unrolling.
	tmpl, err := htmlTemplate.New("page").Delims("[[VIZB", "VIZB]]").Parse(HTMLtemplate)
	if err != nil {
		return "", fmt.Errorf("parse HTML template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, pd); err != nil {
		return "", fmt.Errorf("execute HTML template: %w", err)
	}

	return buf.String(), nil
}

// GenerateUI renders the embedded-data page, shipping only the chunks the
// selected charts (+ needs3D + needsHeatmapChunk) can reach.
func GenerateUI(benchmarkJSON []byte, charts []string, needs3D bool, needsHeatmapChunk bool, HTMLtemplate string) string {
	html, err := GenerateUIE(benchmarkJSON, charts, needs3D, needsHeatmapChunk, HTMLtemplate)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}
	return html
}

// GenerateUIE renders embedded data and returns template failures to the
// caller. It is safe for request handlers and never writes or exits.
func GenerateUIE(benchmarkJSON []byte, charts []string, needs3D bool, needsHeatmapChunk bool, HTMLtemplate string) (string, error) {
	return renderPage(PageData{
		Version:   version.Version,
		Data:      htmlTemplate.JS(benchmarkJSON),
		Chunks:    SelectChunks(charts, needs3D, needsHeatmapChunk),
		ChartList: chartListJS(charts),
	}, HTMLtemplate)
}

// GenerateRemoteUI renders the runtime-fetch page. Data is unknown at generation
// time, so chunk pruning follows the --charts selection directly.
func GenerateRemoteUI(dataURL string, charts []string, needs3D bool, needsHeatmapChunk bool, HTMLtemplate string) string {
	html, err := GenerateRemoteUIE(dataURL, charts, needs3D, needsHeatmapChunk, HTMLtemplate)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}
	return html
}

// GenerateRemoteUIE is the error-returning remote-data variant.
func GenerateRemoteUIE(dataURL string, charts []string, needs3D bool, needsHeatmapChunk bool, HTMLtemplate string) (string, error) {
	return renderPage(PageData{
		Version:   version.Version,
		Data:      htmlTemplate.JS("null"),
		DataURL:   dataURL,
		Chunks:    SelectChunks(charts, needs3D, needsHeatmapChunk),
		ChartList: chartListJS(charts),
	}, HTMLtemplate)
}
