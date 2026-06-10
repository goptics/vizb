package template

import (
	"bytes"
	htmlTemplate "html/template"

	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/version"
)

type PageData struct {
	Version string
	Data    htmlTemplate.JS
	DataURL string
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

func GenerateUI(benchmarkJSON []byte, HTMLtemplate string) string {
	return renderPage(PageData{
		Version: version.Version,
		Data:    htmlTemplate.JS(benchmarkJSON),
	}, HTMLtemplate)
}

func GenerateRemoteUI(dataURL string, HTMLtemplate string) string {
	return renderPage(PageData{
		Version: version.Version,
		Data:    htmlTemplate.JS("null"),
		DataURL: dataURL,
	}, HTMLtemplate)
}
