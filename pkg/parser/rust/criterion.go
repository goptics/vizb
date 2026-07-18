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
	parser.Parsers["rs:criterion"] = ParseCriterionBenchmark
}

var criterionRe = regexp.MustCompile(`^(\S+)\s+time:\s+\[([\d.]+)\s*(ns|µs|μs|ms|s)\s+([\d.]+)\s*(ns|µs|μs|ms|s)\s+([\d.]+)\s*(ns|µs|μs|ms|s)\]`)

func ParseCriterionBenchmark(input io.Reader, cfg parser.Config) ([]shared.DataPoint, parser.Config, *shared.Meta, error) {
	scanner := bufio.NewScanner(input)
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
		include, err := parser.ShouldIncludeBenchmark(name, cfg)
		if err != nil {
			return nil, cfg, nil, err
		}
		if !include {
			continue
		}

		lowerNs := parseNum(match[2]) * toNsMultiplier(match[3])
		estimateNs := parseNum(match[4]) * toNsMultiplier(match[5])
		upperNs := parseNum(match[6]) * toNsMultiplier(match[7])

		group, groupErr := parser.GroupBenchmarkName(name, cfg)
		if groupErr != nil {
			return nil, cfg, nil, fmt.Errorf("parse criterion benchmark name: %w", groupErr)
		}

		benchName, xAxis, yAxis, zAxis := group["name"], group["xAxis"], group["yAxis"], group["zAxis"]

		results = append(results, shared.DataPoint{
			Name:  benchName,
			XAxis: xAxis,
			YAxis: yAxis,
			ZAxis: zAxis,
			Stats: []shared.Stat{
				{Type: utils.CreateStatType("Latency avg", cfg.TimeUnit, ""), Value: shared.F64(utils.ConvertTime(estimateNs, "ns", cfg.TimeUnit))},
				{Type: utils.CreateStatType("Latency lower", cfg.TimeUnit, ""), Value: shared.F64(utils.ConvertTime(lowerNs, "ns", cfg.TimeUnit))},
				{Type: utils.CreateStatType("Latency upper", cfg.TimeUnit, ""), Value: shared.F64(utils.ConvertTime(upperNs, "ns", cfg.TimeUnit)), Symbol: "±"},
			},
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, cfg, nil, fmt.Errorf("read criterion benchmark: %w", err)
	}

	return results, cfg, nil, nil
}
