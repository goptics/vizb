package cmd

import (
	"strings"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
)

var sharedFlagValidationRules = []utils.ValidationRule{
	{
		Label:    "memory unit",
		Value:    &shared.BenchSettings.MemUnit,
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
		Value:      &shared.BenchSettings.TimeUnit,
		ValidSet:   []string{"ns", "us", "ms", "s"},
		Normalizer: nil,
		Default:    "ns",
	},
	{
		Label:      "number unit",
		Value:      &shared.BenchSettings.NumberUnit,
		ValidSet:   []string{"K", "M", "B", "T"},
		Normalizer: strings.ToUpper,
		Default:    "",
	},
	{
		Label:      "sort order",
		Value:      &shared.BenchSettings.Sort,
		ValidSet:   []string{"asc", "desc"},
		Normalizer: strings.ToLower,
		Default:    "",
	},
	{
		Label:        "charts",
		SliceValue:   &shared.BenchSettings.Charts,
		ValidSet:     []string{"bar", "line", "pie"},
		Normalizer:   strings.ToLower,
		SliceDefault: []string{"bar", "line", "pie"},
	},
}

var flagValidationRules = append(sharedFlagValidationRules, utils.ValidationRule{
	Label:     "group pattern",
	Value:     &shared.FlagState.GroupPattern,
	Validator: parser.ValidateGroupPattern,
	Default:   "xAxis",
})
