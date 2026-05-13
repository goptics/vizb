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

type Dimension string

const (
	DimName  Dimension = "name"
	DimXAxis Dimension = "xAxis"
	DimYAxis Dimension = "yAxis"
)

var allDimensions = []Dimension{DimName, DimXAxis, DimYAxis}

type ActionOpts struct {
	Input         string
	IdentifyValue string
	Date          time.Time
	AppendFile    string
	KeepCount     int
	GroupPattern  string
	GroupRegex    string
}

func RunAction(opts ActionOpts) (*shared.Benchmark, error) {
	if _, err := os.Stat(opts.Input); err != nil {
		return nil, fmt.Errorf("input file: %w", err)
	}

	tagDim, err := TagDimension(opts.GroupPattern, opts.GroupRegex)
	if err != nil {
		return nil, err
	}

	shared.FlagState.GroupPattern = opts.GroupPattern
	shared.FlagState.GroupRegex = opts.GroupRegex

	data := parser.ParseBenchmarkData(opts.Input)
	data = InjectTag(data, opts.IdentifyValue, tagDim)
	bench := shared.NewBenchmark(data)

	if err := appendExistingRuns(&bench, opts.AppendFile, opts.IdentifyValue, tagDim); err != nil {
		return nil, err
	}

	if bench.Runtimes == nil {
		bench.Runtimes = make(map[string]time.Time)
	}
	bench.Runtimes[opts.IdentifyValue] = opts.Date

	if opts.KeepCount > 0 {
		pruneBenchData(&bench, opts.KeepCount, tagDim)
	}

	return &bench, nil
}

func parsePresentDimensions(pattern, regexStr string) ([]Dimension, error) {
	if regexStr != "" {
		re, err := regexp.Compile(regexStr)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %w", err)
		}
		var present []Dimension
		for _, name := range re.SubexpNames() {
			if name == "" {
				continue
			}
			present = append(present, expandCIShorthand(name))
		}
		return present, nil
	}

	if err := validateCIPattern(pattern); err != nil {
		return nil, err
	}
	parts := parser.ParsePatternParts(pattern)
	dims := make([]Dimension, len(parts))
	for i, p := range parts {
		dims[i] = Dimension(p)
	}
	return dims, nil
}

func TagDimension(pattern, regexStr string) (Dimension, error) {
	present, err := parsePresentDimensions(pattern, regexStr)
	if err != nil {
		return "", err
	}

	presentSet := make(map[Dimension]bool, len(present))
	for _, d := range present {
		presentSet[d] = true
	}

	for _, d := range allDimensions {
		if !presentSet[d] {
			return d, nil
		}
	}

	return "", fmt.Errorf("CI pattern must define exactly 2 dimensions, got %d", len(presentSet))
}

func InjectTag(data []shared.BenchmarkData, tag string, tagDim Dimension) []shared.BenchmarkData {
	if tag == "" {
		return data
	}

	result := make([]shared.BenchmarkData, len(data))
	for i, d := range data {
		setDim(&d, tagDim, tag)
		result[i] = d
	}
	return result
}

func appendExistingRuns(bench *shared.Benchmark, appendFile, identifyValue string, tagDim Dimension) error {
	if appendFile == "" {
		return nil
	}

	existing, err := shared.ReadJSONFile[shared.Benchmark](appendFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read existing benchmarks: %w", err)
	}

	for _, d := range existing.Data {
		if getDim(d, tagDim) != identifyValue {
			bench.Data = append(bench.Data, d)
		}
	}

	bench.CPU = existing.CPU
	bench.OS = existing.OS
	bench.Arch = existing.Arch

	if bench.Runtimes == nil && existing.Runtimes != nil {
		bench.Runtimes = make(map[string]time.Time, len(existing.Runtimes))
	}
	for k, v := range existing.Runtimes {
		if bench.Runtimes == nil {
			bench.Runtimes = make(map[string]time.Time)
		}
		bench.Runtimes[k] = v
	}

	return nil
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

func expandCIShorthand(part string) Dimension {
	shortcuts := map[string]Dimension{"n": DimName, "x": DimXAxis, "y": DimYAxis}
	if expanded, exists := shortcuts[part]; exists {
		return expanded
	}
	return Dimension(part)
}

func setDim(d *shared.BenchmarkData, dim Dimension, value string) {
	switch dim {
	case DimName:
		d.Name = value
	case DimXAxis:
		d.XAxis = value
	case DimYAxis:
		d.YAxis = value
	}
}

func getDim(d shared.BenchmarkData, dim Dimension) string {
	switch dim {
	case DimName:
		return d.Name
	case DimXAxis:
		return d.XAxis
	case DimYAxis:
		return d.YAxis
	}
	return ""
}

func pruneBenchData(bench *shared.Benchmark, keep int, tagDim Dimension) {
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
