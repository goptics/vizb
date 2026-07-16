package shared

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MergeSuite struct {
	suite.Suite
}

func makeBench(tag, name, timestamp string, data []DataPoint) Dataset {
	return Dataset{
		Tag:       tag,
		Timestamp: timestamp,
		Name:      name,
		Data:      data,
	}
}

func axisKeys(axes []Axis) []string {
	keys := make([]string, len(axes))
	for i, a := range axes {
		keys[i] = a.Key
	}
	return keys
}

func (s *MergeSuite) TestMergeInjectsTagIntoEmptyDimension() {
	bench := makeBench("v2", "DS", "2026-01-01T00:00:00Z", []DataPoint{
		{XAxis: "", YAxis: "10", Stats: []Stat{{Type: "time", Value: F64(1)}}},
	})

	result := MergeDatasets([]Dataset{bench}, DimensionXAxis)
	s.Require().Len(result, 1)
	s.Equal("v2", result[0].Data[0].XAxis)
}

func (s *MergeSuite) TestDeepCloneDatasetCopiesStats() {
	bench1 := makeBench("1", "D", "t1", []DataPoint{{Name: "p1", Stats: []Stat{{Type: "time", Value: F64(1)}}}})
	bench2 := makeBench("2", "D", "t2", []DataPoint{{Name: "p2", Stats: []Stat{{Type: "time", Value: F64(2)}}}})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	s.Require().Len(result, 1)
	s.Require().Len(result[0].Data, 2)
	s.Equal(1.0, *result[0].Data[0].Stats[0].Value)
	s.Equal(2.0, *result[0].Data[1].Stats[0].Value)
}

func (s *MergeSuite) TestMergeDatasetsSmartMerge() {
	bench1 := makeBench("1", "My Dataset", "2026-05-13T10:00:00Z", []DataPoint{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})
	bench2 := makeBench("2", "My Dataset", "2026-05-13T10:05:00Z", []DataPoint{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	s.Len(result, 1)

	merged := result[0]
	s.Equal("2", merged.Tag)
	s.Equal("My Dataset", merged.Name)
	s.Equal("2026-05-13T10:05:00Z", merged.Timestamp)
	s.Equal([]HistoryEntry{
		{Tag: "1", Timestamp: "2026-05-13T10:00:00Z"},
	}, merged.History)
	s.Len(merged.Data, 2)
	s.Equal("1", merged.Data[0].Name)
	s.Equal("2", merged.Data[1].Name)
}

func (s *MergeSuite) TestMergeDatasetsMixedGroup() {
	bench1 := makeBench("1", "My Dataset", "2026-05-13T10:00:00Z", []DataPoint{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})
	bench2 := makeBench("2", "My Dataset", "2026-05-13T10:05:00Z", []DataPoint{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})
	noTagBench := Dataset{
		Tag:  "",
		Name: "My Dataset",
		Data: []DataPoint{{Name: "legacy", XAxis: "x", YAxis: "y"}},
	}

	result := MergeDatasets([]Dataset{bench1, bench2, noTagBench}, DimensionName)
	s.Len(result, 1)

	merged := result[0]
	s.Len(merged.Data, 3)
	s.Equal("legacy", merged.Data[0].Name)
	s.Equal("1", merged.Data[1].Name)
	s.Equal("2", merged.Data[2].Name)
	s.Equal("2", merged.Tag)
	s.Equal("2026-05-13T10:05:00Z", merged.Timestamp)
}

func (s *MergeSuite) TestMergeDatasetsAllNoTag() {
	bench1 := Dataset{Name: "Bench A", Data: []DataPoint{{Name: "a"}}}
	bench2 := Dataset{Name: "Bench A", Data: []DataPoint{{Name: "b"}}}

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	s.Len(result, 1)
	s.Equal("a", result[0].Data[0].Name)
}

func (s *MergeSuite) TestMergeDatasetsTimestampTie() {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{{Name: "a"}})
	bench2 := makeBench("2", "Test", "2026-05-13T10:00:00Z", []DataPoint{{Name: "b"}})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	s.Len(result, 1)
	s.Len(result[0].Data, 2)
	s.ElementsMatch([]string{"a", "b"}, []string{result[0].Data[0].Name, result[0].Data[1].Name})
	s.Contains([]string{"1", "2"}, result[0].Tag)
}

func (s *MergeSuite) TestMergeDatasetsSingleDataset() {
	bench := makeBench("1", "Solo", "2026-05-13T10:00:00Z", []DataPoint{{Name: "x"}})
	result := MergeDatasets([]Dataset{bench}, DimensionName)
	s.Len(result, 1)
	s.Equal("Solo", result[0].Name)
	s.Equal("2026-05-13T10:00:00Z", result[0].Timestamp)
	s.Nil(result[0].History)
}

func (s *MergeSuite) TestMergeDatasetsPopulatedName() {
	bench1 := makeBench("1", "digits", "2026-05-13T10:00:00Z", []DataPoint{
		{Name: "digits", XAxis: "speed", YAxis: "1e4"},
	})
	bench2 := makeBench("2", "digits", "2026-05-13T10:05:00Z", []DataPoint{
		{Name: "digits", XAxis: "speed", YAxis: "1e4"},
	})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	s.Len(result, 1)
	s.Equal("digits", result[0].Data[0].Name)
	s.Equal("digits", result[0].Data[1].Name)
}

func (s *MergeSuite) TestMergeDatasetsInjectDimensionX() {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{
		{XAxis: "", YAxis: "100"},
	})
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []DataPoint{
		{XAxis: "", YAxis: "200"},
	})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionXAxis)
	s.Len(result, 1)
	s.Equal("1", result[0].Data[0].XAxis)
	s.Equal("2", result[0].Data[1].XAxis)
	s.Equal("", result[0].Data[0].Name)
	s.Equal([]string{"x"}, axisKeys(result[0].Axes))
}

func (s *MergeSuite) TestMergeDatasetsInjectDimensionY() {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{
		{XAxis: "x", YAxis: ""},
	})
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []DataPoint{
		{XAxis: "x", YAxis: ""},
	})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionYAxis)
	s.Len(result, 1)
	s.Equal("1", result[0].Data[0].YAxis)
	s.Equal("2", result[0].Data[1].YAxis)
	s.Equal([]string{"y"}, axisKeys(result[0].Axes))
}

func (s *MergeSuite) TestMergeDatasetsInjectDimensionZ() {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{
		{XAxis: "x", YAxis: "y", ZAxis: ""},
	})
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []DataPoint{
		{XAxis: "x", YAxis: "y", ZAxis: ""},
	})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionZAxis)
	s.Len(result, 1)
	s.Equal("1", result[0].Data[0].ZAxis)
	s.Equal("2", result[0].Data[1].ZAxis)
	s.Equal([]string{"z"}, axisKeys(result[0].Axes))
}

func (s *MergeSuite) TestMergeEnsuresZAxisInAxes() {
	axes := []Axis{{Key: "x"}, {Key: "y"}}
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{
		{XAxis: "size", YAxis: "100", ZAxis: ""},
	})
	bench1.Axes = axes
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []DataPoint{
		{XAxis: "size", YAxis: "200", ZAxis: ""},
	})
	bench2.Axes = axes

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionZAxis)
	s.Require().Len(result, 1)
	s.Equal([]string{"x", "y", "z"}, axisKeys(result[0].Axes))
	s.Equal("1", result[0].Data[0].ZAxis)
	s.Equal("2", result[0].Data[1].ZAxis)
}

func (s *MergeSuite) TestMergeEnsuresNameAxisInAxes() {
	axes := []Axis{{Key: "x"}, {Key: "y"}}
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{
		{Name: "", XAxis: "speed", YAxis: "100"},
	})
	bench1.Axes = axes
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []DataPoint{
		{Name: "", XAxis: "speed", YAxis: "200"},
	})
	bench2.Axes = axes

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	s.Require().Len(result, 1)
	s.Equal([]string{"name", "x", "y"}, axisKeys(result[0].Axes))
	s.Equal("1", result[0].Data[0].Name)
	s.Equal("2", result[0].Data[1].Name)
}

func (s *MergeSuite) TestMergeDoesNotDuplicateAxis() {
	axes := []Axis{{Key: "x"}, {Key: "y"}}
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{
		{XAxis: "", YAxis: "100"},
	})
	bench1.Axes = axes
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []DataPoint{
		{XAxis: "", YAxis: "200"},
	})
	bench2.Axes = axes

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionXAxis)
	s.Require().Len(result, 1)
	s.Equal([]string{"x", "y"}, axisKeys(result[0].Axes))
}

func (s *MergeSuite) TestMergePreservesMetricAxisWhenEnsuringInjectAxis() {
	axes := []Axis{{Key: "x"}, {Key: "y"}, {Key: "metric", Label: "value"}}
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{
		{XAxis: "size", YAxis: "100", ZAxis: ""},
	})
	bench1.Axes = axes
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []DataPoint{
		{XAxis: "size", YAxis: "200", ZAxis: ""},
	})
	bench2.Axes = axes

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionZAxis)
	s.Require().Len(result, 1)
	s.Equal([]string{"x", "y", "z", "metric"}, axisKeys(result[0].Axes))
}

func (s *MergeSuite) TestMergeDatasetsHistoryMerge() {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{{Name: "a"}})
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []DataPoint{{Name: "b"}})
	bench2.History = []HistoryEntry{
		{Tag: "extra", Timestamp: "2026-05-13T11:00:00Z"},
	}

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	s.Len(result, 1)
	s.Equal("2", result[0].Tag)
	s.Equal("2026-05-13T10:05:00Z", result[0].Timestamp)
	s.Equal([]HistoryEntry{
		{Tag: "1", Timestamp: "2026-05-13T10:00:00Z"},
		{Tag: "extra", Timestamp: "2026-05-13T11:00:00Z"},
	}, result[0].History)
}

func (s *MergeSuite) TestMergeDatasetsHistoryMetaPropagation() {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{{Name: "a"}})
	bench1.Meta = &Meta{
		CPU:  &CPUInfo{Name: "Intel i7", Cores: 8},
		OS:   "linux",
		Arch: "amd64",
		Pkg:  "github.com/foo/bar",
	}
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []DataPoint{{Name: "b"}})
	bench2.Meta = &Meta{
		CPU:  &CPUInfo{Name: "Apple M2", Cores: 10},
		OS:   "darwin",
		Arch: "arm64",
		Pkg:  "github.com/baz/qux",
	}

	datasets := []Dataset{bench1, bench2}
	result := MergeDatasets(datasets, DimensionName)
	s.Len(result, 1)
	s.Equal("2", result[0].Tag)

	s.Require().Len(result[0].History, 1)
	entry := result[0].History[0]
	s.Equal("1", entry.Tag)
	s.Equal("2026-05-13T10:00:00Z", entry.Timestamp)

	// FULL meta propagates into history (not just cpu/os).
	s.Require().NotNil(entry.Meta)
	s.Require().NotNil(entry.Meta.CPU)
	s.Equal("Intel i7", entry.Meta.CPU.Name)
	s.Equal(8, entry.Meta.CPU.Cores)
	s.Equal("linux", entry.Meta.OS)
	s.Equal("amd64", entry.Meta.Arch)
	s.Equal("github.com/foo/bar", entry.Meta.Pkg)

	// Pointer independence: history CPU must not alias the source dataset's CPU.
	s.NotSame(datasets[0].Meta.CPU, entry.Meta.CPU, "history Meta.CPU aliases source dataset Meta.CPU; expected a deep copy")
}

func (s *MergeSuite) TestMergeDatasetsDifferentNames() {
	bench1 := makeBench("1", "Bench A", "2026-05-13T10:00:00Z", []DataPoint{{Name: "a"}})
	bench2 := makeBench("2", "Bench B", "2026-05-13T10:05:00Z", []DataPoint{{Name: "b"}})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	s.Len(result, 2)
}

func (s *MergeSuite) TestMergeDatasetsEmptyInput() {
	result := MergeDatasets(nil, DimensionName)
	s.Empty(result)
}

func (s *MergeSuite) TestMergeDatasetsNoMutation() {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{{Name: "a"}})
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []DataPoint{{Name: "b"}})

	original := []Dataset{bench1, bench2}
	result := MergeDatasets(original, DimensionName)

	s.Equal("a", original[0].Data[0].Name)
	s.Equal("b", original[1].Data[0].Name)
	s.Len(result, 1)
}

func (s *MergeSuite) TestMergeDatasetsSameNameSameTagDedup() {
	bench1 := makeBench("v1", "Bench", "2026-05-13T10:00:00Z",
		[]DataPoint{{Name: "", XAxis: "speed", YAxis: "1e4"}})
	bench2 := makeBench("v1", "Bench", "2026-05-14T10:00:00Z",
		[]DataPoint{{Name: "", XAxis: "speed", YAxis: "1e4"}})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	s.Len(result, 1)
	s.Len(result[0].Data, 1)
	s.Equal("v1", result[0].Data[0].Name)
	s.Equal("v1", result[0].Tag)
	s.Equal("2026-05-14T10:00:00Z", result[0].Timestamp)
	s.Nil(result[0].History)
}

func (s *MergeSuite) TestMergeUsesNewestNonEmptyTheme() {
	legacy := Dataset{
		Name: "Bench", Theme: "vintage", Timestamp: "2026-05-12T10:00:00Z",
		Data: []DataPoint{{Name: "legacy"}},
	}
	latest := makeBench("v2", "Bench", "2026-05-14T10:00:00Z",
		[]DataPoint{{Name: "", XAxis: "speed", YAxis: "1e4"}})
	latest.Theme = "roma"

	result := MergeDatasets([]Dataset{legacy, latest}, DimensionName)
	s.Require().Len(result, 1)
	s.Equal("roma", result[0].Theme)

	updated := latest
	updated.Timestamp = "2026-05-15T10:00:00Z"
	updated.Theme = "chalk"
	result = MergeDatasets([]Dataset{result[0], updated}, DimensionName)
	s.Require().Len(result, 1)
	s.Equal("chalk", result[0].Theme)
}

func (s *MergeSuite) TestMergePreservesThemeWhenIncomingLegacyDatasetHasNone() {
	existing := makeBench("v1", "Bench", "2026-05-13T10:00:00Z",
		[]DataPoint{{Name: "", XAxis: "speed", YAxis: "1e4"}})
	existing.Theme = "walden"
	incoming := makeBench("v1", "Bench", "2026-05-14T10:00:00Z",
		[]DataPoint{{Name: "", XAxis: "speed", YAxis: "2e4"}})

	result := MergeDatasets([]Dataset{existing, incoming}, DimensionName)
	s.Require().Len(result, 1)
	s.Equal("walden", result[0].Theme)
}

func (s *MergeSuite) TestMergeDatasetsSameNameNoTagDedup() {
	bench1 := Dataset{Name: "Bench", Data: []DataPoint{{Name: "first"}}}
	bench2 := Dataset{Name: "Bench", Data: []DataPoint{{Name: "second"}}}

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	s.Len(result, 1)
	s.Equal("first", result[0].Data[0].Name)
}

func (s *MergeSuite) TestMergeDatasetsTagOrderChronological() {
	bench1 := makeBench("v1", "Bench", "2026-05-13T10:00:00Z",
		[]DataPoint{{Name: "", XAxis: "speed", YAxis: "1e4"}})
	bench2 := makeBench("v2", "Bench", "2026-05-14T10:00:00Z",
		[]DataPoint{{Name: "", XAxis: "speed", YAxis: "1e4"}})
	bench3 := makeBench("v3", "Bench", "2026-05-12T10:00:00Z",
		[]DataPoint{{Name: "", XAxis: "speed", YAxis: "1e4"}})

	result := MergeDatasets([]Dataset{bench1, bench2, bench3}, DimensionName)
	s.Len(result, 1)
	s.Equal("v2", result[0].Tag)
	s.Equal("2026-05-14T10:00:00Z", result[0].Timestamp)
	s.Len(result[0].Data, 3)
	s.Equal("v3", result[0].Data[0].Name)
	s.Equal("v1", result[0].Data[1].Name)
	s.Equal("v2", result[0].Data[2].Name)
	s.Equal([]HistoryEntry{
		{Tag: "v3", Timestamp: "2026-05-12T10:00:00Z"},
		{Tag: "v1", Timestamp: "2026-05-13T10:00:00Z"},
	}, result[0].History)
}

func (s *MergeSuite) TestMergeReplaceSameTagPreservesOlderVersions() {
	accumulated := makeBench("v1.8.0", "Bench", "2026-07-04T05:09:39Z", []DataPoint{
		{Name: "Add", XAxis: "v1.3.0", YAxis: "Queue"},
		{Name: "Add", XAxis: "v1.7.1", YAxis: "Queue"},
		{Name: "Add", XAxis: "v1.8.0", YAxis: "Queue", Stats: []Stat{{Type: "time", Value: F64(99)}}},
	})
	accumulated.History = []HistoryEntry{
		{Tag: "v1.3.0", Timestamp: "2026-05-25T08:44:05Z"},
		{Tag: "v1.7.1", Timestamp: "2026-05-25T09:21:58Z"},
	}

	incoming := makeBench("v1.8.0", "Bench", "2026-07-04T05:33:48Z", []DataPoint{
		{Name: "Add", XAxis: "", YAxis: "Queue", Stats: []Stat{{Type: "time", Value: F64(2141)}}},
		{Name: "AddAll", XAxis: "", YAxis: "Queue", Stats: []Stat{{Type: "time", Value: F64(1489434)}}},
	})

	result := MergeDatasets([]Dataset{incoming, accumulated}, DimensionXAxis)
	s.Require().Len(result, 1)
	s.Equal("v1.8.0", result[0].Tag)
	s.Equal("2026-07-04T05:33:48Z", result[0].Timestamp)
	s.Equal([]HistoryEntry{
		{Tag: "v1.3.0", Timestamp: "2026-05-25T08:44:05Z"},
		{Tag: "v1.7.1", Timestamp: "2026-05-25T09:21:58Z"},
	}, result[0].History)

	s.Require().Len(result[0].Data, 4)
	s.Equal("v1.3.0", result[0].Data[0].XAxis)
	s.Equal("v1.7.1", result[0].Data[1].XAxis)
	s.Equal("v1.8.0", result[0].Data[2].XAxis)
	s.Equal(2141.0, *result[0].Data[2].Stats[0].Value)
	s.Equal("v1.8.0", result[0].Data[3].XAxis)
	s.Equal("AddAll", result[0].Data[3].Name)
}

func (s *MergeSuite) TestMergeNoTagNewerReplacesOlder() {
	bench1 := Dataset{
		Name:      "Bench",
		Timestamp: "2026-01-01T00:00:00Z",
		Data:      []DataPoint{{Name: "first"}},
	}
	bench2 := Dataset{
		Name:      "Bench",
		Timestamp: "2026-02-01T00:00:00Z",
		Data:      []DataPoint{{Name: "second"}},
	}

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	s.Require().Len(result, 1)
	s.Equal("second", result[0].Data[0].Name)
}

func (s *MergeSuite) TestPickAccumulatedBasePrefersMoreHistory() {
	withHistory := Dataset{History: []HistoryEntry{{Tag: "v1", Timestamp: "t1"}}}
	withoutHistory := Dataset{Data: []DataPoint{{Name: "a"}, {Name: "b"}}}

	s.Equal(withHistory, pickAccumulatedBase(withoutHistory, withHistory))
	s.Equal(withHistory, pickAccumulatedBase(withHistory, withoutHistory))
}

func (s *MergeSuite) TestPickAccumulatedBasePrefersMoreDataWhenHistoryEqual() {
	smaller := Dataset{Data: []DataPoint{{Name: "one"}}}
	larger := Dataset{Data: []DataPoint{{Name: "one"}, {Name: "two"}}}

	s.Equal(larger, pickAccumulatedBase(smaller, larger))
	s.Equal(larger, pickAccumulatedBase(larger, smaller))
}

func (s *MergeSuite) TestReplaceTagDataUpdatesMeta() {
	existing := makeBench("v1", "Bench", "2026-01-01T00:00:00Z", []DataPoint{
		{Name: "Add", XAxis: "v1", YAxis: "Queue"},
	})
	incoming := makeBench("v1", "Bench", "2026-02-01T00:00:00Z", []DataPoint{
		{Name: "Add", XAxis: "", YAxis: "Queue"},
	})
	incoming.Meta = &Meta{
		CPU:  &CPUInfo{Name: "Apple M2", Cores: 10},
		OS:   "darwin",
		Arch: "arm64",
		Pkg:  "github.com/example/pkg",
	}

	result := replaceTagData(existing, incoming, DimensionXAxis)
	s.Equal("2026-02-01T00:00:00Z", result.Timestamp)
	s.Require().NotNil(result.Meta)
	s.Require().NotNil(result.Meta.CPU)
	s.Equal("Apple M2", result.Meta.CPU.Name)
	s.Equal("darwin", result.Meta.OS)
	s.NotSame(incoming.Meta.CPU, result.Meta.CPU)
}

func (s *MergeSuite) TestMergeReplaceSameTagOnYAxis() {
	accumulated := makeBench("v2", "Bench", "2026-07-04T05:09:39Z", []DataPoint{
		{Name: "Add", XAxis: "Queue", YAxis: "v1"},
		{Name: "Add", XAxis: "Queue", YAxis: "v2", Stats: []Stat{{Type: "time", Value: F64(99)}}},
	})
	incoming := makeBench("v2", "Bench", "2026-07-04T05:33:48Z", []DataPoint{
		{Name: "Add", XAxis: "Queue", YAxis: "", Stats: []Stat{{Type: "time", Value: F64(200)}}},
	})

	result := MergeDatasets([]Dataset{incoming, accumulated}, DimensionYAxis)
	s.Require().Len(result, 1)
	s.Require().Len(result[0].Data, 2)
	s.Equal("v1", result[0].Data[0].YAxis)
	s.Equal("v2", result[0].Data[1].YAxis)
	s.Equal(200.0, *result[0].Data[1].Stats[0].Value)
}

func (s *MergeSuite) TestMergeReplaceSameTagOnZAxis() {
	accumulated := makeBench("v2", "Bench", "2026-07-04T05:09:39Z", []DataPoint{
		{Name: "Add", XAxis: "x", YAxis: "y", ZAxis: "v1"},
		{Name: "Add", XAxis: "x", YAxis: "y", ZAxis: "v2"},
	})
	incoming := makeBench("v2", "Bench", "2026-07-04T05:33:48Z", []DataPoint{
		{Name: "Add", XAxis: "x", YAxis: "y", ZAxis: ""},
	})

	result := MergeDatasets([]Dataset{incoming, accumulated}, DimensionZAxis)
	s.Require().Len(result, 1)
	s.Require().Len(result[0].Data, 2)
	s.Equal("v1", result[0].Data[0].ZAxis)
	s.Equal("v2", result[0].Data[1].ZAxis)
}

func (s *MergeSuite) TestMergeDataForTagKeepsPreInjectedPoints() {
	ds := makeBench("v1.8.0", "Bench", "2026-07-04T05:33:48Z", []DataPoint{
		{Name: "Add", XAxis: "v1.8.0", YAxis: "Queue", Stats: []Stat{{Type: "time", Value: F64(42)}}},
		{Name: "Add", XAxis: "v1.3.0", YAxis: "Queue"},
	})

	points := mergeDataForTag(ds, DimensionXAxis)
	s.Require().Len(points, 1)
	s.Equal("v1.8.0", points[0].XAxis)
	s.Equal(42.0, *points[0].Stats[0].Value)
}

func (s *MergeSuite) TestDataPointBelongsToTag() {
	s.True(dataPointBelongsToTag(DataPoint{XAxis: ""}, "v1", DimensionXAxis))
	s.True(dataPointBelongsToTag(DataPoint{XAxis: "v1"}, "v1", DimensionXAxis))
	s.False(dataPointBelongsToTag(DataPoint{XAxis: "v2"}, "v1", DimensionXAxis))
	s.True(dataPointBelongsToTag(DataPoint{YAxis: "v1"}, "v1", DimensionYAxis))
	s.True(dataPointBelongsToTag(DataPoint{ZAxis: "v1"}, "v1", DimensionZAxis))
	s.True(dataPointBelongsToTag(DataPoint{Name: "v1"}, "v1", DimensionName))
}

func (s *MergeSuite) TestMergeReplaceSameTagSkipsOlderIncoming() {
	accumulated := makeBench("v1.8.0", "Bench", "2026-07-04T05:33:48Z", []DataPoint{
		{Name: "Add", XAxis: "v1.3.0", YAxis: "Queue"},
		{Name: "Add", XAxis: "v1.8.0", YAxis: "Queue", Stats: []Stat{{Type: "time", Value: F64(2141)}}},
	})
	accumulated.History = []HistoryEntry{{Tag: "v1.3.0", Timestamp: "2026-05-25T08:44:05Z"}}

	older := makeBench("v1.8.0", "Bench", "2026-07-04T05:09:39Z", []DataPoint{
		{Name: "Add", XAxis: "", YAxis: "Queue", Stats: []Stat{{Type: "time", Value: F64(99)}}},
	})

	result := MergeDatasets([]Dataset{older, accumulated}, DimensionXAxis)
	s.Require().Len(result, 1)
	s.Equal("2026-07-04T05:33:48Z", result[0].Timestamp)
	s.Require().Len(result[0].Data, 2)
	s.Equal("v1.3.0", result[0].Data[0].XAxis)
	s.Equal(2141.0, *result[0].Data[1].Stats[0].Value)
	s.Equal([]HistoryEntry{
		{Tag: "v1.3.0", Timestamp: "2026-05-25T08:44:05Z"},
	}, result[0].History)
}

func (s *MergeSuite) TestEnsureInjectAxisEmptyAxes() {
	s.Equal([]string{"x"}, axisKeys(EnsureAxis(nil, DimensionXAxis)))
	s.Equal([]string{"y"}, axisKeys(EnsureAxis([]Axis{}, DimensionYAxis)))
}

func (s *MergeSuite) TestEnsureInjectAxisInsertsAmongNonCanonicalAxes() {
	axes := []Axis{{Key: "custom"}, {Key: "y"}}
	result := EnsureAxis(axes, DimensionXAxis)
	s.Equal([]string{"custom", "x", "y"}, axisKeys(result))
}

func (s *MergeSuite) TestEnsureInjectAxisNoTagMergePath() {
	noTag := Dataset{
		Name: "Bench",
		Data: []DataPoint{{Name: "legacy", XAxis: "x", YAxis: "y"}},
		Axes: []Axis{{Key: "name"}, {Key: "y"}},
	}
	tagged := makeBench("v1", "Bench", "2026-05-13T10:00:00Z", []DataPoint{
		{Name: "", XAxis: "speed", YAxis: "100"},
	})
	tagged.Axes = []Axis{{Key: "name"}, {Key: "y"}}

	result := MergeDatasets([]Dataset{noTag, tagged}, DimensionXAxis)
	s.Require().Len(result, 1)
	s.Equal([]string{"name", "x", "y"}, axisKeys(result[0].Axes))
}

func TestMergeSuite(t *testing.T) {
	suite.Run(t, new(MergeSuite))
}
