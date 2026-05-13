package shared

import "maps"

// MergeBenchmarks performs tag-based smart merging on a slice of benchmarks.
// Benchmarks with the same Name and valid Tag fields are deep-merged into a
// single object. Benchmarks lacking a Tag are appended individually (legacy).
// injectDim controls which inner data dimension receives the benchmark tag
// annotation: "n" for Name, "x" for XAxis, "y" for YAxis.
func MergeBenchmarks(benchmarks []Benchmark, injectDim string) []Benchmark {
	groups := groupByName(benchmarks)
	var result []Benchmark

	for _, group := range groups {
		noTag, withTag := splitByTag(group)
		result = append(result, noTag...)

		switch len(withTag) {
		case 0:
			continue
		case 1:
			result = append(result, withTag[0])
		default:
			result = append(result, tagBasedMerge(withTag, injectDim)...)
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

func tagBasedMerge(benchmarks []Benchmark, injectDim string) []Benchmark {
	latestIdx, tie := findLatest(benchmarks)
	if tie {
		return benchmarks
	}

	merged := deepCloneBenchmark(benchmarks[latestIdx])
	merged.Runtimes = mergeRuntimes(benchmarks)
	merged.Data = mergeData(benchmarks, injectDim)
	return []Benchmark{merged}
}

func findLatest(benchmarks []Benchmark) (idx int, tie bool) {
	var maxTS string
	idx = -1

	for i, bench := range benchmarks {
		for _, ts := range bench.Runtimes {
			if ts > maxTS {
				maxTS = ts
				idx = i
				tie = false
			} else if ts == maxTS && i != idx {
				tie = true
			}
		}
	}

	return idx, tie
}

func deepCloneBenchmark(src Benchmark) Benchmark {
	dst := src
	dst.Data = make([]BenchmarkData, len(src.Data))
	copy(dst.Data, src.Data)

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

func mergeData(benchmarks []Benchmark, injectDim string) []BenchmarkData {
	var result []BenchmarkData
	for _, bench := range benchmarks {
		for _, item := range bench.Data {
			result = append(result, injectTag(item, bench.Tag, injectDim))
		}
	}
	return result
}

func injectTag(item BenchmarkData, tag string, dim string) BenchmarkData {
	item = deepCloneData(item)

	switch dim {
	case "x":
		item.XAxis = applyInjection(item.XAxis, tag)
	case "y":
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
