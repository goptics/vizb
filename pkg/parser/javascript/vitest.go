package javascript

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
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

func ParseVitestBenchmark(input io.Reader, cfg parser.Config) ([]shared.DataPoint, parser.Config, *shared.Meta, error) {
	scanner := bufio.NewScanner(input)
	var results []shared.DataPoint
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

		include, err := parser.ShouldIncludeBenchmark(name, cfg)
		if err != nil {
			return nil, cfg, nil, err
		}
		if !include {
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

		group, groupErr := parser.GroupBenchmarkName(name, cfg)
		if groupErr != nil {
			return nil, cfg, nil, fmt.Errorf("parse vitest benchmark name: %w", groupErr)
		}

		benchName, xAxis, yAxis, zAxis := group["name"], group["xAxis"], group["yAxis"], group["zAxis"]

		results = append(results, shared.DataPoint{
			Name:  benchName,
			XAxis: xAxis,
			YAxis: yAxis,
			ZAxis: zAxis,
			Stats: []shared.Stat{
				{Type: "Throughput avg (ops/s)", Value: shared.F64(hz)},
				{Type: utils.CreateStatType("Latency min", cfg.TimeUnit, ""), Value: shared.F64(utils.ConvertTime(minVal, "ms", cfg.TimeUnit))},
				{Type: utils.CreateStatType("Latency max", cfg.TimeUnit, ""), Value: shared.F64(utils.ConvertTime(maxVal, "ms", cfg.TimeUnit))},
				{Type: utils.CreateStatType("Latency avg", cfg.TimeUnit, ""), Value: shared.F64(utils.ConvertTime(mean, "ms", cfg.TimeUnit))},
				{Type: utils.CreateStatType("Latency p75", cfg.TimeUnit, ""), Value: shared.F64(utils.ConvertTime(p75, "ms", cfg.TimeUnit))},
				{Type: utils.CreateStatType("Latency p99", cfg.TimeUnit, ""), Value: shared.F64(utils.ConvertTime(p99, "ms", cfg.TimeUnit))},
				{Type: utils.CreateStatType("Latency p995", cfg.TimeUnit, ""), Value: shared.F64(utils.ConvertTime(p995, "ms", cfg.TimeUnit))},
				{Type: utils.CreateStatType("Latency p999", cfg.TimeUnit, ""), Value: shared.F64(utils.ConvertTime(p999, "ms", cfg.TimeUnit))},
				{Type: "RME (%)", Value: shared.F64(rme), Symbol: "±"},
				{Type: "Samples", Value: shared.F64(samples)},
			},
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, cfg, nil, fmt.Errorf("read vitest benchmark: %w", err)
	}

	return results, cfg, nil, nil
}
