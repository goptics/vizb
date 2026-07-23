package parser

import (
	"fmt"
	"io"
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
	Axes            []ColumnSpec // solo --select / --axes value mode: numeric cols on x,y[,z]
	MetricColumn    string       // solo --select / --axes value mode: 4th numeric col → visualMap metric
	JSONPath        string       // json only: jq-like dot path to the nested array to chart
	AutoGroup       bool         // csv/json: infer group columns / col-axis when no explicit grouping is configured
	ChartTypes      []string     // chart types for the current run (used by chart-aware inference)
	Mode            Mode         // resolved once in ParseConfig so downstream switches on cfg.Mode
	ColAxis         string       // csv/json: place numeric column names on this axis (n/x/y/z); empty = one chart per column
	QuietAutoDetect bool         // suppress csv/json auto-detection notices for request-scoped callers
}

// Mode is the resolved parse mode for a Config. Set once in ParseConfig so
// every downstream call site switches on cfg.Mode instead of re-deriving it
// from overlapping predicates.
type Mode int

const (
	ModeAuto      Mode = iota // no --select and no explicit grouping → auto-group / auto col-axis
	ModeGrouped               // explicit grouping (-g/-r/-p) + --select numeric stat columns
	ModeValue                 // solo --select, all numeric columns → value axes x,y[,z]
	ModeMixed                 // solo --select, one categorical x + numeric y[,z]
	ModeMultiStat             // repeatable solo --select (dim,metric) pairs merged into stats
)

// SelectView is one solo --select flag: column placement plus an optional
// chart-tab name from a trailing (Title) suffix in multi-stat mode.
type SelectView struct {
	Columns   []ColumnSpec
	TypeLabel string
}

// ParseFunc parses request-local input and returns data points, the effective
// config, and any system metadata discovered in the input.
type ParseFunc func(io.Reader, Config) ([]shared.DataPoint, Config, *shared.Meta, error)

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

func ShouldIncludeBenchmark(benchName string, cfg Config) (bool, error) {
	if cfg.Filter == "" {
		return true, nil
	}

	filterRe, err := regexp.Compile(cfg.Filter)
	if err != nil {
		return false, fmt.Errorf("invalid filter regex: %w", err)
	}

	return filterRe.MatchString(benchName), nil
}
