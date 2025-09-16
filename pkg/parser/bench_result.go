package parser

import (
	"bufio"
	"encoding/json"
	"io"
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

func ParseBenchmarkResults(filePath string) (results []shared.BenchmarkResult) {
	f := shared.MustOpenFile(filePath)
	defer f.Close()

	reader := benchfmt.NewReader(f, filePath)

	for reader.Scan() {
		record := reader.Result()
		result, ok := record.(*benchfmt.Result)

		if !ok {
			continue
		}

		rawBenchName, cpu := parseBenchmarkName(result.Name)
		group, err := ParseBenchmarkNameToGroups(rawBenchName, shared.FlagState.GroupPattern)

		if err != nil {

		}

		benchName, workload, subject := group["name"], group["workload"], group["subject"]

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

// ConvertJsonBenchToText converts a JSON benchmark file to text format.
// It reads JSON benchmark events from the input file and extracts the "output" field
// from each event to create a text representation of the benchmark results.
// Returns the path to the temporary text file containing the converted data.
func ConvertJsonBenchToText(filePath string) string {
	f := shared.MustOpenFile(filePath)
	tempFilePath := shared.MustCreateTempFile(shared.TempBenchFilePrefix, "txt")
	tempFile := shared.MustCreateFile(tempFilePath)
	shared.TempFiles.Store(tempFilePath)

	defer f.Close()
	defer tempFile.Close()

	dec := json.NewDecoder(f)
	writer := bufio.NewWriter(tempFile)

	for {
		var ev shared.BenchEvent
		if err := dec.Decode(&ev); err != nil {
			if err == io.EOF {
				break
			}

			shared.ExitWithError("Error on converting json to text", err)
		}

		if ev.Action == "output" {
			writer.WriteString(ev.Output)
		}
	}

	if err := writer.Flush(); err != nil {
		shared.ExitWithError("Error on converting json to text", err)
	}

	tempFile.Sync()

	return tempFilePath
}
