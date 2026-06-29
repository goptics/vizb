// Package flags defines the single flag-descriptor type shared by every vizb
// command. One Flag declares a CLI option once; the cobra binder, the --chart
// override parser, and the chart-seed/parser-config builders all consume the
// same descriptor. Adding a flag means adding one Flag — nothing else.
//
// A Flag also carries zero or more applicability rules (Rule) governing whether
// its value applies to the runtime dataset/context; rule evaluation and outcome
// reduction live in config/charts (this package stays opaque to the context).
//
// This package is a stdlib-only leaf: it must never import shared, config/charts,
// pkg/parser, or cobra. shared imports config/charts, so the descriptor type
// cannot live in shared; keeping it dependency-free lets both config/charts
// (chart descriptors) and cmd/cli (data/metadata descriptors) use it with no
// import cycle. Validators that need other packages (parser.ValidateGroupPattern,
// shared.ValidStatMath) are injected at the descriptor's definition site, never
// referenced here.
package flags

// Kind is the value type of a Flag. It drives how the cobra binder registers the
// flag and how the --chart override parser converts a raw token into a payload
// value.
type Kind int

const (
	KindString      Kind = iota // string
	KindBool                    // bool
	KindFloat                   // float64
	KindStringSlice             // []string (comma-separated), e.g. --group, --charts
	KindStringArray             // []string (repeatable flag), e.g. --select, --chart
	KindStat                    // optional-value string slice, e.g. --stat
)

// Outcome is the result of evaluating one applicability rule. Multiple rules
// on a flag are reduced by worst-outcome wins (precedence Fatal > Skip >
// WarnKeep > Keep). The reduction lives in config/charts.ApplyRules (Phase B);
// this package only declares the vocabulary and the zero value (Keep).
type Outcome int8

const (
	Keep     Outcome = iota // apply the flag's value normally
	WarnKeep                // warn to stderr, keep the value
	Skip                    // warn, drop the flag's contribution
	Fatal                   // CLI error, exit non-zero
)

// RuleFn evaluates whether a flag applies to the runtime context. It is opaque
// (ctx any) so config/flags stays stdlib-only: the concrete RuleContext lives
// in config/charts (which imports shared and parser). Returns the outcome and
// a human-readable message used in warnings/errors.
type RuleFn func(ctx any) (Outcome, string)

// Flag is the single source of truth for one CLI option.
//
// A descriptor is one of three species:
//
//   - chart flag: JSONKey is set. Its value is encoded into the chart-config
//     seed (and decoded into a typed chart Config). Validation, when present, is
//     fatal via Validate — unless the soft trio is set (e.g. scale), in which
//     case it is warn-and-default AND still seeded.
//   - data flag: JSONKey is empty. Its value feeds parser.Config / dataset
//     metadata, read back by name. Validation is warn-and-default via the soft
//     trio (ValidSet / Normalizer / SoftValidate), never fatal.
//   - applicability rules: the optional Rule list decides at runtime whether the
//     flag's value even applies to the current dataset/context; precedence is
//     Fatal > Skip > WarnKeep > Keep (worst wins), reduced in config/charts.
type Flag struct {
	Name      string // CLI flag name (cobra long flag), e.g. "scale", "mem-unit", "show-labels"
	Shorthand string // optional cobra shorthand, e.g. "M"
	Usage     string // help text
	Kind      Kind   // value type
	Default   any    // bind default; also the warn-and-default fallback value
	Key       string // --chart override key when it differs from Name (e.g. "labels" for --show-labels); empty = use Name

	// --- chart-payload application (set ⇒ this flag contributes to the seed) ---
	JSONKey string        // chart Config json tag the value maps to, e.g. "threeDRotate"
	Encode  func(any) any // transform the Kind-converted value into its payload shape; nil = identity

	// --- fatal validation (chart flags): invalid input ⇒ error ---
	Validate func(string) error // context-free; nil = none (swap is validated against axes by the caller)

	// --- warn-and-default validation (data flags, and scale): invalid ⇒ warn + reset to Default ---
	Label        string              // human label in warn-and-default messages ("memory unit"); empty = use Name
	ValidSet     []string            // allowed values (post-normalization)
	Normalizer   func(string) string // canonicalize before checking ValidSet, e.g. strings.ToLower
	SoftValidate func(string) error  // custom non-fatal validator, e.g. parser.ValidateGroupPattern

	// --- applicability rules: 0+ rules; all evaluated; worst outcome wins ---
	// Precedence: Fatal > Skip > WarnKeep > Keep. Nil ⇒ always Keep (flag is
	// unconditionally applicable). Reduced by config/charts.ApplyRules (Phase B);
	// this package does not interpret the outcomes itself.
	Rule []RuleFn
}

// EffectiveLabel is the human label used in warn-and-default messages: Label
// when set, else Name.
func (f Flag) EffectiveLabel() string {
	if f.Label != "" {
		return f.Label
	}
	return f.Name
}

// IsChart reports whether the flag contributes to the chart-config seed.
func (f Flag) IsChart() bool { return f.JSONKey != "" }

// EffectiveKey is the --chart override key for the flag: Key when set, else Name.
func (f Flag) EffectiveKey() string {
	if f.Key != "" {
		return f.Key
	}
	return f.Name
}

// IsSoft reports whether the flag uses warn-and-default validation rather than
// fatal validation.
func (f Flag) IsSoft() bool {
	return f.ValidSet != nil || f.Normalizer != nil || f.SoftValidate != nil
}
