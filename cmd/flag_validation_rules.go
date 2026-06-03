package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
)

var flagValidationRules = []utils.ValidationRule{
	{
		Label:    "memory unit",
		Value:    &shared.FlagState.MemUnit,
		ValidSet: []string{"b", "B", "KB", "MB", "GB"},
		Normalizer: func(s string) string {
			mapping := map[string]string{
				"kb": "KB",
				"mb": "MB",
				"gb": "GB",
			}

			if val, ok := mapping[s]; ok {
				return val
			}

			return s
		},
		Default: "B",
	},
	{
		Label:      "time unit",
		Value:      &shared.FlagState.TimeUnit,
		ValidSet:   []string{"ns", "us", "ms", "s"},
		Normalizer: nil,
		Default:    "ns",
	},
	{
		Label:      "number unit",
		Value:      &shared.FlagState.NumberUnit,
		ValidSet:   []string{"K", "M", "B", "T"},
		Normalizer: strings.ToUpper,
		Default:    "",
	},
	{
		Label:     "group pattern",
		Value:     &shared.FlagState.GroupPattern,
		Validator: parser.ValidateGroupPattern,
		Default:   "xAxis",
	},
	{
		Label:      "sort order",
		Value:      &shared.FlagState.Sort,
		ValidSet:   []string{"asc", "desc"},
		Normalizer: strings.ToLower,
		Default:    "",
	},
	{
		Label:        "charts",
		SliceValue:   &shared.FlagState.Charts,
		ValidSet:     []string{"bar", "line", "pie"},
		Normalizer:   strings.ToLower,
		SliceDefault: []string{"bar", "line", "pie"},
	},
	{
		Label:      "inject dimension",
		Value:      &shared.FlagState.TagAxis,
		ValidSet:   []string{"n", "x", "y"},
		Normalizer: strings.ToLower,
		Default:    "n",
	},
	{
		Label:     "parser",
		Value:     &shared.FlagState.Parser,
		Validator: validateParser,
		Default:   "go",
	},
	{
		Label:     "data url",
		Value:     &shared.FlagState.DataURL,
		Validator: validateAPIURL,
		Default:   "",
	},
}

func validateParser(key string) error {
	if _, ok := parser.Parsers[key]; !ok {
		return fmt.Errorf("unknown parser '%s'; available: %v", key, parser.AvailableParsers())
	}
	return nil
}

func validateAPIURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("must be a valid http:// or https:// URL")
	}
	return nil
}
