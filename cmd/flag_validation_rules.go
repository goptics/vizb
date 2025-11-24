package cmd

import (
	"strings"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
)

var flagValidationRules = []utils.ValidationRule{
	{
		Label:    "memory unit",
		Value:    &shared.FlagState.MemUnit,
		ValidSet: []string{"b", "B", "kb", "mb", "gb"},
		Normalizer: func(s string) string {
			// Skip normalization for B (Byte)
			if s == "B" {
				return s
			}

			return strings.ToLower(s)
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
		Label:      "format",
		Value:      &shared.FlagState.Format,
		ValidSet:   []string{"html", "json"},
		Normalizer: strings.ToLower,
		Default:    "html",
	},
	{
		Label: "group pattern",
		Value: &shared.FlagState.GroupPattern,
		Validator: func(pattern string) bool {
			if err := parser.ValidatePattern(pattern); err != nil {
				return false
			}

			return true
		},
		Default: "xAxis",
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
}
