// Package flags defines the single flag-descriptor type shared by every vizb
// command. One Flag declares a CLI option once; the cobra binder, the --chart
// override parser, and the chart-seed/parser-config builders all consume the
// same descriptor. Adding a flag means adding one Flag — nothing else.
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
	KindStringSlice             // []string (comma/repeat), e.g. --group, --charts
	KindStat                    // optional-value string slice, e.g. --stat
)

// Flag is the single source of truth for one CLI option.
//
// A descriptor is one of two species:
//
//   - chart flag: JSONKey is set. Its value is encoded into the chart-config
//     seed (and decoded into a typed chart Config). Validation, when present, is
//     fatal via Validate — unless the soft trio is set (e.g. scale), in which
//     case it is warn-and-default AND still seeded.
//   - data flag: JSONKey is empty. Its value feeds parser.Config / dataset
//     metadata, read back by name. Validation is warn-and-default via the soft
//     trio (ValidSet / Normalizer / SoftValidate), never fatal.
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
