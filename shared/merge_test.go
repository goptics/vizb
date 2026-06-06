package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeBench(tag, name, timestamp string, data []DataPoint) Dataset {
	return Dataset{
		Tag:       tag,
		Timestamp: timestamp,
		Name:      name,
		Data:      data,
	}
}

func TestMergeDatasets_SmartMerge(t *testing.T) {
	bench1 := makeBench("1", "My benchmark", "2026-05-13T10:00:00Z", []DataPoint{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})
	bench2 := makeBench("2", "My benchmark", "2026-05-13T10:05:00Z", []DataPoint{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)

	merged := result[0]
	assert.Equal(t, "2", merged.Tag)
	assert.Equal(t, "My benchmark", merged.Name)
	assert.Equal(t, "2026-05-13T10:05:00Z", merged.Timestamp)
	assert.Equal(t, []HistoryEntry{
		{Tag: "1", Timestamp: "2026-05-13T10:00:00Z"},
	}, merged.History)
	assert.Len(t, merged.Data, 2)
	assert.Equal(t, "1", merged.Data[0].Name)
	assert.Equal(t, "2", merged.Data[1].Name)
}

func TestMergeDatasets_MixedGroup(t *testing.T) {
	bench1 := makeBench("1", "My benchmark", "2026-05-13T10:00:00Z", []DataPoint{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})
	bench2 := makeBench("2", "My benchmark", "2026-05-13T10:05:00Z", []DataPoint{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})
	noTagBench := Dataset{
		Tag:  "",
		Name: "My benchmark",
		Data: []DataPoint{{Name: "legacy", XAxis: "x", YAxis: "y"}},
	}

	result := MergeDatasets([]Dataset{bench1, bench2, noTagBench}, DimensionName)
	assert.Len(t, result, 1)

	merged := result[0]
	assert.Len(t, merged.Data, 3)
	assert.Equal(t, "legacy", merged.Data[0].Name)
	assert.Equal(t, "1", merged.Data[1].Name)
	assert.Equal(t, "2", merged.Data[2].Name)
	assert.Equal(t, "2", merged.Tag)
	assert.Equal(t, "2026-05-13T10:05:00Z", merged.Timestamp)
}

func TestMergeDatasets_AllNoTag(t *testing.T) {
	bench1 := Dataset{Name: "Bench A", Data: []DataPoint{{Name: "a"}}}
	bench2 := Dataset{Name: "Bench A", Data: []DataPoint{{Name: "b"}}}

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)
	assert.Equal(t, "a", result[0].Data[0].Name)
}

func TestMergeDatasets_TimestampTie(t *testing.T) {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{{Name: "a"}})
	bench2 := makeBench("2", "Test", "2026-05-13T10:00:00Z", []DataPoint{{Name: "b"}})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)
	assert.Len(t, result[0].Data, 2)
	assert.ElementsMatch(t, []string{"a", "b"}, []string{result[0].Data[0].Name, result[0].Data[1].Name})
	assert.Contains(t, []string{"1", "2"}, result[0].Tag)
}

func TestMergeDatasets_SingleBenchmark(t *testing.T) {
	bench := makeBench("1", "Solo", "2026-05-13T10:00:00Z", []DataPoint{{Name: "x"}})
	result := MergeDatasets([]Dataset{bench}, DimensionName)
	assert.Len(t, result, 1)
	assert.Equal(t, "Solo", result[0].Name)
	assert.Equal(t, "2026-05-13T10:00:00Z", result[0].Timestamp)
	assert.Nil(t, result[0].History)
}

func TestMergeDatasets_PopulatedName(t *testing.T) {
	bench1 := makeBench("1", "digits", "2026-05-13T10:00:00Z", []DataPoint{
		{Name: "digits", XAxis: "speed", YAxis: "1e4"},
	})
	bench2 := makeBench("2", "digits", "2026-05-13T10:05:00Z", []DataPoint{
		{Name: "digits", XAxis: "speed", YAxis: "1e4"},
	})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)
	assert.Equal(t, "digits", result[0].Data[0].Name)
	assert.Equal(t, "digits", result[0].Data[1].Name)
}

func TestMergeDatasets_InjectDimensionX(t *testing.T) {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{
		{XAxis: "", YAxis: "100"},
	})
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []DataPoint{
		{XAxis: "", YAxis: "200"},
	})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionXAxis)
	assert.Len(t, result, 1)
	assert.Equal(t, "1", result[0].Data[0].XAxis)
	assert.Equal(t, "2", result[0].Data[1].XAxis)
	assert.Equal(t, "", result[0].Data[0].Name)
}

func TestMergeDatasets_InjectDimensionY(t *testing.T) {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{
		{XAxis: "x", YAxis: ""},
	})
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []DataPoint{
		{XAxis: "x", YAxis: ""},
	})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionYAxis)
	assert.Len(t, result, 1)
	assert.Equal(t, "1", result[0].Data[0].YAxis)
	assert.Equal(t, "2", result[0].Data[1].YAxis)
}

func TestMergeDatasets_HistoryMerge(t *testing.T) {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{{Name: "a"}})
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []DataPoint{{Name: "b"}})
	bench2.History = []HistoryEntry{
		{Tag: "extra", Timestamp: "2026-05-13T11:00:00Z"},
	}

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)
	assert.Equal(t, "2", result[0].Tag)
	assert.Equal(t, "2026-05-13T10:05:00Z", result[0].Timestamp)
	assert.Equal(t, []HistoryEntry{
		{Tag: "1", Timestamp: "2026-05-13T10:00:00Z"},
		{Tag: "extra", Timestamp: "2026-05-13T11:00:00Z"},
	}, result[0].History)
}

func TestMergeDatasets_DifferentNames(t *testing.T) {
	bench1 := makeBench("1", "Bench A", "2026-05-13T10:00:00Z", []DataPoint{{Name: "a"}})
	bench2 := makeBench("2", "Bench B", "2026-05-13T10:05:00Z", []DataPoint{{Name: "b"}})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	assert.Len(t, result, 2)
}

func TestMergeDatasets_EmptyInput(t *testing.T) {
	result := MergeDatasets(nil, DimensionName)
	assert.Empty(t, result)
}

func TestMergeDatasets_NoMutation(t *testing.T) {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []DataPoint{{Name: "a"}})
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []DataPoint{{Name: "b"}})

	original := []Dataset{bench1, bench2}
	result := MergeDatasets(original, DimensionName)

	assert.Equal(t, "a", original[0].Data[0].Name)
	assert.Equal(t, "b", original[1].Data[0].Name)
	assert.Len(t, result, 1)
}

func TestMergeDatasets_SameNameSameTagDedup(t *testing.T) {
	bench1 := makeBench("v1", "Bench", "2026-05-13T10:00:00Z",
		[]DataPoint{{Name: "", XAxis: "speed", YAxis: "1e4"}})
	bench2 := makeBench("v1", "Bench", "2026-05-14T10:00:00Z",
		[]DataPoint{{Name: "", XAxis: "speed", YAxis: "1e4"}})

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)
	assert.Len(t, result[0].Data, 1)
	assert.Equal(t, "v1", result[0].Data[0].Name)
	assert.Equal(t, "v1", result[0].Tag)
	assert.Equal(t, "2026-05-14T10:00:00Z", result[0].Timestamp)
	assert.Nil(t, result[0].History)
}

func TestMergeDatasets_SameNameNoTagDedup(t *testing.T) {
	bench1 := Dataset{Name: "Bench", Data: []DataPoint{{Name: "first"}}}
	bench2 := Dataset{Name: "Bench", Data: []DataPoint{{Name: "second"}}}

	result := MergeDatasets([]Dataset{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)
	assert.Equal(t, "first", result[0].Data[0].Name)
}

func TestMergeDatasets_TagOrderChronological(t *testing.T) {
	bench1 := makeBench("v1", "Bench", "2026-05-13T10:00:00Z",
		[]DataPoint{{Name: "", XAxis: "speed", YAxis: "1e4"}})
	bench2 := makeBench("v2", "Bench", "2026-05-14T10:00:00Z",
		[]DataPoint{{Name: "", XAxis: "speed", YAxis: "1e4"}})
	bench3 := makeBench("v3", "Bench", "2026-05-12T10:00:00Z",
		[]DataPoint{{Name: "", XAxis: "speed", YAxis: "1e4"}})

	result := MergeDatasets([]Dataset{bench1, bench2, bench3}, DimensionName)
	assert.Len(t, result, 1)
	assert.Equal(t, "v2", result[0].Tag)
	assert.Equal(t, "2026-05-14T10:00:00Z", result[0].Timestamp)
	assert.Len(t, result[0].Data, 3)
	assert.Equal(t, "v3", result[0].Data[0].Name)
	assert.Equal(t, "v1", result[0].Data[1].Name)
	assert.Equal(t, "v2", result[0].Data[2].Name)
	assert.Equal(t, []HistoryEntry{
		{Tag: "v3", Timestamp: "2026-05-12T10:00:00Z"},
		{Tag: "v1", Timestamp: "2026-05-13T10:00:00Z"},
	}, result[0].History)
}
