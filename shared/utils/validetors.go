package utils

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"

	"github.com/goptics/vizb/shared"
)

func IsBenchJSONFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var ev shared.BenchEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			return false
		}

		// very Go-specific validation rules
		if ev.Action == "" {
			return false
		}

		// if it's a benchmark, "Test" usually begins with "Benchmark"
		if ev.Test != "" && strings.HasPrefix(ev.Test, "Benchmark") {
			return true
		}

		return true
	}

	return false
}
