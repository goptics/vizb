package parser

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/goptics/vizb/shared"
)

type ParseFunc func(filename string) []shared.DataPoint

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

func ShouldIncludeBenchmark(benchName string) bool {
	if shared.FlagState.FilterRegex == "" {
		return true
	}

	filterRe, err := regexp.Compile(shared.FlagState.FilterRegex)
	if err != nil {
		shared.ExitWithError("Invalid filter regex", err)
	}

	return filterRe.MatchString(benchName)
}
