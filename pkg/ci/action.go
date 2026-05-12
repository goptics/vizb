package ci

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"golang.org/x/perf/benchfmt"
)

type ActionOpts struct {
	Input      string
	Version    string
	Tag        string
	Branch     string
	Date       time.Time
	MergeFile  string
	Output    string
	KeepCount int
}

func RunAction(opts ActionOpts) (*shared.Run, *shared.Benchmark, error) {
	f, err := os.Open(opts.Input)
	if err != nil {
		return nil, nil, fmt.Errorf("open input: %w", err)
	}
	defer f.Close()

	reader := benchfmt.NewReader(f, opts.Input)
	var goos, goarch, pkgName, cpu string
	var benchResults []shared.BenchmarkResult

	for reader.Scan() {
		result, ok := reader.Result().(*benchfmt.Result)
		if !ok {
			continue
		}
		goos = result.GetConfig("goos")
		goarch = result.GetConfig("goarch")
		pkgName = result.GetConfig("pkg")
		cpu = result.GetConfig("cpu")

		rawBenchName, _ := parseBenchName(result.Name)
		br := shared.BenchmarkResult{Name: rawBenchName, Pkg: pkgName}

		for _, value := range result.Values {
			switch value.Unit {
			case "sec/op":
				br.NsPerOp = value.Value * 1e9
			case "B/op":
				br.BytesPerOp = value.Value
			case "allocs/op":
				br.AllocsPerOp = value.Value
			case "B/s", "MB/s", "GB/s":
				br.MBPerSec = value.Value
			}
		}
		benchResults = append(benchResults, br)
	}
	if err := reader.Err(); err != nil {
		return nil, nil, fmt.Errorf("read benchmarks: %w", err)
	}

	run := &shared.Run{
		Version:    opts.Version,
		Tag:        opts.Tag,
		Date:       opts.Date,
		Branch:     opts.Branch,
		Goos:       goos,
		Goarch:     goarch,
		CPU:        cpu,
		Benchmarks: benchResults,
	}

	var newData []shared.BenchmarkData
	for _, br := range benchResults {
		name, yAxis := splitBenchName(br.Name)
		newData = append(newData, shared.BenchmarkData{
			Name:  name,
			XAxis: opts.Tag,
			YAxis: yAxis,
			Stats: resultToStats(br),
		})
	}

	bench := shared.Benchmark{
		Name: pkgName,
		Pkg:  pkgName,
		OS:   goos,
		Arch: goarch,
		Data: newData,
	}
	bench.CPU.Name = cpu
	bench.Settings.Charts = []string{"bar", "line", "pie"}
	bench.Settings.ShowLabels = true

	// Handle merge: replace data items with matching xAxis (tag)
	if opts.MergeFile != "" {
		existing, err := shared.ReadJSONFile[shared.Benchmark](opts.MergeFile)
		if err == nil {
			// Remove existing items with same xAxis (tag)
			var filteredData []shared.BenchmarkData
			for _, d := range existing.Data {
				if d.XAxis != opts.Tag {
					filteredData = append(filteredData, d)
				}
			}
			bench.Data = append(filteredData, newData...)
			bench.Name = existing.Name
			bench.Description = existing.Description
			bench.CPU = existing.CPU
			bench.OS = existing.OS
			bench.Arch = existing.Arch
			bench.Settings = existing.Settings
			bench.Runtimes = existing.Runtimes
		} else if !errors.Is(err, os.ErrNotExist) {
			// Check if the underlying error is "file not found"
			// ReadJSONFile wraps: "read file %s: %w"
			// We just ignore if file doesn't exist
		}
	}

	// Track runtime for this tag/commit
	if bench.Runtimes == nil {
		bench.Runtimes = make(map[string]time.Time)
	}
	bench.Runtimes[opts.Tag] = opts.Date

	// Prune: keep at most PruneCount unique xAxis values (by most recent runtime)
	if opts.KeepCount > 0 {
		pruneBenchData(&bench, opts.KeepCount)
	}

	return run, &bench, nil
}

// TagDimension returns which dimension (name, xAxis, yAxis) is NOT covered
// by the 2D CI pattern or regex — this is where the tag/commit goes.
func TagDimension(pattern, regexStr string) (string, error) {
	all := []string{"name", "xAxis", "yAxis"}
	var present []string

	if regexStr != "" {
		re, err := regexp.Compile(regexStr)
		if err != nil {
			return "", fmt.Errorf("invalid regex: %w", err)
		}
		for _, name := range re.SubexpNames() {
			if name == "" {
				continue
			}
			present = append(present, expandCIShorthand(name))
		}
	} else {
		if err := validateCIPattern(pattern); err != nil {
			return "", err
		}
		present = parser.ParsePatternParts(pattern)
	}

	presentSet := make(map[string]bool, len(present))
	for _, p := range present {
		presentSet[p] = true
	}

	var missing []string
	for _, d := range all {
		if !presentSet[d] {
			missing = append(missing, d)
		}
	}

	if len(missing) != 1 {
		return "", fmt.Errorf("CI pattern must define exactly 2 dimensions, got %d defined (%v), missing %d (%v)",
			len(presentSet), present, len(missing), missing)
	}

	return missing[0], nil
}

// InjectTag fills the missing dimension in each BenchmarkData with the tag value.
func InjectTag(data []shared.BenchmarkData, tag, pattern, regexStr string) ([]shared.BenchmarkData, error) {
	if tag == "" {
		return data, nil
	}

	tagDim, err := TagDimension(pattern, regexStr)
	if err != nil {
		return nil, err
	}

	result := make([]shared.BenchmarkData, len(data))
	for i, d := range data {
		tagged := d
		setDim(&tagged, tagDim, tag)
		result[i] = tagged
	}
	return result, nil
}

func validateCIPattern(pattern string) error {
	parts := parser.ParsePatternParts(pattern)
	if len(parts) != 2 {
		return fmt.Errorf("CI patterns must have exactly 2 dimensions, got %d: use n/y, n/x, x/y, etc.", len(parts))
	}
	if err := parser.ValidateGroupPattern(pattern); err != nil {
		return err
	}
	return nil
}

func setDim(d *shared.BenchmarkData, dim, value string) {
	switch dim {
	case "name":
		d.Name = value
	case "xAxis":
		d.XAxis = value
	case "yAxis":
		d.YAxis = value
	}
}

func getDim(d shared.BenchmarkData, dim string) string {
	switch dim {
	case "name":
		return d.Name
	case "xAxis":
		return d.XAxis
	case "yAxis":
		return d.YAxis
	}
	return ""
}

func expandCIShorthand(part string) string {
	shortcuts := map[string]string{"n": "name", "x": "xAxis", "y": "yAxis"}
	if expanded, exists := shortcuts[part]; exists {
		return expanded
	}
	return part
}

func pruneBenchData(bench *shared.Benchmark, keep int) {
	tags := make([]string, 0, len(bench.Runtimes))
	for tag := range bench.Runtimes {
		tags = append(tags, tag)
	}
	if len(tags) <= keep {
		return
	}

	sort.Slice(tags, func(i, j int) bool {
		return bench.Runtimes[tags[i]].After(bench.Runtimes[tags[j]])
	})

	keepSet := make(map[string]bool, keep)
	for i := 0; i < keep; i++ {
		keepSet[tags[i]] = true
	}

	var filtered []shared.BenchmarkData
	for _, d := range bench.Data {
		if keepSet[d.XAxis] {
			filtered = append(filtered, d)
		}
	}
	bench.Data = filtered

	for tag := range bench.Runtimes {
		if !keepSet[tag] {
			delete(bench.Runtimes, tag)
		}
	}
}

func parseBenchName(name benchfmt.Name) (string, string) {
	b, ps := name.Parts()
	benchName := string(b)
	var cpu string
	for _, p := range ps {
		part := string(p)
		if len(part) > 0 && part[0] == '-' {
			cpu = part[1:]
		} else {
			benchName += part
		}
	}
	return benchName, cpu
}

func splitBenchName(fullName string) (string, string) {
	name := strings.TrimPrefix(fullName, "Benchmark")
	parts := strings.SplitN(name, "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func resultToStats(br shared.BenchmarkResult) []shared.Stat {
	var stats []shared.Stat
	if br.NsPerOp > 0 {
		stats = append(stats, shared.Stat{Type: "Execution Time (ns/op)", Value: br.NsPerOp})
	}
	if br.BytesPerOp > 0 {
		stats = append(stats, shared.Stat{Type: "Memory Usage (B/op)", Value: br.BytesPerOp})
	}
	if br.AllocsPerOp > 0 {
		stats = append(stats, shared.Stat{Type: "Allocations/op", Value: br.AllocsPerOp})
	}
	if br.MBPerSec > 0 {
		stats = append(stats, shared.Stat{Type: "Throughput (MB/s)", Value: br.MBPerSec})
	}
	return stats
}
