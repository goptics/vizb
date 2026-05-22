package javascript

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
)

func init() {
	parser.Parsers["js:vitest"] = ParseVitestBenchmark
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func parseNum(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.ParseFloat(s, 64)
	return n
}

func ParseVitestBenchmark(filename string) []shared.BenchmarkData {
	f, err := os.Open(filename)
	if err != nil {
		shared.ExitWithError("Error opening file", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var results []shared.BenchmarkData
	var currentSuite string

	for scanner.Scan() {
		line := scanner.Text()
		line = ansiRe.ReplaceAllString(line, "")
		line = strings.TrimLeftFunc(line, unicode.IsSpace)

		if idx := strings.Index(line, ">"); idx >= 0 && strings.Contains(line, "ms") {
			after := strings.TrimSpace(line[idx+1:])
			parts := strings.Fields(after)
			if len(parts) >= 2 {
				currentSuite = strings.Join(parts[:len(parts)-1], " ")
			}
			continue
		}

		if !strings.HasPrefix(line, "·") {
			continue
		}

		line = strings.TrimPrefix(line, "·")
		line = strings.TrimSpace(line)

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		name := fields[0]
		if currentSuite != "" {
			name = currentSuite + "/" + name
		}

		if !parser.ShouldIncludeBenchmark(name) {
			continue
		}

		hz := parseNum(fields[1])
		minVal := parseNum(fields[2])
		maxVal := parseNum(fields[3])
		mean := parseNum(fields[4])
		p75 := parseNum(fields[5])
		p99 := parseNum(fields[6])
		p995 := parseNum(fields[7])
		p999 := parseNum(fields[8])
		rmeStr := strings.TrimPrefix(fields[9], "±")
		rmeStr = strings.TrimSuffix(rmeStr, "%")
		rme := parseNum(rmeStr)
		samples := parseNum(fields[10])

		group, groupErr := parser.GroupBenchmarkName(name)
		if groupErr != nil {
			shared.ExitWithError("Error parsing vitest benchmark name", groupErr)
		}

		benchName, xAxis, yAxis := group["name"], group["xAxis"], group["yAxis"]

		results = append(results, shared.BenchmarkData{
			Name:  benchName,
			XAxis: xAxis,
			YAxis: yAxis,
			Stats: []shared.Stat{
				{Type: "Throughput avg (ops/s)", Value: hz},
				{Type: "Latency min (ms)", Value: minVal},
				{Type: "Latency max (ms)", Value: maxVal},
				{Type: "Latency avg (ms)", Value: mean},
				{Type: "Latency p75 (ms)", Value: p75},
				{Type: "Latency p99 (ms)", Value: p99},
				{Type: "Latency p995 (ms)", Value: p995},
				{Type: "Latency p999 (ms)", Value: p999},
				{Type: "RME (%)", Value: rme, Symbol: "±"},
				{Type: "Samples", Value: samples},
			},
		})
	}

	if err := scanner.Err(); err != nil {
		shared.ExitWithError("failed to read file", err)
	}

	return results
}
