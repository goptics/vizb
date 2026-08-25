package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goptics/vizb/internal/flags"
	"github.com/goptics/vizb/pkg/cliout"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FlagBag binds a set of flag descriptors onto a cobra FlagSet, validates them
// (warn-and-default for soft flags, fatal for the rest), and exposes the parsed
// values by name. It is the single binder/validator/reader for every command:
// chart subcommands bind DataFlags + their chart Flags; the root binds DataFlags
// + the shared chart-seed flags. Adding a flag is adding a descriptor — no code
// changes here.
type FlagBag struct {
	flags        []flags.Flag
	strs         map[string]*string
	bools        map[string]*bool
	floats       map[string]*float64
	ints         map[string]*int
	stringSlices map[string]*[]string // backs KindStringSlice, KindStat, and KindStringArray
}

// NewFlagBag allocates a bag and one typed pointer per flag.
func NewFlagBag(fl []flags.Flag) *FlagBag {
	b := &FlagBag{
		flags:        fl,
		strs:         map[string]*string{},
		bools:        map[string]*bool{},
		floats:       map[string]*float64{},
		ints:         map[string]*int{},
		stringSlices: map[string]*[]string{},
	}
	for _, f := range fl {
		switch f.Kind {
		case flags.KindString, flags.KindObject:
			b.strs[f.Name] = new(string)
		case flags.KindBool:
			b.bools[f.Name] = new(bool)
		case flags.KindFloat:
			b.floats[f.Name] = new(float64)
		case flags.KindInt:
			b.ints[f.Name] = new(int)
		case flags.KindStringSlice, flags.KindStat, flags.KindStringArray:
			b.stringSlices[f.Name] = new([]string)
		}
	}
	return b
}

// Bind registers every flag on fs by kind, honouring shorthand and default.
func (b *FlagBag) Bind(fs *pflag.FlagSet) {
	for _, f := range b.flags {
		switch f.Kind {
		case flags.KindString:
			def, _ := f.Default.(string)
			if f.Shorthand != "" {
				fs.StringVarP(b.strs[f.Name], f.Name, f.Shorthand, def, f.Usage)
			} else {
				fs.StringVar(b.strs[f.Name], f.Name, def, f.Usage)
			}
		case flags.KindBool:
			def, _ := f.Default.(bool)
			if f.Shorthand != "" {
				fs.BoolVarP(b.bools[f.Name], f.Name, f.Shorthand, def, f.Usage)
			} else {
				fs.BoolVar(b.bools[f.Name], f.Name, def, f.Usage)
			}
		case flags.KindFloat:
			def, _ := f.Default.(float64)
			if f.Shorthand != "" {
				fs.Float64VarP(b.floats[f.Name], f.Name, f.Shorthand, def, f.Usage)
			} else {
				fs.Float64Var(b.floats[f.Name], f.Name, def, f.Usage)
			}
		case flags.KindInt:
			def, _ := f.Default.(int)
			if f.Shorthand != "" {
				fs.IntVarP(b.ints[f.Name], f.Name, f.Shorthand, def, f.Usage)
			} else {
				fs.IntVar(b.ints[f.Name], f.Name, def, f.Usage)
			}
		case flags.KindStringSlice:
			def, _ := f.Default.([]string)
			if f.Shorthand != "" {
				fs.StringSliceVarP(b.stringSlices[f.Name], f.Name, f.Shorthand, def, f.Usage)
			} else {
				fs.StringSliceVar(b.stringSlices[f.Name], f.Name, def, f.Usage)
			}
		case flags.KindStat:
			sv := &statValue{value: b.stringSlices[f.Name]}
			if f.Shorthand != "" {
				fs.VarP(sv, f.Name, f.Shorthand, f.Usage)
			} else {
				fs.Var(sv, f.Name, f.Usage)
			}
			fs.Lookup(f.Name).NoOptDefVal = statFlagAll
		case flags.KindObject:
			ov := &objectValue{value: b.strs[f.Name]}
			if f.Shorthand != "" {
				fs.VarP(ov, f.Name, f.Shorthand, f.Usage)
			} else {
				fs.Var(ov, f.Name, f.Usage)
			}
			fs.Lookup(f.Name).NoOptDefVal = objectFlagOn
		case flags.KindStringArray:
			if f.Shorthand != "" {
				fs.StringArrayVarP(b.stringSlices[f.Name], f.Name, f.Shorthand, nil, f.Usage)
			} else {
				fs.StringArrayVar(b.stringSlices[f.Name], f.Name, nil, f.Usage)
			}
		}
	}
}

// Validate normalises and validates every flag. Soft flags (data flags + scale/
// sort) warn-and-default via utils.ApplyValidationRules; stat warn-and-defaults
// against the canonical category set. Fatal flags (symbol/symbol-size) error on
// invalid input, but only when the user actually set them.
func (b *FlagBag) Validate(cmd *cobra.Command) {
	for _, f := range b.flags {
		switch {
		case f.Kind == flags.KindStat:
			utils.ApplyValidationRules([]utils.ValidationRule{{
				Label:        "stat",
				SliceValue:   b.stringSlices[f.Name],
				ValidSet:     shared.ValidStatMath,
				Normalizer:   strings.ToLower,
				SliceDefault: nil,
			}})
		case f.Kind == flags.KindObject:
			b.validateObjectFlag(cmd, f)
		case f.Name == "scale":
			b.validateScaleFlag(cmd, f)
		case f.IsSoft():
			b.applySoftRule(f)
		case f.Validate != nil && cmd.Flags().Changed(f.Name):
			if err := f.Validate(b.fatalValue(f)); err != nil {
				shared.ExitWithError(err.Error(), nil)
			}
		}
	}
}

// validateScaleFlag accepts a bare linear|log (warn-and-default via ValidSet)
// or an unwrapped bag. Unknown bag keys are fatal; invalid bases and axis
// names are handled later by EncodeScaleValue (warn-and-default / skip).
func (b *FlagBag) validateScaleFlag(cmd *cobra.Command, f flags.Flag) {
	raw := *b.strs[f.Name]
	if shared.IsScaleBag(raw) {
		if !cmd.Flags().Changed(f.Name) {
			return
		}
		if _, err := shared.ParseObjectBagString(raw, f.ObjectFields); err != nil {
			shared.ExitWithError(fmt.Sprintf("--%s: %v", f.Name, err), nil)
		}
		return
	}
	b.applySoftRule(f)
}

// validateObjectFlag fatals on an invalid object-flag bag (unknown field or a
// field value that fails its validation). The bare sentinel needs no parse.
func (b *FlagBag) validateObjectFlag(cmd *cobra.Command, f flags.Flag) {
	if !cmd.Flags().Changed(f.Name) {
		return
	}
	raw := *b.strs[f.Name]
	if raw == objectFlagOn {
		return
	}
	if _, err := shared.ParseObjectBagString(raw, f.ObjectFields); err != nil {
		shared.ExitWithError(fmt.Sprintf("--%s: %v", f.Name, err), nil)
	}
}

// applySoftRule runs the warn-and-default rule for a soft string/slice flag.
// KindStringArray / KindStringSlice with SoftValidate: normalize each entry and
// warn-and-skip invalid ones (keeps valid siblings). ValidSet-only slices and
// soft string flags still use utils.ApplyValidationRules (warn-and-default).
func (b *FlagBag) applySoftRule(f flags.Flag) {
	if f.Kind == flags.KindStringSlice || f.Kind == flags.KindStringArray {
		if f.SoftValidate != nil {
			b.applySoftSkipEntries(f)
			return
		}
		rule := utils.ValidationRule{
			Label:      f.EffectiveLabel(),
			ValidSet:   f.ValidSet,
			Normalizer: f.Normalizer,
			Validator:  f.SoftValidate,
		}
		rule.SliceValue = b.stringSlices[f.Name]
		rule.SliceDefault, _ = f.Default.([]string)
		utils.ApplyValidationRules([]utils.ValidationRule{rule})
		return
	}
	rule := utils.ValidationRule{
		Label:      f.EffectiveLabel(),
		ValidSet:   f.ValidSet,
		Normalizer: f.Normalizer,
		Validator:  f.SoftValidate,
		Value:      b.strs[f.Name],
	}
	rule.Default, _ = f.Default.(string)
	utils.ApplyValidationRules([]utils.ValidationRule{rule})
}

// applySoftSkipEntries normalizes each string-array/slice entry and drops those
// that fail SoftValidate, warning once per skipped value.
func (b *FlagBag) applySoftSkipEntries(f flags.Flag) {
	slice := b.stringSlices[f.Name]
	if slice == nil {
		return
	}
	kept := make([]string, 0, len(*slice))
	for _, raw := range *slice {
		v := raw
		if f.Normalizer != nil {
			v = f.Normalizer(v)
		}
		if err := f.SoftValidate(v); err != nil {
			msg := fmt.Sprintf("Warning: Invalid %s '%v'. Reason: %s. Skipping", f.EffectiveLabel(), raw, err.Error())
			cliout.Warn(msg)
			continue
		}
		kept = append(kept, v)
	}
	*slice = kept
}

// fatalValue renders the current value of a fatal-validated flag as a string for
// its Validate func.
func (b *FlagBag) fatalValue(f flags.Flag) string {
	switch f.Kind {
	case flags.KindFloat:
		return strconv.FormatFloat(*b.floats[f.Name], 'g', -1, 64)
	case flags.KindInt:
		return strconv.Itoa(*b.ints[f.Name])
	default:
		return *b.strs[f.Name]
	}
}

// String returns the parsed value of a string flag.
func (b *FlagBag) String(name string) string {
	if p := b.strs[name]; p != nil {
		return *p
	}
	return ""
}

// Bool returns the parsed value of a bool flag.
func (b *FlagBag) Bool(name string) bool {
	if p := b.bools[name]; p != nil {
		return *p
	}
	return false
}

// Float returns the parsed value of a float flag.
func (b *FlagBag) Float(name string) float64 {
	if p := b.floats[name]; p != nil {
		return *p
	}
	return 0
}

// Int returns the parsed value of an integer flag.
func (b *FlagBag) Int(name string) int {
	if p := b.ints[name]; p != nil {
		return *p
	}
	return 0
}

// StringSlice returns the parsed value of a slice/stat/array flag.
func (b *FlagBag) StringSlice(name string) []string {
	if p := b.stringSlices[name]; p != nil {
		return *p
	}
	return nil
}

// StringArray returns the parsed value of a repeatable string-array flag.
func (b *FlagBag) StringArray(name string) []string {
	return b.StringSlice(name) // same store; semantics differ only at Bind
}

// StringSliceRef exposes the backing pointer of a slice flag (test helper for
// resetting/forcing values; cobra StringSlice appends to a non-nil default).
func (b *FlagBag) StringSliceRef(name string) *[]string { return b.stringSlices[name] }

// ChartSeed builds the chart-config seed from the chart flags in the bag. A
// chart flag contributes when the user changed it, or when it carries a default
// (e.g. scale → "linear"). Stat and object flags are tri-state: omitted unless
// changed (bare --stat / --bg seeds their on-switch payload).
func (b *FlagBag) ChartSeed(cmd *cobra.Command) map[string]any {
	seed := map[string]any{}
	for _, f := range b.flags {
		if !f.IsChart() {
			continue
		}
		changed := cmd.Flags().Changed(f.Name)
		switch f.Kind {
		case flags.KindStat:
			if !changed {
				continue
			}
			math := *b.stringSlices[f.Name]
			if math == nil {
				math = []string{}
			}
			seed[f.JSONKey] = map[string]any{"enabled": true, "math": math}
		case flags.KindObject:
			if !changed {
				continue
			}
			seed[f.JSONKey] = encodeFlag(f, b.objectBagValue(f))
		case flags.KindBool:
			if changed {
				seed[f.JSONKey] = encodeFlag(f, *b.bools[f.Name])
			}
		case flags.KindFloat:
			if changed {
				seed[f.JSONKey] = encodeFlag(f, *b.floats[f.Name])
			}
		case flags.KindInt:
			if changed {
				seed[f.JSONKey] = encodeFlag(f, *b.ints[f.Name])
			}
		default: // KindString
			if changed || f.Default != nil {
				if f.Name == "scale" {
					seed[f.JSONKey] = shared.EncodeScaleValue(*b.strs[f.Name], f.ObjectFields)
				} else {
					seed[f.JSONKey] = encodeFlag(f, *b.strs[f.Name])
				}
			}
		}
	}
	return seed
}

// Reset restores every bound value to its descriptor default. Test-only helper
// for ResetTestState; production flow re-parses from a fresh process.
func (b *FlagBag) Reset() {
	for _, f := range b.flags {
		switch f.Kind {
		case flags.KindString, flags.KindObject:
			*b.strs[f.Name], _ = f.Default.(string)
		case flags.KindBool:
			*b.bools[f.Name], _ = f.Default.(bool)
		case flags.KindFloat:
			*b.floats[f.Name], _ = f.Default.(float64)
		case flags.KindInt:
			*b.ints[f.Name], _ = f.Default.(int)
		case flags.KindStringSlice, flags.KindStat, flags.KindStringArray:
			def, _ := f.Default.([]string)
			*b.stringSlices[f.Name] = def
		}
	}
}

// ParseConfig assembles the parser.Config from the bag's data flags, mirroring
// the former CommonOptions.ParseConfig.
func (b *FlagBag) ParseConfig() parser.Config {
	cfg, err := parser.ResolveGroupConfig(parser.Config{
		GroupPattern: b.String("group-pattern"),
		GroupRegex:   b.String("group-regex"),
		Group:        b.StringSlice("group"),
		Filter:       b.String("filter"),
		MemUnit:      b.String("mem-unit"),
		TimeUnit:     b.String("time-unit"),
		NumberUnit:   b.String("number-unit"),
		Round:        b.Bool("round"),
	})
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}
	cfg.JSONPath = b.String("json-path")
	cfg.ColAxis = b.String("col-axis")
	selectRaws := b.StringArray("select")
	if len(selectRaws) > 0 {
		if parser.IsExplicitGrouping(cfg) {
			seen := map[string]bool{}
			for _, raw := range selectRaws {
				selected, err := parser.ParseSelectFlag(raw)
				if err != nil {
					shared.ExitWithError(err.Error(), nil)
				}
				for _, col := range selected {
					if seen[col.Source] {
						shared.ExitWithError("duplicate column '"+col.Source+"' in --select", nil)
					}
					seen[col.Source] = true
					cfg.Select = append(cfg.Select, col)
				}
			}
			groupSet := map[string]bool{}
			for _, g := range parser.EffectiveGroupColumns(cfg) {
				groupSet[g] = true
			}
			for _, col := range cfg.Select {
				if groupSet[col.Source] {
					shared.ExitWithError("column '"+col.Source+"' cannot be in both --select and --group", nil)
				}
			}
		} else {
			for _, raw := range selectRaws {
				view, err := parser.ParseSelectViewFlag(raw)
				if err != nil {
					shared.ExitWithError(err.Error(), nil)
				}
				cfg.SelectViews = append(cfg.SelectViews, view)
			}
			if len(cfg.SelectViews) > 1 {
				// 2-col multi-stat or 3-col sankey multi-measure (chart-checked later).
				if err := parser.ValidateRepeatableSelectViews(cfg.SelectViews); err != nil {
					shared.ExitWithError(err.Error(), nil)
				}
			}
		}
	}

	cfg.Mode = parser.ResolveMode(cfg)
	return cfg
}

// Meta builds the pipeline RunMeta from the bag's metadata/parser data flags.
// Theme specs stay as raw strings; the pipeline expands them via style.ResolveThemes.
func (b *FlagBag) Meta() RunMeta {
	return RunMeta{
		ID:          b.String("id"),
		Name:        b.String("name"),
		Title:       b.String("title"),
		ThemeSpecs:  b.StringArray("theme"),
		Description: b.String("description"),
		Tag:         b.String("tag"),
		OutputFile:  b.String("output"),
		Parser:      b.String("parser"),
	}
}

// objectBagValue parses a changed object flag's raw value into its typed bag
// payload: the bare sentinel maps to the empty bag. Validate already rejected
// invalid bags, so a parse error here exits defensively.
func (b *FlagBag) objectBagValue(f flags.Flag) map[string]any {
	raw := *b.strs[f.Name]
	if raw == objectFlagOn {
		return map[string]any{}
	}
	bag, err := shared.ParseObjectBagString(raw, f.ObjectFields)
	if err != nil {
		shared.ExitWithError(fmt.Sprintf("--%s: %v", f.Name, err), nil)
	}
	return bag
}

// encodeFlag applies the flag's payload transform when present.
func encodeFlag(f flags.Flag, v any) any {
	if f.Encode != nil {
		return f.Encode(v)
	}
	return v
}
