package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeBench(tag, name, timestamp string, data []BenchmarkData) Benchmark {
	return Benchmark{
		Tag:       tag,
		Timestamp: timestamp,
		Name:      name,
		Data:      data,
	}
}

func TestMergeBenchmarks_SmartMerge(t *testing.T) {
	bench1 := makeBench("1", "My benchmark", "2026-05-13T10:00:00Z", []BenchmarkData{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})
	bench2 := makeBench("2", "My benchmark", "2026-05-13T10:05:00Z", []BenchmarkData{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionName)
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

func TestMergeBenchmarks_MixedGroup(t *testing.T) {
	bench1 := makeBench("1", "My benchmark", "2026-05-13T10:00:00Z", []BenchmarkData{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})
	bench2 := makeBench("2", "My benchmark", "2026-05-13T10:05:00Z", []BenchmarkData{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})
	noTagBench := Benchmark{
		Tag:  "",
		Name: "My benchmark",
		Data: []BenchmarkData{{Name: "legacy", XAxis: "x", YAxis: "y"}},
	}

	result := MergeBenchmarks([]Benchmark{bench1, bench2, noTagBench}, DimensionName)
	assert.Len(t, result, 1)

	merged := result[0]
	assert.Len(t, merged.Data, 3)
	assert.Equal(t, "legacy", merged.Data[0].Name)
	assert.Equal(t, "1", merged.Data[1].Name)
	assert.Equal(t, "2", merged.Data[2].Name)
	assert.Equal(t, "2", merged.Tag)
	assert.Equal(t, "2026-05-13T10:05:00Z", merged.Timestamp)
}

func TestMergeBenchmarks_AllNoTag(t *testing.T) {
	bench1 := Benchmark{Name: "Bench A", Data: []BenchmarkData{{Name: "a"}}}
	bench2 := Benchmark{Name: "Bench A", Data: []BenchmarkData{{Name: "b"}}}

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)
	assert.Equal(t, "a", result[0].Data[0].Name)
}

func TestMergeBenchmarks_TimestampTie(t *testing.T) {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []BenchmarkData{{Name: "a"}})
	bench2 := makeBench("2", "Test", "2026-05-13T10:00:00Z", []BenchmarkData{{Name: "b"}})

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)
	assert.Len(t, result[0].Data, 2)
	assert.ElementsMatch(t, []string{"a", "b"}, []string{result[0].Data[0].Name, result[0].Data[1].Name})
	assert.Contains(t, []string{"1", "2"}, result[0].Tag)
}

func TestMergeBenchmarks_SingleBenchmark(t *testing.T) {
	bench := makeBench("1", "Solo", "2026-05-13T10:00:00Z", []BenchmarkData{{Name: "x"}})
	result := MergeBenchmarks([]Benchmark{bench}, DimensionName)
	assert.Len(t, result, 1)
	assert.Equal(t, "Solo", result[0].Name)
	assert.Equal(t, "2026-05-13T10:00:00Z", result[0].Timestamp)
	assert.Nil(t, result[0].History)
}

func TestMergeBenchmarks_PopulatedName(t *testing.T) {
	bench1 := makeBench("1", "digits", "2026-05-13T10:00:00Z", []BenchmarkData{
		{Name: "digits", XAxis: "speed", YAxis: "1e4"},
	})
	bench2 := makeBench("2", "digits", "2026-05-13T10:05:00Z", []BenchmarkData{
		{Name: "digits", XAxis: "speed", YAxis: "1e4"},
	})

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)
	assert.Equal(t, "digits", result[0].Data[0].Name)
	assert.Equal(t, "digits", result[0].Data[1].Name)
}

func TestMergeBenchmarks_InjectDimensionX(t *testing.T) {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []BenchmarkData{
		{XAxis: "", YAxis: "100"},
	})
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []BenchmarkData{
		{XAxis: "", YAxis: "200"},
	})

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionXAxis)
	assert.Len(t, result, 1)
	assert.Equal(t, "1", result[0].Data[0].XAxis)
	assert.Equal(t, "2", result[0].Data[1].XAxis)
	assert.Equal(t, "", result[0].Data[0].Name)
}

func TestMergeBenchmarks_InjectDimensionY(t *testing.T) {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []BenchmarkData{
		{XAxis: "x", YAxis: ""},
	})
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []BenchmarkData{
		{XAxis: "x", YAxis: ""},
	})

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionYAxis)
	assert.Len(t, result, 1)
	assert.Equal(t, "1", result[0].Data[0].YAxis)
	assert.Equal(t, "2", result[0].Data[1].YAxis)
}

func TestMergeBenchmarks_HistoryMerge(t *testing.T) {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []BenchmarkData{{Name: "a"}})
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []BenchmarkData{{Name: "b"}})
	bench2.History = []HistoryEntry{
		{Tag: "extra", Timestamp: "2026-05-13T11:00:00Z"},
	}

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)
	assert.Equal(t, "2", result[0].Tag)
	assert.Equal(t, "2026-05-13T10:05:00Z", result[0].Timestamp)
	assert.Equal(t, []HistoryEntry{
		{Tag: "1", Timestamp: "2026-05-13T10:00:00Z"},
		{Tag: "extra", Timestamp: "2026-05-13T11:00:00Z"},
	}, result[0].History)
}

func TestMergeBenchmarks_DifferentNames(t *testing.T) {
	bench1 := makeBench("1", "Bench A", "2026-05-13T10:00:00Z", []BenchmarkData{{Name: "a"}})
	bench2 := makeBench("2", "Bench B", "2026-05-13T10:05:00Z", []BenchmarkData{{Name: "b"}})

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionName)
	assert.Len(t, result, 2)
}

func TestMergeBenchmarks_EmptyInput(t *testing.T) {
	result := MergeBenchmarks(nil, DimensionName)
	assert.Empty(t, result)
}

func TestMergeBenchmarks_NoMutation(t *testing.T) {
	bench1 := makeBench("1", "Test", "2026-05-13T10:00:00Z", []BenchmarkData{{Name: "a"}})
	bench2 := makeBench("2", "Test", "2026-05-13T10:05:00Z", []BenchmarkData{{Name: "b"}})

	original := []Benchmark{bench1, bench2}
	result := MergeBenchmarks(original, DimensionName)

	assert.Equal(t, "a", original[0].Data[0].Name)
	assert.Equal(t, "b", original[1].Data[0].Name)
	assert.Len(t, result, 1)
}

func TestMergeBenchmarks_SameNameSameTagDedup(t *testing.T) {
	bench1 := makeBench("v1", "Bench", "2026-05-13T10:00:00Z",
		[]BenchmarkData{{Name: "", XAxis: "speed", YAxis: "1e4"}})
	bench2 := makeBench("v1", "Bench", "2026-05-14T10:00:00Z",
		[]BenchmarkData{{Name: "", XAxis: "speed", YAxis: "1e4"}})

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)
	assert.Len(t, result[0].Data, 1)
	assert.Equal(t, "v1", result[0].Data[0].Name)
	assert.Equal(t, "v1", result[0].Tag)
	assert.Equal(t, "2026-05-14T10:00:00Z", result[0].Timestamp)
	assert.Nil(t, result[0].History)
}

func TestMergeBenchmarks_SameNameNoTagDedup(t *testing.T) {
	bench1 := Benchmark{Name: "Bench", Data: []BenchmarkData{{Name: "first"}}}
	bench2 := Benchmark{Name: "Bench", Data: []BenchmarkData{{Name: "second"}}}

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)
	assert.Equal(t, "first", result[0].Data[0].Name)
}

func TestMergeBenchmarks_TagOrderChronological(t *testing.T) {
	bench1 := makeBench("v1", "Bench", "2026-05-13T10:00:00Z",
		[]BenchmarkData{{Name: "", XAxis: "speed", YAxis: "1e4"}})
	bench2 := makeBench("v2", "Bench", "2026-05-14T10:00:00Z",
		[]BenchmarkData{{Name: "", XAxis: "speed", YAxis: "1e4"}})
	bench3 := makeBench("v3", "Bench", "2026-05-12T10:00:00Z",
		[]BenchmarkData{{Name: "", XAxis: "speed", YAxis: "1e4"}})

	result := MergeBenchmarks([]Benchmark{bench1, bench2, bench3}, DimensionName)
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
