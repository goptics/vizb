package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/goptics/vizb/pkg/style"
	"github.com/goptics/vizb/shared"
)

// ResolveOutputFileName decides the final output file name. An empty name yields
// a temp HTML file (printed to stdout later); a name without an extension gets
// the default .html.
func ResolveOutputFileName(outFile string) string {
	if outFile == "" {
		tmpFilePath := shared.MustCreateTempFile(shared.TempBenchFilePrefix, "html")
		shared.TempFiles.Store(tmpFilePath)
		return tmpFilePath
	}

	if filepath.Ext(outFile) == "" {
		outFile += ".html"
	}

	return outFile
}

// InferFormatFromExtension returns "json" for .json output, otherwise "html".
func InferFormatFromExtension(outFile string) string {
	switch ext := strings.ToLower(filepath.Ext(outFile)); ext {
	case ".json":
		return "json"
	default:
		return "html"
	}
}

// HandleOutputResult prints the output path when the user named one, otherwise
// dumps the (temp) file's contents to stdout. userOutput is the raw -o value.
func HandleOutputResult(f *os.File, userOutput string) {
	if userOutput != "" {
		fmt.Println(style.Info.Render(fmt.Sprintf("📄 Output file: %s", f.Name())))
		return
	}

	content, err := os.ReadFile(f.Name())
	if err != nil {
		shared.ExitWithError("Error reading output file", err)
	}
	fmt.Print("\033[H\033[2J") // clear screen
	fmt.Println(string(content))
}

// convertToDataset tries to read filePath as an existing vizb Dataset JSON
// (single object). Returns nil when the content is not Dataset JSON.
func convertToDataset(filePath string) (dataSet *shared.Dataset) {
	f := shared.MustOpenFile(filePath)
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		shared.ExitWithError("Failed to read file: %v", err)
	}

	if err := json.Unmarshal(content, &dataSet); err != nil {
		return nil
	}
	shared.MigrateDataset(dataSet, content)
	return dataSet
}

// ParseDatasetFile reads a vizb Dataset JSON file (single object or array) and
// returns the datasets, applying schema migration. Shared by ui and merge.
func ParseDatasetFile(file string) ([]shared.Dataset, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	trimmed := bytes.TrimLeft(content, " \t\r\n")
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	switch trimmed[0] {
	case '[':
		// Two-pass: decode each element from its own raw bytes so MigrateDataset
		// can recover the legacy top-level axisLabels field (lost after Unmarshal).
		var rawElems []json.RawMessage
		if err := json.Unmarshal(content, &rawElems); err != nil {
			return nil, fmt.Errorf("invalid data set array: %w", err)
		}
		dataSets := make([]shared.Dataset, 0, len(rawElems))
		for _, rawElem := range rawElems {
			var ds shared.Dataset
			if err := json.Unmarshal(rawElem, &ds); err != nil {
				return nil, fmt.Errorf("invalid data set array: %w", err)
			}
			shared.MigrateDataset(&ds, rawElem)
			dataSets = append(dataSets, ds)
		}
		return dataSets, nil
	case '{':
		var ds shared.Dataset
		if err := json.Unmarshal(content, &ds); err != nil {
			return nil, fmt.Errorf("invalid data set object: %w", err)
		}
		shared.MigrateDataset(&ds, content)
		return []shared.Dataset{ds}, nil
	default:
		return nil, fmt.Errorf("not valid JSON")
	}
}
