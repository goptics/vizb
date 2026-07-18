package golang

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
	"golang.org/x/perf/benchfmt"
)

func init() {
	parser.Parsers["go"] = ParseGoBenchmark
}

func storeCpuCount(cpu string) {
	if shared.CPUCount == 0 {
		if cpuCount, err := strconv.Atoi(cpu); err == nil {
			shared.CPUCount = cpuCount
		}
	}
}

func parseBenchmarkName(name benchfmt.Name) (benchName string, cpu string) {
	b, ps := name.Parts()
	benchName = string(b)

	for _, b := range ps {
		part := string(b)

		if has := strings.HasPrefix(part, "-"); has {
			cpu = strings.TrimPrefix(part, "-")
		} else {
			benchName += part
		}
	}

	return
}

func ParseGoBenchmark(input io.Reader, cfg parser.Config) ([]shared.DataPoint, parser.Config, error) {
	var results []shared.DataPoint

	benchmarkInput, err := prepareBenchmarkInput(input)
	if err != nil {
		return nil, cfg, err
	}
	reader := benchfmt.NewReader(benchmarkInput, "input")

	var allIters []int

	for reader.Scan() {
		record := reader.Result()
		result, ok := record.(*benchfmt.Result)

		if !ok {
			continue
		}

		shared.OS, shared.Arch, shared.Pkg, shared.CPU = result.GetConfig("goos"), result.GetConfig("goarch"), result.GetConfig("pkg"), result.GetConfig("cpu")
		rawBenchName, cpuCore := parseBenchmarkName(result.Name)

		include, err := parser.ShouldIncludeBenchmark(rawBenchName, cfg)
		if err != nil {
			return nil, cfg, err
		}
		if !include {
			continue
		}

		group, err := parser.GroupBenchmarkName(rawBenchName, cfg)

		if err != nil {
			return nil, cfg, fmt.Errorf("parse group from benchmark name: %w", err)
		}

		benchName, xAxis, yAxis, zAxis := group["name"], group["xAxis"], group["yAxis"], group["zAxis"]

		storeCpuCount(cpuCore)

		var benchStats []shared.Stat

		for _, value := range result.Values {
			var benchStat shared.Stat

			switch value.Unit {
			case "sec/op":
				benchStat = shared.Stat{
					Type:  utils.CreateStatType("Execution Time", cfg.TimeUnit, "op"),
					Value: shared.F64(utils.FormatTime(value.OrigValue, cfg.TimeUnit)),
				}
			case "B/op":
				benchStat = shared.Stat{
					Type:  utils.CreateStatType("Memory Usage", cfg.MemUnit, "op"),
					Value: shared.F64(utils.FormatMem(value.Value, cfg.MemUnit)),
				}
			case "allocs/op":
				benchStat = shared.Stat{
					Type:  utils.CreateStatType("Allocations", cfg.NumberUnit, "op"),
					Value: shared.F64(utils.FormatNumber(value.Value, cfg.NumberUnit)),
				}
			case "B/s", "MB/s", "GB/s":
				val, unit := value.OrigValue, value.OrigUnit

				if val == 0 || unit == "" {
					val, unit = value.Value, value.Unit
				}

				benchStat = shared.Stat{
					Type:  utils.CreateStatType("Throughput", unit, ""),
					Value: shared.F64(val),
				}
			default:
				customType := "Metric"

				if strings.HasSuffix(value.Unit, "/s") {
					customType = "Throughput"
				}

				benchStat = shared.Stat{
					Type:  utils.CreateStatType(customType, value.Unit, ""),
					Value: shared.F64(value.Value),
				}
			}

			benchStats = append(benchStats, benchStat)
		}

		results = append(results, shared.DataPoint{
			Name:  benchName,
			XAxis: xAxis,
			YAxis: yAxis,
			ZAxis: zAxis,
			Stats: benchStats,
		})

		allIters = append(allIters, result.Iters)
	}
	if err := reader.Err(); err != nil {
		return nil, cfg, fmt.Errorf("read Go benchmark: %w", err)
	}

	hasDifferentIters := false
	if len(allIters) > 1 {
		firstIter := allIters[0]
		for _, iter := range allIters[1:] {
			if iter != firstIter {
				hasDifferentIters = true
				break
			}
		}
	}

	if hasDifferentIters {
		for i := range results {
			results[i].Stats = append(results[i].Stats, shared.Stat{
				Type:  utils.CreateStatType("Iterations", cfg.NumberUnit, ""),
				Value: shared.F64(utils.FormatNumber(float64(allIters[i]), cfg.NumberUnit)),
			})
		}
	}

	return results, cfg, nil
}

// prepareBenchmarkInput converts Go test -json events to their benchmark text
// while leaving regular benchmark output streaming through to benchfmt.
func prepareBenchmarkInput(input io.Reader) (io.Reader, error) {
	reader := bufio.NewReader(input)
	var prefix strings.Builder

	for {
		line, err := reader.ReadString('\n')
		prefix.WriteString(line)
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			var event shared.BenchEvent
			if json.Unmarshal([]byte(trimmed), &event) != nil {
				return io.MultiReader(strings.NewReader(prefix.String()), reader), nil
			}
			if event.Action == "" {
				return io.MultiReader(strings.NewReader(prefix.String()), reader), nil
			}
			return readBenchmarkEvents(reader, event, err)
		}

		if err == io.EOF {
			return strings.NewReader(prefix.String()), nil
		}
		if err != nil {
			return nil, fmt.Errorf("read Go benchmark: %w", err)
		}
	}
}

func readBenchmarkEvents(reader *bufio.Reader, first shared.BenchEvent, firstErr error) (io.Reader, error) {
	var output strings.Builder
	if first.Action == "output" {
		output.WriteString(first.Output)
	}
	if firstErr == io.EOF {
		return strings.NewReader(output.String()), nil
	}
	if firstErr != nil {
		return nil, fmt.Errorf("read Go benchmark JSON: %w", firstErr)
	}

	for {
		line, err := reader.ReadString('\n')
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			var event shared.BenchEvent
			if decodeErr := json.Unmarshal([]byte(trimmed), &event); decodeErr != nil {
				return nil, fmt.Errorf("read Go benchmark JSON: %w", decodeErr)
			}
			if event.Action == "output" {
				output.WriteString(event.Output)
			}
		}

		if err == io.EOF {
			return strings.NewReader(output.String()), nil
		}
		if err != nil {
			return nil, fmt.Errorf("read Go benchmark JSON: %w", err)
		}
	}
}
