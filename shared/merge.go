package shared

import (
	"sort"

	config_charts "github.com/goptics/vizb/config/charts"
)

const noTagKey = "__no_tag__"

// MergeDatasets performs tag-based smart merging on a slice of benchmarks.
// Benchmarks with the same Name and valid Tag fields are deep-merged into a
// single object. Benchmarks lacking a Tag are appended individually (legacy).
// When both legacy and tagged benchmarks share a Name they are combined into
// one object so accumulated data is preserved across incremental merges.
// dim controls which inner data dimension receives the benchmark tag annotation.
func MergeDatasets(benchmarks []Dataset, dim Dimension) []Dataset {
	nameOrder := make([]string, 0)
	groups := make(map[string]map[string]*Dataset)

	for _, ds := range benchmarks {
		tags, ok := groups[ds.Name]
		if !ok {
			tags = make(map[string]*Dataset)
			groups[ds.Name] = tags
			nameOrder = append(nameOrder, ds.Name)
		}

		tag := ds.Tag
		if tag == "" {
			tag = noTagKey
		}

		if existing, exists := tags[tag]; exists && existing.Timestamp >= ds.Timestamp {
			continue
		}

		tags[tag] = &ds
	}

	result := make([]Dataset, 0, len(nameOrder))

	for _, name := range nameOrder {
		tags := groups[name]

		noTag := tags[noTagKey]
		delete(tags, noTagKey)

		tagged := make([]Dataset, 0, len(tags))
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
				allDatasets := make([]Dataset, 0, 1+len(tagged))
				allDatasets = append(allDatasets, *noTag)
				allDatasets = append(allDatasets, tagged...)
				base := deepCloneDataset(*noTag)
				latest := tagged[len(tagged)-1]
				base.Tag = latest.Tag
				base.Timestamp = latest.Timestamp
				base.History = buildHistory(allDatasets, latest.Tag)
				base.Data = mergeData(allDatasets, dim)
				result = append(result, base)
				continue
			}

			latest := tagged[len(tagged)-1]
			base := deepCloneDataset(latest)
			base.History = buildHistory(tagged, latest.Tag)
			base.Data = mergeData(tagged, dim)
			result = append(result, base)
		}
	}

	return result
}

func deepCloneDataset(src Dataset) Dataset {
	dst := src
	dst.Data = make([]DataPoint, len(src.Data))
	for i := range src.Data {
		dst.Data[i] = deepCloneData(src.Data[i])
	}

	if src.History != nil {
		dst.History = make([]HistoryEntry, len(src.History))
		copy(dst.History, src.History)
	}

	if src.Meta != nil {
		m := *src.Meta
		if src.Meta.CPU != nil {
			cpu := *src.Meta.CPU
			m.CPU = &cpu
		}
		dst.Meta = &m
	}

	if src.Axes != nil {
		dst.Axes = make([]Axis, len(src.Axes))
		copy(dst.Axes, src.Axes)
	}

	if src.Settings != nil {
		dst.Settings = make([]config_charts.ChartConfig, len(src.Settings))
		copy(dst.Settings, src.Settings)
	}

	return dst
}

type historyCandidate struct {
	timestamp string
	meta      *Meta
}

// buildHistory collects tag+timestamp+meta from all benchmarks and their
// existing History entries, excluding the latest tag. Entries are deduplicated
// by tag (keeping the latest timestamp per tag) and sorted chronologically.
func buildHistory(benchmarks []Dataset, latestTag string) []HistoryEntry {
	seen := make(map[string]historyCandidate)
	for _, ds := range benchmarks {
		if ds.Tag != "" && ds.Tag != latestTag {
			if c, ok := seen[ds.Tag]; !ok || ds.Timestamp > c.timestamp {
				var metaPtr *Meta
				if ds.Meta != nil {
					metaCopy := *ds.Meta
					if ds.Meta.CPU != nil {
						cpu := *ds.Meta.CPU
						metaCopy.CPU = &cpu
					}
					metaPtr = &metaCopy
				}
				seen[ds.Tag] = historyCandidate{timestamp: ds.Timestamp, meta: metaPtr}
			}
		}
		for _, entry := range ds.History {
			if entry.Tag == latestTag {
				continue
			}
			if c, ok := seen[entry.Tag]; !ok || entry.Timestamp > c.timestamp {
				meta := entry.Meta
				if meta != nil {
					m := *meta
					if meta.CPU != nil {
						cpu := *meta.CPU
						m.CPU = &cpu
					}
					meta = &m
				}
				seen[entry.Tag] = historyCandidate{timestamp: entry.Timestamp, meta: meta}
			}
		}
	}

	if len(seen) == 0 {
		return nil
	}

	entries := make([]HistoryEntry, 0, len(seen))
	for tag, c := range seen {
		entries = append(entries, HistoryEntry{Tag: tag, Timestamp: c.timestamp, Meta: c.meta})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp < entries[j].Timestamp
	})
	return entries
}

func mergeData(benchmarks []Dataset, dim Dimension) []DataPoint {
	var result []DataPoint
	for _, ds := range benchmarks {
		for _, item := range ds.Data {
			if dimFieldEmpty(item, dim) {
				result = append(result, injectTag(item, ds.Tag, dim))
			} else {
				result = append(result, deepCloneData(item))
			}
		}
	}
	return result
}

func dimFieldEmpty(item DataPoint, dim Dimension) bool {
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

func injectTag(item DataPoint, tag string, dim Dimension) DataPoint {
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

func deepCloneData(src DataPoint) DataPoint {
	dst := src
	if src.Stats != nil {
		dst.Stats = make([]Stat, len(src.Stats))
		copy(dst.Stats, src.Stats)
		for i, s := range src.Stats {
			if s.Value != nil {
				v := *s.Value
				dst.Stats[i].Value = &v
			}
		}
	}
	return dst
}
