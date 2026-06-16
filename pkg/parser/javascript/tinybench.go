package javascript

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
)

func init() {
	parser.Parsers["js:tinybench"] = ParseTinyBenchBenchmark
}

var dataRowRe = regexp.MustCompile(`│\s*\d+\s*│\s*(.+)\s*│`)

func parseMeanWithRME(s string) (mean, rme float64) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "'")
	s = strings.TrimSuffix(s, "'")

	parts := strings.Split(s, "±")
	if len(parts) == 2 {
		mean, _ = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		rmeStr := strings.TrimSpace(parts[1])
		rmeStr = strings.TrimSuffix(rmeStr, "%")
		rme, _ = strconv.ParseFloat(rmeStr, 64)
	} else {
		mean, _ = strconv.ParseFloat(strings.TrimSpace(s), 64)
	}
	return
}

func parseMedWithMAD(s string) (med, mad float64) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "'")
	s = strings.TrimSuffix(s, "'")

	parts := strings.Split(s, "±")
	if len(parts) == 2 {
		med, _ = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		mad, _ = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	} else {
		med, _ = strconv.ParseFloat(strings.TrimSpace(s), 64)
	}
	return
}

func ParseTinyBenchBenchmark(filename string, cfg parser.Config) []shared.DataPoint {
	f, err := os.Open(filename)
	if err != nil {
		shared.ExitWithError("Error opening file", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var results []shared.DataPoint

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.Contains(line, "│") || strings.Contains(line, "Task name") || strings.Contains(line, "┌") || strings.Contains(line, "├") || strings.Contains(line, "└") {
			continue
		}

		match := dataRowRe.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		columns := strings.Split(match[1], "│")
		if len(columns) < 6 {
			continue
		}

		taskName := strings.Trim(columns[0], " '")

		if !parser.ShouldIncludeBenchmark(taskName, cfg) {
			continue
		}

		latencyAvgStr := strings.Trim(columns[1], " '")
		latencyMedStr := strings.Trim(columns[2], " '")
		throughputAvgStr := strings.Trim(columns[3], " '")
		throughputMedStr := strings.Trim(columns[4], " '")
		samplesStr := strings.Trim(columns[5], " ")

		latencyAvg, latencyRME := parseMeanWithRME(latencyAvgStr)
		latencyMed, latencyMAD := parseMedWithMAD(latencyMedStr)
		throughputAvg, throughputRME := parseMeanWithRME(throughputAvgStr)
		throughputMed, throughputMAD := parseMedWithMAD(throughputMedStr)
		samples, _ := strconv.ParseFloat(samplesStr, 64)

		group, groupErr := parser.GroupBenchmarkName(taskName, cfg)
		if groupErr != nil {
			shared.ExitWithError("Error parsing tinybench name", groupErr)
		}

		benchName, xAxis, yAxis, zAxis := group["name"], group["xAxis"], group["yAxis"], group["zAxis"]

		results = append(results, shared.DataPoint{
			Name:  benchName,
			XAxis: xAxis,
			YAxis: yAxis,
			ZAxis: zAxis,
			Stats: []shared.Stat{
				{Type: utils.CreateStatType("Latency avg", cfg.TimeUnit, ""), Value: utils.FormatTime(latencyAvg, cfg.TimeUnit)},
				{Type: "Latency RME (%)", Value: latencyRME, Symbol: "±"},
				{Type: utils.CreateStatType("Latency med", cfg.TimeUnit, ""), Value: utils.FormatTime(latencyMed, cfg.TimeUnit)},
				{Type: utils.CreateStatType("Latency MAD", cfg.TimeUnit, ""), Value: utils.FormatTime(latencyMAD, cfg.TimeUnit), Symbol: "±"},
				{Type: "Throughput avg (ops/s)", Value: throughputAvg},
				{Type: "Throughput RME (%)", Value: throughputRME, Symbol: "±"},
				{Type: "Throughput med (ops/s)", Value: throughputMed},
				{Type: "Throughput MAD (ops/s)", Value: throughputMAD, Symbol: "±"},
				{Type: "Samples", Value: samples},
			},
		})
	}

	if err := scanner.Err(); err != nil {
		shared.ExitWithError("failed to read file", err)
	}

	return results
}
