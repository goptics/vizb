package shared

import (
	"maps"
	"sort"
)

type benchGroup struct {
	noTag       Benchmark
	hasNoTag    bool
	taggedByTag map[string]*taggedEntry
}

type taggedEntry struct {
	benchmark Benchmark
	timestamp string
}

// latestRuntime returns the latest (greatest) timestamp from a runtimes map.
// Timestamps are expected to be in ISO 8601 / RFC 3339 format, which
// sorts lexicographically in the same order as chronologically.
func latestRuntime(runtimes map[string]string) string {
	var latest string
	for _, ts := range runtimes {
		if ts > latest {
			latest = ts
		}
	}
	return latest
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
			group = &benchGroup{taggedByTag: make(map[string]*taggedEntry)}
			groups[bench.Name] = group
			nameOrder = append(nameOrder, bench.Name)
		}

		if bench.Tag == "" {
			if !group.hasNoTag {
				group.noTag = bench
				group.hasNoTag = true
			}
			continue
		}

		latestTS := latestRuntime(bench.Runtimes)
		if existing, exists := group.taggedByTag[bench.Tag]; exists && existing.timestamp >= latestTS {
			continue
		}

		group.taggedByTag[bench.Tag] = &taggedEntry{
			benchmark: bench,
			timestamp: latestTS,
		}
	}

	result := make([]Benchmark, 0, len(nameOrder))

	for _, name := range nameOrder {
		group := groups[name]

		tagged := make([]*taggedEntry, 0, len(group.taggedByTag))
		for _, entry := range group.taggedByTag {
			tagged = append(tagged, entry)
		}
		sort.SliceStable(tagged, func(i, j int) bool {
			return tagged[i].timestamp < tagged[j].timestamp
		})

		switch {
		case group.hasNoTag && len(tagged) == 0:
			result = append(result, group.noTag)
		default:
			benches := make([]Benchmark, len(tagged))
			for i, e := range tagged {
				benches[i] = e.benchmark
			}

			if group.hasNoTag {
				allBenches := make([]Benchmark, 0, 1+len(benches))
				allBenches = append(allBenches, group.noTag)
				allBenches = append(allBenches, benches...)
				base := deepCloneBenchmark(group.noTag)
				base.Runtimes = mergeRuntimes(allBenches)
				base.Data = mergeData(allBenches, dim)
				base.Tag = ""
				result = append(result, base)
			} else {
				base := deepCloneBenchmark(benches[len(benches)-1])
				base.Runtimes = mergeRuntimes(benches)
				base.Data = mergeData(benches, dim)
				base.Tag = ""
				result = append(result, base)
			}
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

	if src.Runtimes != nil {
		dst.Runtimes = make(map[string]string, len(src.Runtimes))
		maps.Copy(dst.Runtimes, src.Runtimes)
	}

	return dst
}

func mergeRuntimes(benchmarks []Benchmark) map[string]string {
	result := make(map[string]string)
	for _, bench := range benchmarks {
		maps.Copy(result, bench.Runtimes)
	}
	return result
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
