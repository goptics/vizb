package parser

import (
	"bufio"
	"encoding/json"
	"io"
	"regexp"
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

// shouldIncludeBenchmark returns true if the benchmark name should be included
// based on the filter regex. If no filter is set, all benchmarks are included.
func shouldIncludeBenchmark(benchName string) bool {
	if shared.FlagState.FilterRegex == "" {
		return true
	}

	filterRe, err := regexp.Compile(shared.FlagState.FilterRegex)
	if err != nil {
		shared.ExitWithError("Invalid filter regex", err)
	}

	return filterRe.MatchString(benchName)
}

func ParseBenchmarkData(filePath string) (results []shared.BenchmarkData) {
	f := shared.MustOpenFile(filePath)
	defer f.Close()

	reader := benchfmt.NewReader(f, filePath)

	var allIters []int

	for reader.Scan() {
		record := reader.Result()
		result, ok := record.(*benchfmt.Result)

		if !ok {
			continue
		}

		shared.OS, shared.Arch, shared.Pkg, shared.CPU = result.GetConfig("goos"), result.GetConfig("goarch"), result.GetConfig("pkg"), result.GetConfig("cpu")
		rawBenchName, cpuCore := parseBenchmarkName(result.Name)

		if !shouldIncludeBenchmark(rawBenchName) {
			continue
		}

		var group map[string]string
		var err error

		if shared.FlagState.GroupRegex != "" {
			group, err = ParseBenchmarkNameWithRegex(rawBenchName, shared.FlagState.GroupRegex)
		} else {
			group, err = ParseBenchmarkNameToGroups(rawBenchName, shared.FlagState.GroupPattern)
		}

		if err != nil {
			shared.ExitWithError("Error on parsing group from bench name", err)
		}

		benchName, xAxis, yAxis := group["name"], group["xAxis"], group["yAxis"]

		storeCpuCount(cpuCore)

		var benchStats []shared.Stat

		for _, value := range result.Values {
			var benchStat shared.Stat

			switch value.Unit {
			case "sec/op":
				benchStat = shared.Stat{
					Type:  utils.CreateStatType("Execution Time", shared.FlagState.TimeUnit, "op"),
					Value: utils.FormatTime(value.OrigValue, shared.FlagState.TimeUnit),
				}
			case "B/op":
				benchStat = shared.Stat{
					Type:  utils.CreateStatType("Memory Usage", shared.FlagState.MemUnit, "op"),
					Value: utils.FormatMem(value.Value, shared.FlagState.MemUnit),
				}
			case "allocs/op":
				benchStat = shared.Stat{
					Type:  utils.CreateStatType("Allocations", shared.FlagState.NumberUnit, "op"),
					Value: utils.FormatNumber(value.Value, shared.FlagState.NumberUnit),
				}
			case "B/s", "MB/s", "GB/s":
				// benchfmt only populates OrigValue/OrigUnit for MB/s
				// For B/s and GB/s, fall back to Value/Unit
				val, unit := value.OrigValue, value.OrigUnit

				if val == 0 || unit == "" {
					val, unit = value.Value, value.Unit
				}

				benchStat = shared.Stat{
					Type:  utils.CreateStatType("Throughput", unit, ""),
					Value: val,
				}
			default:
				customType := "Metric"

				if strings.HasSuffix(value.Unit, "/s") {
					customType = "Throughput"
				}

				benchStat = shared.Stat{
					Type:  utils.CreateStatType(customType, value.Unit, ""),
					Value: value.Value,
				}
			}

			benchStats = append(benchStats, benchStat)
		}

		results = append(results, shared.BenchmarkData{
			Name:  benchName,
			XAxis: xAxis,
			YAxis: yAxis,
			Stats: benchStats,
		})

		allIters = append(allIters, result.Iters)
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
				Type:  utils.CreateStatType("Iterations", shared.FlagState.NumberUnit, ""),
				Value: utils.FormatNumber(float64(allIters[i]), shared.FlagState.NumberUnit),
			})
		}
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
