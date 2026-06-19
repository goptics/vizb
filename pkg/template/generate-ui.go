package template

import (
	"bytes"
	"encoding/json"
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

func renderPage(pd PageData, HTMLtemplate string) string {
	// Custom delimiters avoid clashing with echarts-gl's clay.gl GLSL shaders,
	// which embed literal {{ }} sequences for shader loop unrolling.
	tmpl, err := htmlTemplate.New("page").Delims("[[VIZB", "VIZB]]").Parse(HTMLtemplate)
	if err != nil {
		shared.ExitWithError("failed to parse HTML template:", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, pd); err != nil {
		shared.ExitWithError("failed to execute HTML template:", err)
	}

	return buf.String()
}

// GenerateUI renders the embedded-data page, shipping only the chunks the
// selected charts (+ needs3D + needsHeatmapChunk) can reach.
func GenerateUI(benchmarkJSON []byte, charts []string, needs3D bool, needsHeatmapChunk bool, HTMLtemplate string) string {
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
	return renderPage(PageData{
		Version:   version.Version,
		Data:      htmlTemplate.JS("null"),
		DataURL:   dataURL,
		Chunks:    SelectChunks(charts, needs3D, needsHeatmapChunk),
		ChartList: chartListJS(charts),
	}, HTMLtemplate)
}
