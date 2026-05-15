package shared

import (
	"sort"
)

type benchGroup struct {
	noTag  *Benchmark
	tagged []Benchmark
}

func (g *benchGroup) addTagged(bench Benchmark) {
	for i, t := range g.tagged {
		if t.Tag != bench.Tag {
			continue
		}
		if t.Timestamp >= bench.Timestamp {
			return
		}
		g.tagged[i] = bench
		return
	}
	g.tagged = append(g.tagged, bench)
}

// MergeBenchmarks performs tag-based smart merging on a slice of benchmarks.
// Benchmarks with the same Name and valid Tag fields are deep-merged into a
// single object. Benchmarks lacking a Tag are appended individually (legacy).
// When both legacy and tagged benchmarks share a Name they are combined into
// one object so accumulated data is preserved across incremental merges.
// dim controls which inner data dimension receives the benchmark tag annotation.
func MergeBenchmarks(benchmarks []Benchmark, dim Dimension) []Benchmark {
	nameOrder := make([]string, 0)
	groups := make(map[string]*benchGroup)

	for _, bench := range benchmarks {
		group, ok := groups[bench.Name]
		if !ok {
			group = new(benchGroup)
			groups[bench.Name] = group
			nameOrder = append(nameOrder, bench.Name)
		}

		if bench.Tag == "" {
			if group.noTag == nil {
				b := bench
				group.noTag = &b
			}
			continue
		}

		group.addTagged(bench)
	}

	result := make([]Benchmark, 0, len(nameOrder))

	for _, name := range nameOrder {
		group := groups[name]

		tagged := make([]Benchmark, len(group.tagged))
		copy(tagged, group.tagged)
		sort.SliceStable(tagged, func(i, j int) bool {
			return tagged[i].Timestamp < tagged[j].Timestamp
		})

		switch {
		case group.noTag != nil && len(tagged) == 0:
			result = append(result, *group.noTag)
		default:
			if group.noTag != nil {
				allBenches := make([]Benchmark, 0, 1+len(tagged))
				allBenches = append(allBenches, *group.noTag)
				allBenches = append(allBenches, tagged...)
				base := deepCloneBenchmark(*group.noTag)
				latest := tagged[len(tagged)-1]
				base.Tag = latest.Tag
				base.Timestamp = latest.Timestamp
				base.History = buildHistory(allBenches, latest.Tag)
				base.Data = mergeData(allBenches, dim)
				result = append(result, base)
				continue
			}

			latest := tagged[len(tagged)-1]
			base := deepCloneBenchmark(latest)
			base.History = buildHistory(tagged, latest.Tag)
			base.Data = mergeData(tagged, dim)
			result = append(result, base)
		}
	}

	return result
}

func deepCloneBenchmark(src Benchmark) Benchmark {
	dst := src
	dst.Data = make([]BenchmarkData, len(src.Data))
	for i := range src.Data {
		dst.Data[i] = deepCloneData(src.Data[i])
	}

	if src.History != nil {
		dst.History = make([]HistoryEntry, len(src.History))
		copy(dst.History, src.History)
	}

	return dst
}

// buildHistory collects tag+timestamp pairs from all benchmarks and their
// existing History entries, excluding the latest tag. Entries are deduplicated
// by tag (keeping the latest timestamp per tag) and sorted chronologically.
func buildHistory(benchmarks []Benchmark, latestTag string) []HistoryEntry {
	seen := make(map[string]string)
	for _, bench := range benchmarks {
		if bench.Tag != "" && bench.Tag != latestTag {
			if ts, ok := seen[bench.Tag]; !ok || bench.Timestamp > ts {
				seen[bench.Tag] = bench.Timestamp
			}
		}
		for _, entry := range bench.History {
			if entry.Tag == latestTag {
				continue
			}
			if ts, ok := seen[entry.Tag]; !ok || entry.Timestamp > ts {
				seen[entry.Tag] = entry.Timestamp
			}
		}
	}

	if len(seen) == 0 {
		return nil
	}

	entries := make([]HistoryEntry, 0, len(seen))
	for tag, ts := range seen {
		entries = append(entries, HistoryEntry{Tag: tag, Timestamp: ts})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp < entries[j].Timestamp
	})
	return entries
}

func mergeData(benchmarks []Benchmark, dim Dimension) []BenchmarkData {
	var result []BenchmarkData
	for _, bench := range benchmarks {
		for _, item := range bench.Data {
			if dimFieldEmpty(item, dim) {
				result = append(result, injectTag(item, bench.Tag, dim))
			} else {
				result = append(result, deepCloneData(item))
			}
		}
	}
	return result
}

func dimFieldEmpty(item BenchmarkData, dim Dimension) bool {
	switch dim {
	case DimensionXAxis:
		return item.XAxis == ""
	case DimensionYAxis:
		return item.YAxis == ""
	default:
		return item.Name == ""
	}
}

func injectTag(item BenchmarkData, tag string, dim Dimension) BenchmarkData {
	item = deepCloneData(item)

	switch dim {
	case DimensionXAxis:
		item.XAxis = applyInjection(item.XAxis, tag)
	case DimensionYAxis:
		item.YAxis = applyInjection(item.YAxis, tag)
	default:
		item.Name = applyInjection(item.Name, tag)
	}

	return item
}

func applyInjection(existing string, tag string) string {
	if existing == "" {
		return tag
	}
	return existing + " - " + tag
}

func deepCloneData(src BenchmarkData) BenchmarkData {
	dst := src
	if src.Stats != nil {
		dst.Stats = make([]Stat, len(src.Stats))
		copy(dst.Stats, src.Stats)
	}
	return dst
}
