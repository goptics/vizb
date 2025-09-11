package parser

import (
	"encoding/json"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
)

var (
	// Match benchmark output lines with memory stats (when -benchmem is used)
	benchMemLineRe = regexp.MustCompile(`Benchmark[^\s]+\s+\d+\s+([\d\.]+)\s+ns/op\s+([\d\.]+)\s+B/op\s+([\d\.]+)\s+allocs/op`)
	// Match benchmark output lines without memory stats
	benchLineRe = regexp.MustCompile(`Benchmark[^\s]+\s+\d+\s+([\d\.]+)\s+ns/op`)
)

func parseBenchGroupsFromName(name string) (benchName, workload, subject string) {
	nameParts := strings.Split(
		strings.TrimPrefix(name, "Benchmark"), // remove the prefix Benchmark keyword
		shared.FlagState.Separator,
	)

	switch l := len(nameParts); l {
	case 1:
		subject = nameParts[0]
	case 2:
		benchName, subject = nameParts[0], nameParts[1]
	default:
		benchName, workload, subject = nameParts[l-3], nameParts[l-2], nameParts[l-1]
	}
	return
}

func ParseBenchmarkResults(jsonPath string) (results []shared.BenchmarkResult, e error) {
	f, err := os.Open(jsonPath)
	if err != nil {
		e = err
		return
	}
	defer f.Close()

	dec := json.NewDecoder(f)

	for {
		var ev shared.BenchEvent
		if err := dec.Decode(&ev); err != nil {
			if err == io.EOF {
				break
			}
			e = err
			return
		}

		// We're looking for output lines that contain benchmark results
		if ev.Action != "output" || !strings.Contains(ev.Output, "ns/op") {
			continue
		}

		var stats []string

		// First try to match with memory stats
		if memMatch := benchMemLineRe.FindStringSubmatch(ev.Output); memMatch != nil {
			stats = memMatch
			shared.HasMemStats = true
		} else if basicMatch := benchLineRe.FindStringSubmatch(ev.Output); basicMatch != nil {
			stats = basicMatch
		} else {
			continue
		}
		// Extract the benchmark name from the output
		parts := strings.Fields(ev.Output)

		if len(parts) == 0 {
			continue
		}

		benchName, workload, subject := parseBenchGroupsFromName(parts[0]) // the first item is the name of the benchmark

		// Remove CPU suffix from subject (e.g., "Subject-8" -> "Subject")
		if idx := strings.LastIndex(subject, "-"); idx > 0 {
			// Check if everything after the dash is a number
			if cpuCount, err := strconv.Atoi(subject[idx+1:]); err == nil {
				// Store CPU count in global bench state
				subject = subject[:idx]

				if shared.CPUCount == 0 {
					shared.CPUCount = cpuCount
				}
			}
		}

		var benchStats []shared.Stat

		// Parse metrics
		timePerOp, _ := strconv.ParseFloat(stats[1], 64)

		benchStats = append(benchStats, shared.Stat{
			Type:  "Execution Time",
			Value: utils.FormatTime(timePerOp, shared.FlagState.TimeUnit),
			Unit:  shared.FlagState.TimeUnit,
		})
		// Default values for memory stats
		var memPerOp float64
		var allocsPerOp uint64

		// If we have memory stats, parse them
		if shared.HasMemStats && len(stats) >= 4 {
			memPerOp, _ = strconv.ParseFloat(stats[2], 64)
			allocsPerOp, _ = strconv.ParseUint(stats[3], 10, 64)

			benchStats = append(
				benchStats,
				shared.Stat{
					Type:  "Memory Usage",
					Value: utils.FormatMem(memPerOp, shared.FlagState.MemUnit),
					Unit:  shared.FlagState.MemUnit,
				},
				shared.Stat{
					Type:  "Allocations",
					Value: float64(utils.FormatAllocs(allocsPerOp, shared.FlagState.AllocUnit)),
					Unit:  shared.FlagState.AllocUnit,
				},
			)
		}

		benchResult := shared.BenchmarkResult{
			Name:     benchName,
			Workload: workload,
			Subject:  subject,
			Stats:    benchStats,
		}

		results = append(results, benchResult)
	}

	return
}
