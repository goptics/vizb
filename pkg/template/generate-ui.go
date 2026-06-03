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
	tmpl, err := htmlTemplate.New("page").Parse(HTMLtemplate)
	if err != nil {
		shared.ExitWithError("failed to parse HTML template:", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, pd); err != nil {
		shared.ExitWithError("failed to execute HTML template:", err)
	}

	return buf.String()
}

func GenerateHTMLBenchmarkUI(benchmarkJSON []byte, HTMLtemplate string) string {
	return renderPage(PageData{
		Version: version.Version,
		Data:    htmlTemplate.JS(benchmarkJSON),
	}, HTMLtemplate)
}

func GenerateRemoteHTMLBenchmarkUI(dataURL string, HTMLtemplate string) string {
	return renderPage(PageData{
		Version: version.Version,
		Data:    htmlTemplate.JS("null"),
		DataURL: dataURL,
	}, HTMLtemplate)
}
