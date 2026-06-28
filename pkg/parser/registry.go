package parser

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/goptics/vizb/shared"
)

// Config carries the parse-time settings every parser needs. It replaces the
// former global shared.FlagState reads, making parsers pure functions of
// (input, config).
type Config struct {
	GroupPattern    string
	GroupRegex      string
	Group           []string
	GroupColumns    []string
	LabelSeparators []string
	GroupStructured bool
	TabularPattern  *TabularPattern
	Filter          string
	MemUnit         string
	TimeUnit        string
	NumberUnit      string
	Select          []ColumnSpec // grouped mode: numeric stat columns
	SelectViews     []SelectView // solo axis mode: one entry per --select occurrence
	Axes            []ColumnSpec // auto-value mode: numeric cols placed on x,y[,z]
	MetricColumn    string       // auto-value: 4th numeric col → visualMap metric
	JSONPath        string       // json only: jq-like dot path to the nested array to chart
	AutoGroup       bool         // csv/json: infer group columns when no explicit grouping is configured
	ChartTypes      []string     // csv/json auto-value eligibility check (scatter/bar/line only)
}

// SelectView is one solo --select flag: column placement plus an optional
// chart-tab name from a trailing (Title) suffix in multi-stat mode.
type SelectView struct {
	Columns   []ColumnSpec
	TypeLabel string
}

type ParseFunc func(filename string, cfg Config) []shared.DataPoint

var Parsers = map[string]ParseFunc{}

func GetParser(key string) (ParseFunc, error) {
	fn, ok := Parsers[key]
	if !ok {
		return nil, fmt.Errorf("unknown parser '%s'; available parsers: %v", key, AvailableParsers())
	}
	return fn, nil
}

func AvailableParsers() []string {
	keys := make([]string, 0, len(Parsers))
	for k := range Parsers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func ShouldIncludeBenchmark(benchName string, cfg Config) bool {
	if cfg.Filter == "" {
		return true
	}

	filterRe, err := regexp.Compile(cfg.Filter)
	if err != nil {
		shared.ExitWithError("Invalid filter regex", err)
	}

	return filterRe.MatchString(benchName)
}
