package shared

import "maps"

// MergeBenchmarks performs tag-based smart merging on a slice of benchmarks.
// Benchmarks with the same Name and valid Tag fields are deep-merged into a
// single object. Benchmarks lacking a Tag are appended individually (legacy).
// dim controls which inner data dimension receives the benchmark tag annotation.
func MergeBenchmarks(benchmarks []Benchmark, dim Dimension) []Benchmark {
	groups := groupByName(benchmarks)
	var result []Benchmark

	for _, group := range groups {
		noTag, withTag := splitByTag(group)
		result = append(result, noTag...)

		switch len(withTag) {
		case 0:
		case 1:
			result = append(result, withTag[0])
		default:
			result = append(result, tagBasedMerge(withTag, dim)...)
		}
	}

	return result
}

func groupByName(benchmarks []Benchmark) map[string][]Benchmark {
	groups := make(map[string][]Benchmark)
	for _, bench := range benchmarks {
		groups[bench.Name] = append(groups[bench.Name], bench)
	}
	return groups
}

func splitByTag(benchmarks []Benchmark) (noTag, withTag []Benchmark) {
	for _, bench := range benchmarks {
		if bench.Tag == "" {
			noTag = append(noTag, bench)
		} else {
			withTag = append(withTag, bench)
		}
	}
	return
}

func tagBasedMerge(benchmarks []Benchmark, dim Dimension) []Benchmark {
	latestIdx, tie := findLatest(benchmarks)
	if tie {
		return benchmarks
	}

	merged := deepCloneBenchmark(benchmarks[latestIdx])
	merged.Runtimes = mergeRuntimes(benchmarks)
	merged.Data = mergeData(benchmarks, dim)
	return []Benchmark{merged}
}

func findLatest(benchmarks []Benchmark) (int, bool) {
	var maxTS string

	for _, b := range benchmarks {
		for _, ts := range b.Runtimes {
			if ts > maxTS {
				maxTS = ts
			}
		}
	}

	if maxTS == "" {
		return -1, len(benchmarks) > 1
	}

	var latest, count int
	for i, b := range benchmarks {
		for _, ts := range b.Runtimes {
			if ts == maxTS {
				latest = i
				count++
				break
			}
		}
	}

	return latest, count > 1
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
			result = append(result, injectTag(item, bench.Tag, dim))
		}
	}
	return result
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
