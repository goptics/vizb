package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeBench(tag, name string, runtimes map[string]string, data []BenchmarkData) Benchmark {
	return Benchmark{
		Tag:      tag,
		Name:     name,
		Runtimes: runtimes,
		Data:     data,
	}
}

func TestMergeBenchmarks_SmartMerge(t *testing.T) {
	bench1 := makeBench("1", "My benchmark", map[string]string{"1": "2026-05-13T10:00:00Z"}, []BenchmarkData{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})
	bench2 := makeBench("2", "My benchmark", map[string]string{"2": "2026-05-13T10:05:00Z"}, []BenchmarkData{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)

	merged := result[0]
	assert.Empty(t, merged.Tag)
	assert.Equal(t, "My benchmark", merged.Name)
	assert.Equal(t, map[string]string{
		"1": "2026-05-13T10:00:00Z",
		"2": "2026-05-13T10:05:00Z",
	}, merged.Runtimes)
	assert.Len(t, merged.Data, 2)
	assert.Equal(t, "1", merged.Data[0].Name)
	assert.Equal(t, "2", merged.Data[1].Name)
}

func TestMergeBenchmarks_MixedGroup(t *testing.T) {
	bench1 := makeBench("1", "My benchmark", map[string]string{"1": "2026-05-13T10:00:00Z"}, []BenchmarkData{
		{Name: "", XAxis: "speed", YAxis: "1e4"},
	})
	bench2 := makeBench("2", "My benchmark", map[string]string{"2": "2026-05-13T10:05:00Z"}, []BenchmarkData{
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
	assert.Empty(t, merged.Tag)
}

func TestMergeBenchmarks_AllNoTag(t *testing.T) {
	bench1 := Benchmark{Name: "Bench A", Data: []BenchmarkData{{Name: "a"}}}
	bench2 := Benchmark{Name: "Bench A", Data: []BenchmarkData{{Name: "b"}}}

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionName)
	assert.Len(t, result, 2)
}

func TestMergeBenchmarks_TimestampTie(t *testing.T) {
	bench1 := makeBench("1", "Test", map[string]string{"1": "2026-05-13T10:00:00Z"}, []BenchmarkData{{Name: "a"}})
	bench2 := makeBench("2", "Test", map[string]string{"2": "2026-05-13T10:00:00Z"}, []BenchmarkData{{Name: "b"}})

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionName)
	assert.Len(t, result, 2)
}

func TestMergeBenchmarks_SingleBenchmark(t *testing.T) {
	bench := makeBench("1", "Solo", map[string]string{"1": "2026-05-13T10:00:00Z"}, []BenchmarkData{{Name: "x"}})
	result := MergeBenchmarks([]Benchmark{bench}, DimensionName)
	assert.Len(t, result, 1)
	assert.Equal(t, "Solo", result[0].Name)
}

func TestMergeBenchmarks_PopulatedName(t *testing.T) {
	bench1 := makeBench("1", "digits", map[string]string{"1": "2026-05-13T10:00:00Z"}, []BenchmarkData{
		{Name: "digits", XAxis: "speed", YAxis: "1e4"},
	})
	bench2 := makeBench("2", "digits", map[string]string{"2": "2026-05-13T10:05:00Z"}, []BenchmarkData{
		{Name: "digits", XAxis: "speed", YAxis: "1e4"},
	})

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)
	assert.Equal(t, "digits", result[0].Data[0].Name)
	assert.Equal(t, "digits", result[0].Data[1].Name)
}

func TestMergeBenchmarks_InjectDimensionX(t *testing.T) {
	bench1 := makeBench("1", "Test", map[string]string{"1": "2026-05-13T10:00:00Z"}, []BenchmarkData{
		{XAxis: "", YAxis: "100"},
	})
	bench2 := makeBench("2", "Test", map[string]string{"2": "2026-05-13T10:05:00Z"}, []BenchmarkData{
		{XAxis: "", YAxis: "200"},
	})

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionXAxis)
	assert.Len(t, result, 1)
	assert.Equal(t, "1", result[0].Data[0].XAxis)
	assert.Equal(t, "2", result[0].Data[1].XAxis)
	assert.Equal(t, "", result[0].Data[0].Name)
}

func TestMergeBenchmarks_InjectDimensionY(t *testing.T) {
	bench1 := makeBench("1", "Test", map[string]string{"1": "2026-05-13T10:00:00Z"}, []BenchmarkData{
		{XAxis: "x", YAxis: ""},
	})
	bench2 := makeBench("2", "Test", map[string]string{"2": "2026-05-13T10:05:00Z"}, []BenchmarkData{
		{XAxis: "x", YAxis: ""},
	})

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionYAxis)
	assert.Len(t, result, 1)
	assert.Equal(t, "1", result[0].Data[0].YAxis)
	assert.Equal(t, "2", result[0].Data[1].YAxis)
}

func TestMergeBenchmarks_RuntimeMerge(t *testing.T) {
	bench1 := makeBench("1", "Test", map[string]string{"1": "2026-05-13T10:00:00Z"}, []BenchmarkData{{Name: "a"}})
	bench2 := makeBench("2", "Test", map[string]string{"2": "2026-05-13T10:05:00Z", "extra": "2026-05-13T11:00:00Z"}, []BenchmarkData{{Name: "b"}})

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionName)
	assert.Len(t, result, 1)
	assert.Len(t, result[0].Runtimes, 3)
	assert.Equal(t, "2026-05-13T10:00:00Z", result[0].Runtimes["1"])
	assert.Equal(t, "2026-05-13T10:05:00Z", result[0].Runtimes["2"])
	assert.Equal(t, "2026-05-13T11:00:00Z", result[0].Runtimes["extra"])
}

func TestMergeBenchmarks_DifferentNames(t *testing.T) {
	bench1 := makeBench("1", "Bench A", map[string]string{"1": "2026-05-13T10:00:00Z"}, []BenchmarkData{{Name: "a"}})
	bench2 := makeBench("2", "Bench B", map[string]string{"2": "2026-05-13T10:05:00Z"}, []BenchmarkData{{Name: "b"}})

	result := MergeBenchmarks([]Benchmark{bench1, bench2}, DimensionName)
	assert.Len(t, result, 2)
}

func TestMergeBenchmarks_EmptyInput(t *testing.T) {
	result := MergeBenchmarks(nil, DimensionName)
	assert.Empty(t, result)
}

func TestMergeBenchmarks_NoMutation(t *testing.T) {
	bench1 := makeBench("1", "Test", map[string]string{"1": "2026-05-13T10:00:00Z"}, []BenchmarkData{{Name: "a"}})
	bench2 := makeBench("2", "Test", map[string]string{"2": "2026-05-13T10:05:00Z"}, []BenchmarkData{{Name: "b"}})

	original := []Benchmark{bench1, bench2}
	result := MergeBenchmarks(original, DimensionName)

	assert.Equal(t, "a", original[0].Data[0].Name)
	assert.Equal(t, "b", original[1].Data[0].Name)
	assert.Len(t, result, 1)
}
