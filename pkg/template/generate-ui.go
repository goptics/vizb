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
}

func GenerateHTMLBenchmarkUI(benchmarkJSON []byte, HTMLtemplate string) string {
	tmpl, err := htmlTemplate.New("page").Parse(HTMLtemplate)

	if err != nil {
		shared.ExitWithError("failed to parse HTML template:", err)
	}

	pageData := PageData{
		Version: version.Version,
		Data:    htmlTemplate.JS(benchmarkJSON),
	}

	var buf bytes.Buffer

	// Render into the buffer instead of stdout
	if err := tmpl.Execute(&buf, pageData); err != nil {
		shared.ExitWithError("failed to execute HTML template:", err)
	}

	return buf.String()
}
