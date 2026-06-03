package rust

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
)

func init() {
	parser.Parsers["rs:criterion"] = ParseCriterionBenchmark
}

var criterionRe = regexp.MustCompile(`^(\S+)\s+time:\s+\[([\d.]+)\s*(ns|µs|μs|ms|s)\s+([\d.]+)\s*(ns|µs|μs|ms|s)\s+([\d.]+)\s*(ns|µs|μs|ms|s)\]`)

func ParseCriterionBenchmark(filename string) []shared.BenchmarkData {
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

		benchName, xAxis, yAxis, zAxis := group["name"], group["xAxis"], group["yAxis"], group["zAxis"]

		results = append(results, shared.BenchmarkData{
			Name:  benchName,
			XAxis: xAxis,
			YAxis: yAxis,
			ZAxis: zAxis,
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
