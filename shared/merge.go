package shared

import (
	"sort"
)

const noTagKey = "__no_tag__"

// MergeBenchmarks performs tag-based smart merging on a slice of benchmarks.
// Benchmarks with the same Name and valid Tag fields are deep-merged into a
// single object. Benchmarks lacking a Tag are appended individually (legacy).
// When both legacy and tagged benchmarks share a Name they are combined into
// one object so accumulated data is preserved across incremental merges.
// dim controls which inner data dimension receives the benchmark tag annotation.
func MergeBenchmarks(benchmarks []Benchmark, dim Dimension) []Benchmark {
	nameOrder := make([]string, 0)
	groups := make(map[string]map[string]*Benchmark)

	for _, bench := range benchmarks {
		tags, ok := groups[bench.Name]
		if !ok {
			tags = make(map[string]*Benchmark)
			groups[bench.Name] = tags
			nameOrder = append(nameOrder, bench.Name)
		}

		tag := bench.Tag
		if tag == "" {
			tag = noTagKey
		}

		if existing, exists := tags[tag]; exists && existing.Timestamp >= bench.Timestamp {
			continue
		}

		tags[tag] = &bench
	}

	result := make([]Benchmark, 0, len(nameOrder))

	for _, name := range nameOrder {
		tags := groups[name]

		noTag := tags[noTagKey]
		delete(tags, noTagKey)

		tagged := make([]Benchmark, 0, len(tags))
		for _, b := range tags {
			tagged = append(tagged, *b)
		}
		sort.SliceStable(tagged, func(i, j int) bool {
			return tagged[i].Timestamp < tagged[j].Timestamp
		})

		switch {
		case noTag != nil && len(tagged) == 0:
			result = append(result, *noTag)
		default:
			if noTag != nil {
				allBenches := make([]Benchmark, 0, 1+len(tagged))
				allBenches = append(allBenches, *noTag)
				allBenches = append(allBenches, tagged...)
				base := deepCloneBenchmark(*noTag)
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

type historyCandidate struct {
	timestamp string
	cpu       *CPUInfo
	os        string
}

// buildHistory collects tag+timestamp+cpu+os from all benchmarks and their
// existing History entries, excluding the latest tag. Entries are deduplicated
// by tag (keeping the latest timestamp per tag) and sorted chronologically.
func buildHistory(benchmarks []Benchmark, latestTag string) []HistoryEntry {
	seen := make(map[string]historyCandidate)
	for _, bench := range benchmarks {
		if bench.Tag != "" && bench.Tag != latestTag {
			if c, ok := seen[bench.Tag]; !ok || bench.Timestamp > c.timestamp {
				var cpuPtr *CPUInfo
				if bench.CPU.Name != "" || bench.CPU.Cores != 0 {
					cpu := bench.CPU
					cpuPtr = &cpu
				}
				seen[bench.Tag] = historyCandidate{timestamp: bench.Timestamp, cpu: cpuPtr, os: bench.OS}
			}
		}
		for _, entry := range bench.History {
			if entry.Tag == latestTag {
				continue
			}
			if c, ok := seen[entry.Tag]; !ok || entry.Timestamp > c.timestamp {
				seen[entry.Tag] = historyCandidate{timestamp: entry.Timestamp, cpu: entry.CPU, os: entry.OS}
			}
		}
	}

	if len(seen) == 0 {
		return nil
	}

	entries := make([]HistoryEntry, 0, len(seen))
	for tag, c := range seen {
		entries = append(entries, HistoryEntry{Tag: tag, Timestamp: c.timestamp, CPU: c.cpu, OS: c.os})
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
	case DimensionZAxis:
		return item.ZAxis == ""
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
	case DimensionZAxis:
		item.ZAxis = applyInjection(item.ZAxis, tag)
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
