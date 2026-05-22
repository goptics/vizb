package rust

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
)

func init() {
	parser.Parsers["rs:cargo"] = ParseCargoBenchmark
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

var criterionRe = regexp.MustCompile(`^(\S+)\s+time:\s+\[([\d.]+)\s*(ns|µs|μs|ms|s)\s+([\d.]+)\s*(ns|µs|μs|ms|s)\s+([\d.]+)\s*(ns|µs|μs|ms|s)\]`)

func toNsMultiplier(unit string) float64 {
	switch unit {
	case "ns":
		return 1
	case "µs", "μs":
		return 1e3
	case "ms":
		return 1e6
	case "s":
		return 1e9
	}
	return 1
}

func parseNum(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.ParseFloat(s, 64)
	return n
}

func ParseCargoBenchmark(filename string) []shared.BenchmarkData {
	f, err := os.Open(filename)
	if err != nil {
		shared.ExitWithError("Error opening file", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var results []shared.BenchmarkData

	for scanner.Scan() {
		line := scanner.Text()
		line = ansiRe.ReplaceAllString(line, "")
		line = strings.TrimLeftFunc(line, unicode.IsSpace)

		match := criterionRe.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		name := match[1]
		if !parser.ShouldIncludeBenchmark(name) {
			continue
		}

		lowerNs := parseNum(match[2]) * toNsMultiplier(match[3])
		estimateNs := parseNum(match[4]) * toNsMultiplier(match[5])
		upperNs := parseNum(match[6]) * toNsMultiplier(match[7])

		group, groupErr := parser.GroupBenchmarkName(name)
		if groupErr != nil {
			shared.ExitWithError("Error parsing cargo benchmark name", groupErr)
		}

		benchName, xAxis, yAxis := group["name"], group["xAxis"], group["yAxis"]

		results = append(results, shared.BenchmarkData{
			Name:  benchName,
			XAxis: xAxis,
			YAxis: yAxis,
			Stats: []shared.Stat{
				{Type: utils.CreateStatType("Latency avg", shared.FlagState.TimeUnit, ""), Value: utils.ConvertTime(estimateNs, "ns", shared.FlagState.TimeUnit)},
				{Type: utils.CreateStatType("Latency lower", shared.FlagState.TimeUnit, ""), Value: utils.ConvertTime(lowerNs, "ns", shared.FlagState.TimeUnit)},
				{Type: utils.CreateStatType("Latency upper", shared.FlagState.TimeUnit, ""), Value: utils.ConvertTime(upperNs, "ns", shared.FlagState.TimeUnit), Symbol: "±"},
			},
		})
	}

	if err := scanner.Err(); err != nil {
		shared.ExitWithError("failed to read file", err)
	}

	return results
}
