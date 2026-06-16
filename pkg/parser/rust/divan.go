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
	parser.Parsers["rs:divan"] = ParseDivanBenchmark
}

var divanRowRe = regexp.MustCompile(`^[├╰]─\s+(\S+)\s+(.+)$`)

var divanValRe = regexp.MustCompile(`([\d.]+)\s*(ns|µs|μs|ms|s)`)

func ParseDivanBenchmark(filename string, cfg parser.Config) []shared.DataPoint {
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

		match := divanRowRe.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		name := match[1]
		if !parser.ShouldIncludeBenchmark(name, cfg) {
			continue
		}

		rest := match[2]
		parts := strings.Split(rest, "│")

		var values []float64
		for _, p := range parts {
			p = strings.TrimSpace(p)
			v := divanValRe.FindStringSubmatch(p)
			if v != nil {
				values = append(values, parseNum(v[1])*toNsMultiplier(v[2]))
			} else {
				values = append(values, parseNum(p))
			}
		}

		if len(values) < 5 {
			continue
		}

		fastestNs := values[0]
		slowestNs := values[1]
		medianNs := values[2]
		meanNs := values[3]
		samples := values[4]

		group, groupErr := parser.GroupBenchmarkName(name, cfg)
		if groupErr != nil {
			shared.ExitWithError("Error parsing divan benchmark name", groupErr)
		}

		benchName, xAxis, yAxis, zAxis := group["name"], group["xAxis"], group["yAxis"], group["zAxis"]

		results = append(results, shared.DataPoint{
			Name:  benchName,
			XAxis: xAxis,
			YAxis: yAxis,
			ZAxis: zAxis,
			Stats: []shared.Stat{
				{Type: utils.CreateStatType("Latency fastest", cfg.TimeUnit, ""), Value: utils.ConvertTime(fastestNs, "ns", cfg.TimeUnit)},
				{Type: utils.CreateStatType("Latency slowest", cfg.TimeUnit, ""), Value: utils.ConvertTime(slowestNs, "ns", cfg.TimeUnit), Symbol: "±"},
				{Type: utils.CreateStatType("Latency median", cfg.TimeUnit, ""), Value: utils.ConvertTime(medianNs, "ns", cfg.TimeUnit)},
				{Type: utils.CreateStatType("Latency mean", cfg.TimeUnit, ""), Value: utils.ConvertTime(meanNs, "ns", cfg.TimeUnit)},
				{Type: "Samples", Value: samples},
			},
		})
	}

	if err := scanner.Err(); err != nil {
		shared.ExitWithError("failed to read file", err)
	}

	return results
}
