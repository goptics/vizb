package cmd

import (
	"strings"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
)

var flagValidationRules = []utils.ValidationRule{
	{
		Label:      "memory unit",
		Value:      &shared.FlagState.MemUnit,
		ValidSet:   []string{"b", "kb", "mb", "gb"},
		Normalizer: strings.ToLower,
		Default:    "b",
	},
	{
		Label:      "time unit",
		Value:      &shared.FlagState.TimeUnit,
		ValidSet:   []string{"ns", "us", "ms", "s"},
		Normalizer: nil,
		Default:    "ns",
	},
	{
		Label:      "allocation unit",
		Value:      &shared.FlagState.AllocUnit,
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
		Default: "subject",
	},
}
