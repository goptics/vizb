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

func ParseCriterionBenchmark(filename string, cfg parser.Config) []shared.DataPoint {
	f, err := os.Open(filename)
	if err != nil {
		shared.ExitWithError("Error opening file", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var results []shared.DataPoint

	for scanner.Scan() {
		line := scanner.Text()
		line = ansiRe.ReplaceAllString(line, "")
		line = strings.TrimLeftFunc(line, unicode.IsSpace)

		match := criterionRe.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		name := match[1]
		if !parser.ShouldIncludeBenchmark(name, cfg) {
			continue
		}

		lowerNs := parseNum(match[2]) * toNsMultiplier(match[3])
		estimateNs := parseNum(match[4]) * toNsMultiplier(match[5])
		upperNs := parseNum(match[6]) * toNsMultiplier(match[7])

		group, groupErr := parser.GroupBenchmarkName(name, cfg)
		if groupErr != nil {
			shared.ExitWithError("Error parsing cargo benchmark name", groupErr)
		}

		benchName, xAxis, yAxis, zAxis := group["name"], group["xAxis"], group["yAxis"], group["zAxis"]

		results = append(results, shared.DataPoint{
			Name:  benchName,
			XAxis: xAxis,
			YAxis: yAxis,
			ZAxis: zAxis,
			Stats: []shared.Stat{
				{Type: utils.CreateStatType("Latency avg", cfg.TimeUnit, ""), Value: utils.ConvertTime(estimateNs, "ns", cfg.TimeUnit)},
				{Type: utils.CreateStatType("Latency lower", cfg.TimeUnit, ""), Value: utils.ConvertTime(lowerNs, "ns", cfg.TimeUnit)},
				{Type: utils.CreateStatType("Latency upper", cfg.TimeUnit, ""), Value: utils.ConvertTime(upperNs, "ns", cfg.TimeUnit), Symbol: "±"},
			},
		})
	}

	if err := scanner.Err(); err != nil {
		shared.ExitWithError("failed to read file", err)
	}

	return results
}
