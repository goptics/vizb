package cli

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
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
	Group        []string
	GroupPattern string
	GroupRegex   string
	Filter       string
	MemUnit      string
	TimeUnit     string
	NumberUnit   string
	Select       string
	Axes         string
}

// Bind registers the common flags onto fs.
func (o *CommonOptions) Bind(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Name, "name", "n", "Comparisons", "Name of the comparison")
	fs.StringVarP(&o.Description, "description", "d", "", "Description of the comparison")
	fs.StringVarP(&o.OutputFile, "output", "o", "", "Output file path/name")
	fs.StringVarP(&o.MemUnit, "mem-unit", "M", "B", "Memory unit available: b, B, KB, MB, GB")
	fs.StringVarP(&o.TimeUnit, "time-unit", "T", "ns", "Time unit available: ns, us, ms, s")
	fs.StringVarP(&o.NumberUnit, "number-unit", "N", "", "Number unit available: K, M, B, T (default: as-is)")
	fs.StringVarP(&o.GroupPattern, "group-pattern", "p", "x", "Pattern to extract grouping information from data labels / series names; CSV/JSON: bracket slots [x-y-n] split a column value; {label} sets axis titles (e.g. -p '[n{year}-y{months}],z{category}')")
	fs.StringVarP(&o.GroupRegex, "group-regex", "r", "", "Regex pattern to extract grouping information from data labels / series names")
	fs.StringSliceVarP(&o.Group, "group", "g", nil, "Names dimensions in --group-pattern order; use the same separators as -p (e.g. -g \"name category region\" -p \"x n y\", or -g name,category/region -p x,y/z). csv/json: column/field names; benchmark parsers: axis labels")
	fs.StringVarP(&o.Filter, "filter", "f", "", "Regex pattern to include only matching data labels / series names")
	fs.StringVarP(&o.Tag, "tag", "t", "", "Tag/identifier for the comparison")
	fs.StringVarP(&o.Parser, "parser", "P", "auto", "Benchmark parser to use; 'auto' detects from input content (one of: auto, "+strings.Join(parser.AvailableParsers(), ", ")+")")
	fs.StringVar(&o.Select, "select", "", "csv/json only: select value columns; optional rename with {label} (e.g. --select=price{Unit price},count)")
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
	cfg, err := parser.ResolveGroupConfig(parser.Config{
		GroupPattern: o.GroupPattern,
		GroupRegex:   o.GroupRegex,
		Group:        o.Group,
		Filter:       o.Filter,
		MemUnit:      o.MemUnit,
		TimeUnit:     o.TimeUnit,
		NumberUnit:   o.NumberUnit,
	})
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}
	if strings.TrimSpace(o.Select) != "" {
		selected, err := parser.ParseSelectFlag(o.Select)
		if err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
		cfg.Select = selected
		groupSet := map[string]bool{}
		for _, g := range parser.EffectiveGroupColumns(cfg) {
			groupSet[g] = true
		}
		for _, col := range selected {
			if groupSet[col.Source] {
				shared.ExitWithError(fmt.Sprintf("column '%s' cannot be in both --select and --group", col.Source), nil)
			}
		}
	}
	if strings.TrimSpace(o.Axes) != "" {
		if len(o.Group) > 0 || strings.TrimSpace(o.GroupRegex) != "" {
			shared.ExitWithError("--axes cannot be combined with --group or --group-regex", nil)
		}
		if strings.TrimSpace(o.Select) != "" {
			shared.ExitWithError("--axes cannot be combined with --select", nil)
		}
		axes, err := parser.ParseAxesFlag(o.Axes)
		if err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
		cfg.Axes = axes
	}
	return cfg
}

// LinearOptions adds the dataset-level sort/labels flags shared by linear data
// commands (root and every linear chart subcommand).
type LinearOptions struct {
	CommonOptions
	Sort       string
	ShowLabels bool
	Stat       []string
}

// Bind registers the common flags plus --sort/--show-labels/--stat.
func (o *LinearOptions) Bind(fs *pflag.FlagSet) {
	o.CommonOptions.Bind(fs)
	fs.StringVarP(&o.Sort, "sort", "s", "", "Sort in asc or desc order (default: as-is)")
	fs.BoolVarP(&o.ShowLabels, "show-labels", "l", false, "Show labels on charts")
	fs.Var(&statFlag{value: &o.Stat}, "stat", "Enable stats panel; omit to disable, use alone for all categories, or =cat1,cat2 for specific (counts, center, spread, extremes, shape, percentiles, confidence, correlations)")
	fs.Lookup("stat").NoOptDefVal = statFlagAll
}

func (o *LinearOptions) validationRules() []utils.ValidationRule {
	return append(o.CommonOptions.validationRules(),
		utils.ValidationRule{
			Label:      "sort order",
			Value:      &o.Sort,
			ValidSet:   []string{"asc", "desc"},
			Normalizer: strings.ToLower,
			Default:    "",
		},
		utils.ValidationRule{
			Label:        "stat",
			SliceValue:   &o.Stat,
			ValidSet:     shared.ValidStatMath,
			Normalizer:   strings.ToLower,
			SliceDefault: nil,
		},
	)
}

// Validate normalises and validates the linear flags.
func (o *LinearOptions) Validate() {
	utils.ApplyValidationRules(o.validationRules())
}

// BindStatFlag registers --stat on fs, pointing at target. Exported so commands
// that don't embed LinearOptions can still register the same flag.
func BindStatFlag(fs *pflag.FlagSet, target *[]string) {
	fs.Var(&statFlag{value: target}, "stat", "Enable stats panel; omit to disable, use alone for all categories, or =cat1,cat2 for specific (counts, center, spread, extremes, shape, percentiles, confidence, correlations)")
	fs.Lookup("stat").NoOptDefVal = statFlagAll
}

// RewriteStatArg rewrites --stat VALUE (space-separated) to --stat=VALUE so
// pflag can parse it correctly despite the NoOptDefVal. Without this rewrite,
// pflag consumes the NoOptDefVal "all" and treats VALUE as a positional arg.
func RewriteStatArg(args []string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--stat" && i+1 < len(args) && looksLikeStatValue(args[i+1]) {
			if _, err := os.Stat(args[i+1]); err != nil {
				out = append(out, "--stat="+args[i+1])
				i++
				continue
			}
		} else {
			out = append(out, args[i])
		}
	}
	return out
}

// looksLikeStatValue reports whether s could be an argument to --stat.
// Returns false for anything starting with '-' (another flag) or not composed
// entirely of recognised stat category names (or the "all" sentinel).
func looksLikeStatValue(s string) bool {
	if strings.HasPrefix(s, "-") || s == "" {
		return false
	}
	lower := strings.ToLower(s)
	if lower == statFlagAll {
		return true
	}
	for _, part := range strings.Split(lower, ",") {
		if !slices.Contains(shared.ValidStatMath, strings.TrimSpace(part)) {
			return false
		}
	}
	return true
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

// statFlagAll is the NoOptDefVal sentinel: pflag requires a non-empty string
// for optional-value flags. Set() converts it to []string{} (all categories).
// "all" is also accepted as an explicit value so --stat=all works identically.
const statFlagAll = "all"

// statFlag is a pflag.Value that makes --stat optional-value:
//
//	--stat          → all categories ([]string{})
//	--stat=a,b      → specific categories
//	(omitted)       → nil = disabled
type statFlag struct{ value *[]string }

func (f *statFlag) String() string {
	if *f.value == nil {
		return ""
	}
	return strings.Join(*f.value, ",")
}

func (f *statFlag) Set(val string) error {
	if val == statFlagAll {
		*f.value = []string{}
		return nil
	}
	*f.value = strings.Split(val, ",")
	return nil
}

func (f *statFlag) Type() string { return "string" }

func validateParser(key string) error {
	if key == "auto" {
		return nil
	}
	if _, ok := parser.Parsers[key]; !ok {
		return fmt.Errorf("unknown parser '%s'; available: auto, %v", key, parser.AvailableParsers())
	}
	return nil
}
