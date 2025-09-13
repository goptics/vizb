package parser

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
	"golang.org/x/perf/benchfmt"
)

func storeCpuCount(cpu string) {
	if shared.CPUCount == 0 {
		if cpuCount, err := strconv.Atoi(cpu); err == nil {
			shared.CPUCount = cpuCount
		}
	}
}

func parseBenchmarkName(name benchfmt.Name) (base string, parts []string, cpu string) {
	b, ps := name.Parts()
	base = string(b)

	for _, b := range ps {
		part := string(b)

		if has := strings.HasPrefix(part, "-"); has {
			cpu = strings.TrimPrefix(part, "-")
		} else {
			parts = append(parts, strings.TrimPrefix(part, "/"))
		}
	}

	return
}

func parseBenchGroupsFromName(base string, parts []string) (benchName, workload, subject string) {
	switch l := len(parts); l {
	case 0:
		subject = base
	case 1:
		benchName, subject = base, parts[0]
	default:
		benchName, workload, subject = base, parts[l-2], parts[l-1]
	}
	return
}

func ParseBenchmarkResults(filePath string) (results []shared.BenchmarkResult, e error) {
	f, err := os.Open(filePath)
	if err != nil {
		e = err
		return
	}
	defer f.Close()
	reader := benchfmt.NewReader(f, filePath)

	for reader.Scan() {
		record := reader.Result()
		result, ok := record.(*benchfmt.Result)

		if !ok {
			continue
		}

		base, parts, cpu := parseBenchmarkName(result.Name)
		benchName, workload, subject := parseBenchGroupsFromName(base, parts)
		storeCpuCount(cpu)

		var benchStats []shared.Stat

		for _, value := range result.Values {
			var benchStat shared.Stat

			switch value.Unit {
			case "sec/op":
				benchStat = shared.Stat{
					Type:  "Execution Time",
					Value: utils.FormatTime(value.OrigValue, shared.FlagState.TimeUnit),
					Unit:  shared.FlagState.TimeUnit,
				}
			case "B/op":
				shared.HasMemStats = true

				benchStat = shared.Stat{
					Type:  "Memory Usage",
					Value: utils.FormatMem(value.Value, shared.FlagState.MemUnit),
					Unit:  shared.FlagState.MemUnit,
				}
			case "allocs/op":
				benchStat = shared.Stat{
					Type:  "Allocations",
					Value: utils.FormatAllocs(value.Value, shared.FlagState.AllocUnit),
					Unit:  shared.FlagState.AllocUnit,
				}
			}

			benchStats = append(benchStats, benchStat)
		}

		results = append(results, shared.BenchmarkResult{
			Name:     benchName,
			Workload: workload,
			Subject:  subject,
			Stats:    benchStats,
		})
	}

	return
}

func ConvertJsonBenchToText(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	// TODO: Create a global temp file creator function
	tempFile, err := os.CreateTemp("", "vizb-input.txt")
	if err != nil {
		return "", err
	}

	file := tempFile.Name()

	defer tempFile.Close()

	dec := json.NewDecoder(f)
	writer := bufio.NewWriter(tempFile)

	for {
		var ev shared.BenchEvent
		if err := dec.Decode(&ev); err != nil {
			if err == io.EOF {
				break
			}

			return "", err
		}

		if ev.Action == "output" {
			writer.WriteString(ev.Output)
		}
	}

	if err := writer.Flush(); err != nil {
		return "", err
	}

	tempFile.Sync()

	return file, nil
}
