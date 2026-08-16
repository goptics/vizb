package cli

import (
	"os"
	"slices"
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
	fields := objectFlagFields()
	if len(fields) == 0 {
		return args
	}
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		fieldNames, known := fields[strings.TrimPrefix(arg, "--")]
		known = known && strings.HasPrefix(arg, "--") && !strings.Contains(arg, "=")
		if known && i+1 < len(args) && looksLikeObjectValue(args[i+1], fieldNames) {
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
// or a semicolon bag whose every field starts with a known field name followed
// by '='. Anything starting with '-' (another flag) is never a value.
func looksLikeObjectValue(s string, fieldNames []string) bool {
	if strings.HasPrefix(s, "-") || s == "" {
		return false
	}
	if s == objectFlagOn {
		return true
	}
	if strings.HasPrefix(s, "{") {
		return strings.HasSuffix(s, "}")
	}
	for _, part := range strings.Split(s, ";") {
		key, _, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok || !slices.Contains(fieldNames, strings.TrimSpace(key)) {
			return false
		}
	}
	return true
}

// objectFlagFields collects the registered KindObject flag names with their
// accepted field names, for the space-form rewrite.
func objectFlagFields() map[string][]string {
	fields := map[string][]string{}
	for _, chartType := range internal_charts.Registered() {
		for _, f := range internal_charts.FlagsFor(chartType) {
			if f.Kind != flags.KindObject {
				continue
			}
			if _, ok := fields[f.Name]; ok {
				continue // same descriptor registered by another command path
			}
			fields[f.Name] = f.ObjectFieldNames()
		}
	}
	return fields
}
