package cli

import (
	"os"
	"slices"
	"strings"

	"github.com/goptics/vizb/shared"
	"github.com/spf13/pflag"
)

// statFlagAll is the NoOptDefVal sentinel: pflag requires a non-empty string for
// optional-value flags. Set() converts it to []string{} (all categories). "all"
// is also accepted as an explicit value so --stat=all works identically.
const statFlagAll = "all"

// statValue is a pflag.Value that makes --stat optional-value:
//
//	--stat          → all categories ([]string{})
//	--stat=a,b      → specific categories
//	(omitted)       → nil = disabled
type statValue struct{ value *[]string }

func (f *statValue) String() string {
	if f.value == nil || *f.value == nil {
		return ""
	}
	return strings.Join(*f.value, ",")
}

func (f *statValue) Set(val string) error {
	if val == statFlagAll {
		*f.value = []string{}
		return nil
	}
	*f.value = strings.Split(val, ",")
	return nil
}

func (f *statValue) Type() string { return "string" }

// BindStatFlag registers --stat on fs, pointing at target. Exported so commands
// that don't use a FlagBag (e.g. the ui subcommand) can register the same flag.
func BindStatFlag(fs *pflag.FlagSet, target *[]string) {
	fs.Var(&statValue{value: target}, "stat", "Enable stats panel; omit to disable, use alone for all categories, or =cat1,cat2 for specific (counts, center, spread, extremes, shape, percentiles, confidence, correlations)")
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

// looksLikeStatValue reports whether s could be an argument to --stat. Returns
// false for anything starting with '-' (another flag) or not composed entirely
// of recognised stat category names (or the "all" sentinel).
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
