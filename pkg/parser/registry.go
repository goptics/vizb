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
	Select          []ColumnSpec
	Axes            []ColumnSpec // --axes value mode: numeric cols placed on x,y[,z]
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
