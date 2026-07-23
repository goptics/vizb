package cli

import (
	"fmt"
	"strings"

	"github.com/goptics/vizb/internal/flags"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/pkg/style"
)

// DataFlags are the parser/grouping/metadata descriptors every data-consuming
// command carries (root and every chart subcommand). None has a JSONKey: their
// values feed parser.Config and dataset metadata (read back by name), never the
// chart seed. Soft flags warn-and-default; they are never fatal. This replaces
// the former CommonOptions hand-written Bind + validationRules.
var DataFlags = []flags.Flag{
	{Name: "name", Shorthand: "n", Default: "Comparisons", Usage: "Name of the comparison", Kind: flags.KindString},
	{Name: "title", Usage: "Override the chart title when --col-axis produces one chart; independent of -n/--name, otherwise ignored", Kind: flags.KindString},
	{Name: "theme", Usage: "Initial series color theme (a built-in name or comma-separated hex palette)", Kind: flags.KindString, Default: "default", Normalizer: style.NormalizeTheme, SoftValidate: style.ValidateTheme},
	{Name: "description", Shorthand: "d", Usage: "Description of the comparison", Kind: flags.KindString},
	{Name: "output", Shorthand: "o", Usage: "Output file path/name", Kind: flags.KindString},
	{Name: "tag", Shorthand: "t", Usage: "Tag/identifier for the comparison", Kind: flags.KindString},
	{Name: "id", Usage: "Dataset id for ?id= deep links", Kind: flags.KindString},
	{
		Name: "parser", Shorthand: "P", Default: "auto", Kind: flags.KindString,
		Usage:        "Benchmark parser to use; 'auto' detects from input content (one of: auto, " + strings.Join(parser.AvailableParsers(), ", ") + ")",
		Label:        "parser",
		SoftValidate: validateParser,
	},
	{
		Name: "group-pattern", Shorthand: "p", Default: "x", Kind: flags.KindString,
		Usage:        "Pattern to extract grouping information from data labels / series names; CSV/JSON: bracket slots [x-y-n] split a column value; {label} sets axis titles (e.g. -p '[n{year}-y{months}],z{category}')",
		Label:        "group pattern",
		SoftValidate: parser.ValidateGroupPattern,
	},
	{Name: "group-regex", Shorthand: "r", Usage: "Regex pattern to extract grouping information from data labels / series names", Kind: flags.KindString},
	{Name: "group", Shorthand: "g", Usage: "Names dimensions in --group-pattern order; use the same separators as -p (e.g. -g \"name category region\" -p \"x n y\", or -g name,category/region -p x,y/z). csv/json: column/field names; benchmark parsers: axis labels", Kind: flags.KindStringSlice},
	{Name: "filter", Shorthand: "f", Usage: "Regex to include only matching rows (CSV/JSON: --group label) or benchmark names", Kind: flags.KindString},
	{
		Name: "mem-unit", Shorthand: "M", Default: "B", Kind: flags.KindString,
		Usage:      "Memory unit available: b, B, KB, MB, GB",
		Label:      "memory unit",
		ValidSet:   []string{"b", "B", "KB", "MB", "GB"},
		Normalizer: normalizeMemUnit,
	},
	{
		Name: "time-unit", Shorthand: "T", Default: "ns", Kind: flags.KindString,
		Usage:    "Time unit available: ns, us, ms, s",
		Label:    "time unit",
		ValidSet: []string{"ns", "us", "ms", "s"},
	},
	{
		Name: "number-unit", Shorthand: "N", Kind: flags.KindString,
		Usage:      "Number unit available: K, M, B, T (default: as-is)",
		Label:      "number unit",
		ValidSet:   []string{"K", "M", "B", "T"},
		Normalizer: strings.ToUpper,
	},
	{Name: "select", Usage: "csv/json only: select columns (repeatable); solo mode: 2–4 cols as x,y[,z][,metric] (e.g. --select x,y,z,value); grouped mode: numeric stat columns with optional {label}", Kind: flags.KindStringArray},
	{
		Name: "col-axis", Shorthand: "A", Kind: flags.KindString,
		Usage:    "csv/json: place numeric column names on this axis (n, x, y, or z) so all columns share one chart; works without -g; only numeric columns become series",
		Label:    "col-axis",
		ValidSet: []string{"n", "x", "y", "z"},
	},
	{Name: "json-path", Usage: "json only: select a nested array to chart via a jq-like dot path (e.g. --json-path '.data.results')", Kind: flags.KindString},
}

// normalizeMemUnit canonicalises lowercase memory units (kb/mb/gb) to their
// upper form so validation against the valid set passes.
func normalizeMemUnit(s string) string {
	switch s {
	case "kb":
		return "KB"
	case "mb":
		return "MB"
	case "gb":
		return "GB"
	}
	return s
}

// validateParser reports whether key names a registered parser (or "auto").
func validateParser(key string) error {
	if key == "auto" {
		return nil
	}
	if _, ok := parser.Parsers[key]; !ok {
		return fmt.Errorf("unknown parser '%s'; available: auto, %v", key, parser.AvailableParsers())
	}
	return nil
}
