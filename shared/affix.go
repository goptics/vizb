// Affix lives in internal/flags (which must stay a dependency leaf, unlike
// shared) and is re-exported here: shared already imports internal/charts, so
// internal/charts cannot import shared without an import cycle.
package shared

import "github.com/goptics/vizb/internal/flags"

// Affix is an alias for flags.Affix — the identical underlying type.
type Affix = flags.Affix

// ParseAffix parses a --label-prefix/suffix flag value: a plain symbol
// ("%", "$") treated as global, or a comma/semicolon-separated list of
// global and per-column symbols ("%,col=ms", "col1=%;col2=ms").
// It returns nil only on error.
func ParseAffix(input string) (*Affix, error) {
	return flags.ParseAffix(input)
}
