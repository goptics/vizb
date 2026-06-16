package cli

import (
	"fmt"
	"strings"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared/utils"
	"github.com/spf13/pflag"
)

// CommonOptions holds the flags every data-consuming command shares. It replaces
// the metadata/parser/grouping/unit fields of the former global shared.FlagState.
type CommonOptions struct {
	Name         string
	Description  string
	Tag          string
	OutputFile   string
	Parser       string
	GroupPattern string
	GroupRegex   string
	Group        []string
	Filter       string
	MemUnit      string
	TimeUnit     string
	NumberUnit   string
}

// Bind registers the common flags onto fs.
func (o *CommonOptions) Bind(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Name, "name", "n", "Comparisons", "Name of the comparison")
	fs.StringVarP(&o.Description, "description", "d", "", "Description of the comparison")
	fs.StringVarP(&o.OutputFile, "output", "o", "", "Output file path/name")
	fs.StringVarP(&o.MemUnit, "mem-unit", "M", "B", "Memory unit available: b, B, KB, MB, GB")
	fs.StringVarP(&o.TimeUnit, "time-unit", "T", "ns", "Time unit available: ns, us, ms, s")
	fs.StringVarP(&o.NumberUnit, "number-unit", "N", "", "Number unit available: K, M, B, T (default: as-is)")
	fs.StringVarP(&o.GroupPattern, "group-pattern", "p", "x", "Pattern to extract grouping information from data labels / series names")
	fs.StringVarP(&o.GroupRegex, "group-regex", "r", "", "Regex pattern to extract grouping information from data labels / series names")
	fs.StringSliceVarP(&o.Group, "group", "g", nil, "Names each dimension in --group-pattern/regex order. csv/json: column/field names whose values feed the dimensions; benchmark parsers: human-readable labels for the name/x/y/z axes")
	fs.StringVarP(&o.Filter, "filter", "f", "", "Regex pattern to include only matching data labels / series names")
	fs.StringVarP(&o.Tag, "tag", "t", "", "Tag/identifier for the comparison")
	fs.StringVarP(&o.Parser, "parser", "P", "auto", "Benchmark parser to use; 'auto' detects from input content (one of: auto, "+strings.Join(parser.AvailableParsers(), ", ")+")")
}

// validationRules returns the warn-and-default rules for the common fields,
// mirroring the former cmd/flag_validation_rules.go set.
func (o *CommonOptions) validationRules() []utils.ValidationRule {
	return []utils.ValidationRule{
		{
			Label:    "memory unit",
			Value:    &o.MemUnit,
			ValidSet: []string{"b", "B", "KB", "MB", "GB"},
			Normalizer: func(s string) string {
				mapping := map[string]string{"kb": "KB", "mb": "MB", "gb": "GB"}
				if val, ok := mapping[s]; ok {
					return val
				}
				return s
			},
			Default: "B",
		},
		{
			Label:    "time unit",
			Value:    &o.TimeUnit,
			ValidSet: []string{"ns", "us", "ms", "s"},
			Default:  "ns",
		},
		{
			Label:      "number unit",
			Value:      &o.NumberUnit,
			ValidSet:   []string{"K", "M", "B", "T"},
			Normalizer: strings.ToUpper,
			Default:    "",
		},
		{
			Label:     "group pattern",
			Value:     &o.GroupPattern,
			Validator: parser.ValidateGroupPattern,
			Default:   "xAxis",
		},
		{
			Label:     "parser",
			Value:     &o.Parser,
			Validator: validateParser,
			Default:   "auto",
		},
	}
}

// Validate normalises and validates the common flags (warn-and-default, never fatal).
func (o *CommonOptions) Validate() {
	utils.ApplyValidationRules(o.validationRules())
}

// ParseConfig converts the common flags into the parser.Config that parsers consume.
func (o *CommonOptions) ParseConfig() parser.Config {
	return parser.Config{
		GroupPattern: o.GroupPattern,
		GroupRegex:   o.GroupRegex,
		Group:        o.Group,
		Filter:       o.Filter,
		MemUnit:      o.MemUnit,
		TimeUnit:     o.TimeUnit,
		NumberUnit:   o.NumberUnit,
	}
}

// LinearOptions adds the dataset-level sort/labels flags shared by linear data
// commands (root and every linear chart subcommand).
type LinearOptions struct {
	CommonOptions
	Sort       string
	ShowLabels bool
}

// Bind registers the common flags plus --sort/--show-labels.
func (o *LinearOptions) Bind(fs *pflag.FlagSet) {
	o.CommonOptions.Bind(fs)
	fs.StringVarP(&o.Sort, "sort", "s", "", "Sort in asc or desc order (default: as-is)")
	fs.BoolVarP(&o.ShowLabels, "show-labels", "l", false, "Show labels on charts")
}

func (o *LinearOptions) validationRules() []utils.ValidationRule {
	return append(o.CommonOptions.validationRules(), utils.ValidationRule{
		Label:      "sort order",
		Value:      &o.Sort,
		ValidSet:   []string{"asc", "desc"},
		Normalizer: strings.ToLower,
		Default:    "",
	})
}

// Validate normalises and validates the linear flags.
func (o *LinearOptions) Validate() {
	utils.ApplyValidationRules(o.validationRules())
}

// ChartOptions is the base every chart subcommand embeds: the linear data flags
// plus --swap (a valid axis-permutation override for all chart types).
type ChartOptions struct {
	LinearOptions
	Swap string
}

// Bind registers the linear flags plus --swap.
func (o *ChartOptions) Bind(fs *pflag.FlagSet) {
	o.LinearOptions.Bind(fs)
	fs.StringVar(&o.Swap, "swap", "", "Axis permutation override, e.g. yx, yxn (chars from name/x/y/z)")
}

func validateParser(key string) error {
	if key == "auto" {
		return nil
	}
	if _, ok := parser.Parsers[key]; !ok {
		return fmt.Errorf("unknown parser '%s'; available: auto, %v", key, parser.AvailableParsers())
	}
	return nil
}
