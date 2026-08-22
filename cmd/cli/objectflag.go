package cli

import (
	"os"
	"strings"

	internal_charts "github.com/goptics/vizb/internal/charts"
	"github.com/goptics/vizb/internal/flags"
)

// objectFlagOn is the NoOptDefVal sentinel for optional-value object flags
// (KindObject). pflag requires a non-empty string for optional-value flags;
// Set() stores it verbatim and the seed builder treats it as the empty bag
// (plus the flag's injected on-switch). "on" is also accepted as an explicit
// value so --bg=on works identically.
const objectFlagOn = "on"

// objectValue is a pflag.Value that makes KindObject flags optional-value:
//
//	--bg              → empty object bag (sentinel "on")
//	--bg=k=v;k2=v2    → a semicolon bag of typed fields
//	(omitted)         → "" = disabled
type objectValue struct{ value *string }

func (o *objectValue) String() string {
	if o.value == nil {
		return ""
	}
	return *o.value
}

func (o *objectValue) Set(val string) error {
	*o.value = val
	return nil
}

func (o *objectValue) Type() string { return "string" }

// RewriteObjectArg rewrites `--bg VALUE` (space-separated) to `--bg=VALUE` so
// pflag can parse optional-value object flags despite the NoOptDefVal. Without
// this rewrite, pflag consumes the sentinel "on" and treats VALUE as a
// positional arg. Mirrors RewriteStatArg: the next arg must look like an
// object bag and must not be an existing file.
func RewriteObjectArg(args []string) []string {
	names := objectFlagNames()
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		known := names[strings.TrimPrefix(arg, "--")] &&
			strings.HasPrefix(arg, "--") && !strings.Contains(arg, "=")
		if known && i+1 < len(args) && looksLikeObjectValue(args[i+1]) {
			if _, err := os.Stat(args[i+1]); err != nil {
				out = append(out, arg+"="+args[i+1])
				i++
				continue
			}
		}
		out = append(out, arg)
	}
	return out
}

// looksLikeObjectValue reports whether s could be an argument to an object
// flag: the "on" sentinel, a brace-wrapped token (rejected later with a hint),
// or a semicolon key=value bag. Unknown keys are still rewritten so FlagBag
// validation can reject them instead of treating the bag as a filename.
// Anything starting with '-' (another flag) is never a value.
func looksLikeObjectValue(s string) bool {
	if strings.HasPrefix(s, "-") || s == "" {
		return false
	}
	if s == objectFlagOn {
		return true
	}
	if strings.HasPrefix(s, "{") {
		return strings.HasSuffix(s, "}")
	}
	seen := false
	for _, part := range strings.Split(s, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if !strings.Contains(part, "=") {
			return false
		}
		seen = true
	}
	return seen
}

// objectFlagNames collects the registered KindObject flag names for the
// space-form rewrite.
func objectFlagNames() map[string]bool {
	names := map[string]bool{}
	for _, chartType := range internal_charts.Registered() {
		for _, f := range internal_charts.FlagsFor(chartType) {
			if f.Kind == flags.KindObject {
				names[f.Name] = true
			}
		}
	}
	return names
}
