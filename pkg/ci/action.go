package ci

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
)

type ActionOpts struct {
	Input        string
	Version      string
	Tag          string
	Branch       string
	Date         time.Time
	MergeFile    string
	Output       string
	KeepCount    int
	GroupPattern string
	GroupRegex   string
}

func RunAction(opts ActionOpts) (*shared.Benchmark, error) {
	if _, err := os.Stat(opts.Input); err != nil {
		return nil, fmt.Errorf("input file: %w", err)
	}

	prevPattern := shared.FlagState.GroupPattern
	prevRegex := shared.FlagState.GroupRegex
	shared.FlagState.GroupPattern = opts.GroupPattern
	shared.FlagState.GroupRegex = opts.GroupRegex
	defer func() {
		shared.FlagState.GroupPattern = prevPattern
		shared.FlagState.GroupRegex = prevRegex
	}()

	data := parser.ParseBenchmarkData(opts.Input)

	data, err := InjectTag(data, opts.Tag, opts.GroupPattern, opts.GroupRegex)
	if err != nil {
		return nil, err
	}

	tagDim, err := TagDimension(opts.GroupPattern, opts.GroupRegex)
	if err != nil {
		return nil, err
	}

	bench := shared.NewBenchmark(data)

	if opts.MergeFile != "" {
		existing, err := shared.ReadJSONFile[shared.Benchmark](opts.MergeFile)
		if err == nil {
			var filteredData []shared.BenchmarkData
			for _, d := range existing.Data {
				if getDim(d, tagDim) != opts.Tag {
					filteredData = append(filteredData, d)
				}
			}
			bench.Data = append(filteredData, data...)
			bench.CPU = existing.CPU
			bench.OS = existing.OS
			bench.Arch = existing.Arch
			bench.Runtimes = existing.Runtimes
		} else if !errors.Is(err, os.ErrNotExist) {
			// ignore file-not-found for first run
		}
	}

	if bench.Runtimes == nil {
		bench.Runtimes = make(map[string]time.Time)
	}
	bench.Runtimes[opts.Tag] = opts.Date

	if opts.KeepCount > 0 {
		pruneBenchData(&bench, opts.KeepCount, tagDim)
	}

	return &bench, nil
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

func pruneBenchData(bench *shared.Benchmark, keep int, tagDim string) {
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
		if keepSet[getDim(d, tagDim)] {
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


