package shared

import (
	"fmt"
	"strings"
)

// Stat.Better values: which direction of a metric is the better one.
const (
	BetterHigh = "high"
	BetterLow  = "low"
)

// BetterRule is one --better entry: stats matching Key get direction Dir.
type BetterRule struct {
	Key string
	Dir string
}

// ParseBetterDir normalises a direction word; ok is false unless it is high or low.
func ParseBetterDir(s string) (string, bool) {
	switch d := strings.ToLower(strings.TrimSpace(s)); d {
	case BetterHigh, BetterLow:
		return d, true
	}
	return "", false
}

// ParseBetterFlag parses --better=key:high,key2:low. Each entry splits on its
// last colon so keys may contain colons or units like req/s.
func ParseBetterFlag(raw string) ([]BetterRule, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	seen := map[string]bool{}
	var rules []BetterRule
	for _, entry := range strings.Split(raw, ",") {
		colon := strings.LastIndex(entry, ":")
		if colon == -1 {
			return nil, fmt.Errorf("invalid --better entry %q: expected key:high or key:low", strings.TrimSpace(entry))
		}
		key := strings.TrimSpace(entry[:colon])
		if key == "" {
			return nil, fmt.Errorf("invalid --better entry %q: empty key", strings.TrimSpace(entry))
		}
		dir, ok := ParseBetterDir(entry[colon+1:])
		if !ok {
			return nil, fmt.Errorf("invalid --better direction %q for %q: must be high or low", strings.TrimSpace(entry[colon+1:]), key)
		}
		if seen[strings.ToLower(key)] {
			return nil, fmt.Errorf("duplicate --better key %q", key)
		}
		seen[strings.ToLower(key)] = true
		rules = append(rules, BetterRule{Key: key, Dir: dir})
	}
	return rules, nil
}

// ApplyBetter sets Stat.Better on every stat a rule matches, overriding
// inferred or inline directions, and returns the keys that matched nothing.
func ApplyBetter(data []DataPoint, rules []BetterRule) []string {
	var unmatched []string
	for _, r := range rules {
		hit := false
		for i := range data {
			for j := range data[i].Stats {
				if statMatchesKey(data[i].Stats[j].Type, r.Key) {
					data[i].Stats[j].Better = r.Dir
					hit = true
				}
			}
		}
		if !hit {
			unmatched = append(unmatched, r.Key)
		}
	}
	return unmatched
}

// statMatchesKey matches a stat type ("Name (unit)", "Name/per", "Name") by its
// full type, base name, or unit — whole strings, case-insensitive.
func statMatchesKey(statType, key string) bool {
	if strings.EqualFold(statType, key) {
		return true
	}
	base, unit := statType, ""
	if open := strings.Index(statType, " ("); open != -1 && strings.HasSuffix(statType, ")") {
		base, unit = statType[:open], statType[open+2:len(statType)-1]
	} else if slash := strings.Index(statType, "/"); slash != -1 {
		base = statType[:slash]
	}
	return strings.EqualFold(base, key) || (unit != "" && strings.EqualFold(unit, key))
}
