package rust

import (
	"bufio"
	"fmt"
	"io"
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

// ParseDivanBenchmark converts Divan benchmark output into data points.
func ParseDivanBenchmark(input io.Reader, cfg parser.Config) ([]shared.DataPoint, parser.Config, *shared.Meta, error) {
	scanner := bufio.NewScanner(input)
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
		include, err := parser.ShouldIncludeBenchmark(name, cfg)
		if err != nil {
			return nil, cfg, nil, err
		}
		if !include {
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
			return nil, cfg, nil, fmt.Errorf("parse divan benchmark name: %w", groupErr)
		}

		benchName, xAxis, yAxis, zAxis := group["name"], group["xAxis"], group["yAxis"], group["zAxis"]

		results = append(results, shared.DataPoint{
			Name:  benchName,
			XAxis: xAxis,
			YAxis: yAxis,
			ZAxis: zAxis,
			Stats: []shared.Stat{
				{Type: utils.CreateStatType("Latency fastest", cfg.TimeUnit, ""), Value: shared.F64(utils.ConvertTime(fastestNs, "ns", cfg.TimeUnit))},
				{Type: utils.CreateStatType("Latency slowest", cfg.TimeUnit, ""), Value: shared.F64(utils.ConvertTime(slowestNs, "ns", cfg.TimeUnit)), Symbol: "±"},
				{Type: utils.CreateStatType("Latency median", cfg.TimeUnit, ""), Value: shared.F64(utils.ConvertTime(medianNs, "ns", cfg.TimeUnit))},
				{Type: utils.CreateStatType("Latency mean", cfg.TimeUnit, ""), Value: shared.F64(utils.ConvertTime(meanNs, "ns", cfg.TimeUnit))},
				{Type: "Samples", Value: shared.F64(samples)},
			},
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, cfg, nil, fmt.Errorf("read divan benchmark: %w", err)
	}

	return results, cfg, nil, nil
}
