package core

import (
	"errors"
	"math"
	"strings"
	"sync"
	"testing"

	internalcharts "github.com/goptics/vizb/internal/charts"
	barchart "github.com/goptics/vizb/internal/charts/bar"
	linechart "github.com/goptics/vizb/internal/charts/line"
	scatterchart "github.com/goptics/vizb/internal/charts/scatter"
	"github.com/goptics/vizb/internal/flags"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

type CoreSuite struct{ suite.Suite }

func (s *CoreSuite) TestConvertCSV() {
	result, err := Convert(ConvertInput{
		Input:    []byte("region,latency\nwest,12\neast,18\n"),
		Parser:   "csv",
		Config:   parser.Config{GroupPattern: "x", Group: []string{"region"}},
		Metadata: Metadata{Name: "API latency", Theme: "default"},
		Charts:   []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Scale: "linear"}},
	})
	s.Require().NoError(err)
	s.Equal("API latency", result.Dataset.Name)
	s.Len(result.Dataset.Data, 2)
	s.Equal([]string{"x"}, []string{result.Dataset.Axes[0].Key})
}

func (s *CoreSuite) TestConvertJSONAndValidationFailures() {
	result, err := Convert(ConvertInput{
		Input:  []byte(`[{"region":"west","latency":12},{"region":"east","latency":18}]`),
		Parser: "json",
		Config: parser.Config{GroupPattern: "x", Group: []string{"region"}},
		Charts: []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Scale: "linear"}},
	})
	s.Require().NoError(err)
	s.Len(result.Dataset.Data, 2)

	_, err = Convert(ConvertInput{
		Input:  []byte("region,latency\nwest,12\n"),
		Parser: "csv",
		Config: parser.Config{GroupPattern: "x", Group: []string{"missing"}},
		Charts: []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Scale: "linear"}},
	})
	s.ErrorContains(err, `group column "missing" not found`)

	_, err = Convert(ConvertInput{
		Input:  []byte("region,latency\nwest,12\n"),
		Parser: "csv",
		Config: parser.Config{Axes: []parser.ColumnSpec{{Source: "missing"}}},
		Charts: []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Scale: "linear"}},
	})
	s.ErrorContains(err, `--axes column "missing" not found`)
}

func (s *CoreSuite) TestOperations() {
	chart := &barchart.Config{Type: "bar", Scale: "linear"}
	_, err := Convert(ConvertInput{
		Input:  []byte("region,latency\nwest,nope\n"),
		Parser: "csv",
		Charts: []internalcharts.ChartConfig{chart},
	})
	s.ErrorContains(err, "no numeric columns")

	datasets, err := Merge([]shared.Dataset{
		{Name: "Sort", Tag: "v1", Settings: []internalcharts.ChartConfig{chart}, Data: []shared.DataPoint{{Name: "quick", YAxis: "12"}}},
		{Name: "Sort", Tag: "v2", Settings: []internalcharts.ChartConfig{chart}, Data: []shared.DataPoint{{Name: "quick", YAxis: "10"}}},
	}, shared.DimensionXAxis)
	s.Require().NoError(err)
	s.Len(datasets, 1)
	s.Len(datasets[0].Data, 2)

	html, err := GenerateUI(datasets, []string{"bar"})
	s.Require().NoError(err)
	s.True(strings.Contains(html, "VIZB_DATA"))
}

func (s *CoreSuite) TestConvertBranchErrorsAndAutoDetection() {
	chart := []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Scale: "linear"}}
	for _, tc := range []struct {
		name  string
		input ConvertInput
		want  string
	}{
		{"empty input", ConvertInput{Charts: chart}, "input is empty"},
		{"missing charts", ConvertInput{Input: []byte("x,y\na,1\n"), Parser: "csv"}, "at least one chart"},
		{"unknown parser", ConvertInput{Input: []byte("x,y\na,1\n"), Parser: "unknown", Charts: chart}, "unknown parser"},
		{"wrong JavaScript format", ConvertInput{Input: []byte("x,y\na,1\n"), Parser: "javascript", Charts: chart}, "does not match a supported JavaScript"},
		{"wrong Rust format", ConvertInput{Input: []byte("x,y\na,1\n"), Parser: "rust", Charts: chart}, "does not match a supported Rust"},
		{"bad filter", ConvertInput{Input: []byte("x,y\na,1\n"), Parser: "csv", Config: parser.Config{Filter: "["}, Charts: chart}, "invalid filter regex"},
		{"json path on csv", ConvertInput{Input: []byte("x,y\na,1\n"), Parser: "csv", Config: parser.Config{JSONPath: ".data"}, Charts: chart}, "only supported by the json parser"},
		{"missing json path", ConvertInput{Input: []byte(`[{"x":"a","y":1}]`), Parser: "json", Config: parser.Config{JSONPath: ".data"}, Charts: chart}, "cannot read key 'data'"},
		{"no results", ConvertInput{Input: []byte("header\n"), Parser: "csv", Charts: chart}, "no dataset found"},
		{"invalid swap", ConvertInput{Input: []byte("x,y\na,1\n"), Parser: "csv", Config: parser.Config{GroupPattern: "x", Group: []string{"x"}}, Charts: []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Swap: "x:z"}}}, "swap"},
	} {
		s.Run(tc.name, func() {
			_, err := Convert(tc.input)
			s.ErrorContains(err, tc.want)
		})
	}

	result, err := Convert(ConvertInput{
		Input:  []byte(`[{"region":"west","latency":12}]`),
		Parser: "auto",
		Config: parser.Config{GroupPattern: "x", Group: []string{"region"}},
		Charts: chart,
	})
	s.Require().NoError(err)
	s.Len(result.Dataset.Data, 1)
	result, err = Convert(ConvertInput{
		Input:  []byte("region,latency\nwest,12\n"),
		Parser: "  ",
		Charts: chart,
	})
	s.Require().NoError(err)
	s.Len(result.Dataset.Data, 1)
}

func (s *CoreSuite) TestConvertIdentifiesInapplicableOptions() {
	chart := []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Scale: "linear"}}
	for _, input := range []ConvertInput{
		{
			Input:  []byte("BenchmarkFoo-8 100 123 ns/op\n"),
			Parser: "auto",
			Config: parser.Config{
				SelectViews: []parser.SelectView{{Columns: []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}}},
			},
			Charts: chart,
		},
		{
			Input:  []byte("x,y\na,1\n"),
			Parser: "csv",
			Config: parser.Config{GroupPattern: "x", Group: []string{"x"}},
			Charts: []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Swap: "x:z"}},
		},
		{
			Input:  []byte("a,b,c\nx,y,1\n"),
			Parser: "auto",
			Config: parser.Config{GroupPattern: "x/y", Group: []string{"a", "b"}},
			Charts: chart,
		},
	} {
		_, err := Convert(input)
		var optionErr *OptionError
		s.Require().ErrorAs(err, &optionErr)
		s.NotEmpty(optionErr.Name)
		s.True(errors.Is(optionErr, optionErr.Err))
	}
}

func (s *CoreSuite) TestConvertBenchmarkFormats() {
	chart := []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Scale: "linear"}}
	cases := []struct {
		name   string
		parser string
		input  string
	}{
		{"Go explicit", "go", "BenchmarkFoo-8 100 123 ns/op\n"},
		{"Go auto", "auto", "BenchmarkFoo-8 100 123 ns/op\n"},
		{"Go JSON events auto", "auto", `{"Action":"output","Output":"BenchmarkFoo-8 100 123 ns/op\n"}` + "\n"},
		{"Vitest family", "javascript", " · foo 1234 0.1 0.2 0.3 0.4 0.5 0.6 0.7 ±1.5% 100\n"},
		{"Tinybench auto", "auto", "│ 0 │ 'foo' │ '123 ± 1%' │ '120 ± 2' │ '8000 ± 1%' │ '8100 ± 2' │ 100 │\n"},
		{"Criterion family", "rust", "foo time: [21.234 ns 21.456 ns 21.678 ns]\n"},
		{"Divan auto", "auto", "├─ foo 4.36 µs │ 9.68 µs │ 4.646 µs │ 4.733 µs │ 100 │ 100\n"},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			result, err := Convert(ConvertInput{
				Input:  []byte(tc.input),
				Parser: tc.parser,
				Config: parser.Config{GroupPattern: "y", TimeUnit: "ns"},
				Charts: chart,
			})
			s.Require().NoError(err)
			s.Require().Len(result.Dataset.Data, 1)
		})
	}
}

func (s *CoreSuite) TestConvertGoBenchmarkMetadataIsRequestLocal() {
	cases := []struct {
		input string
		meta  shared.Meta
	}{
		{
			input: "goos: linux\ngoarch: amd64\npkg: example.com/linux\ncpu: Linux CPU\nBenchmarkFoo-4 100 123 ns/op\n",
			meta:  shared.Meta{OS: "linux", Arch: "amd64", Pkg: "example.com/linux", CPU: &shared.CPUInfo{Name: "Linux CPU", Cores: 4}},
		},
		{
			input: "goos: darwin\ngoarch: arm64\npkg: example.com/darwin\ncpu: Darwin CPU\nBenchmarkBar-12 100 456 ns/op\n",
			meta:  shared.Meta{OS: "darwin", Arch: "arm64", Pkg: "example.com/darwin", CPU: &shared.CPUInfo{Name: "Darwin CPU", Cores: 12}},
		},
	}
	type outcome struct {
		index   int
		dataset *shared.Dataset
		err     error
	}

	const conversions = 64
	start := make(chan struct{})
	outcomes := make(chan outcome, conversions)
	var wg sync.WaitGroup
	for i := 0; i < conversions; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			tc := cases[index%len(cases)]
			result, err := Convert(ConvertInput{
				Input:  []byte(tc.input),
				Parser: "go",
				Config: parser.Config{GroupPattern: "y", TimeUnit: "ns"},
				Charts: []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Scale: "linear"}},
			})
			outcomes <- outcome{index: index, dataset: result.Dataset, err: err}
		}(i)
	}
	close(start)
	wg.Wait()
	close(outcomes)

	for outcome := range outcomes {
		s.Require().NoError(outcome.err)
		s.Require().NotNil(outcome.dataset)
		s.Equal(&cases[outcome.index%len(cases)].meta, outcome.dataset.Meta)
	}
}

func (s *CoreSuite) TestConvertKeepsExplicitSystemMetadata() {
	explicit := &shared.Meta{OS: "caller"}
	result, err := Convert(ConvertInput{
		Input:    []byte("goos: linux\nBenchmarkFoo-4 100 123 ns/op\n"),
		Parser:   "go",
		Config:   parser.Config{GroupPattern: "y", TimeUnit: "ns"},
		Metadata: Metadata{System: explicit},
		Charts:   []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Scale: "linear"}},
	})

	s.Require().NoError(err)
	s.Same(explicit, result.Dataset.Meta)
}

func (s *CoreSuite) TestConvertAutoJSONPathEnvelope() {
	result, err := Convert(ConvertInput{
		Input:  []byte(`{"data":[{"region":"west","latency":12},{"region":"east","latency":18}]}`),
		Parser: "auto",
		Config: parser.Config{
			JSONPath:     ".data",
			GroupPattern: "x",
			Group:        []string{"region"},
		},
		Charts: []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Scale: "linear"}},
	})

	s.Require().NoError(err)
	s.Require().Len(result.Dataset.Data, 2)
	s.Equal("west", result.Dataset.Data[0].XAxis)
}

func (s *CoreSuite) TestConvertReturnsParserLookupError() {
	chart := []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Scale: "linear"}}
	saved := parser.Parsers["csv"]
	delete(parser.Parsers, "csv")
	s.T().Cleanup(func() { parser.Parsers["csv"] = saved })

	_, err := Convert(ConvertInput{Input: []byte("x,y\na,1\n"), Parser: "csv", Charts: chart})
	s.ErrorContains(err, "unknown parser")
	parser.Parsers["csv"] = saved
}

func (s *CoreSuite) TestConvertReturnsChartRuleError() {
	saved := internalcharts.FlagsFor("bar")
	internalcharts.SetFlags("bar", []flags.Flag{{
		Name:    "scale",
		JSONKey: "scale",
		Rule: []flags.RuleFn{func(any) (flags.Outcome, string) {
			return flags.Fatal, "forced rule failure"
		}},
	}})
	s.T().Cleanup(func() { internalcharts.SetFlags("bar", saved) })

	_, err := Convert(ConvertInput{
		Input:  []byte("region,latency\nwest,12\n"),
		Parser: "csv",
		Config: parser.Config{GroupPattern: "x", Group: []string{"region"}},
		Charts: []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Scale: "linear"}},
	})
	s.ErrorContains(err, "forced rule failure")
}

func (s *CoreSuite) TestMergeAndUIBranchErrors() {
	_, err := Merge(nil, shared.DimensionName)
	s.ErrorContains(err, "at least one dataset")
	_, err = Merge([]shared.Dataset{{Name: "x"}}, shared.Dimension("bad"))
	s.ErrorContains(err, "invalid tag axis")
	merged, err := Merge([]shared.Dataset{{Name: "x"}}, shared.Dimension("name"))
	s.Require().NoError(err)
	s.Len(merged, 1)

	_, err = GenerateUI(nil, nil)
	s.ErrorContains(err, "at least one dataset")
	html, err := GenerateUI([]shared.Dataset{{Name: "x", Settings: []internalcharts.ChartConfig{&barchart.Config{Type: "bar"}}}}, nil)
	s.Require().NoError(err)
	s.Contains(html, "VIZB_CHARTS")
}

func (s *CoreSuite) TestAssembleModesAndThreeD() {
	v := true
	bar := &barchart.Config{Type: "bar"}
	line := &linechart.Config{Type: "line"}
	scatter := &scatterchart.Config{Type: "scatter"}
	charts := []internalcharts.ChartConfig{bar, line, scatter}
	dataset := Assemble(AssembleInput{
		Points:   []shared.DataPoint{{XAxis: "1", YAxis: "2", ZAxis: "3", Metric: "4"}},
		Parser:   "csv",
		Config:   parser.Config{Axes: []parser.ColumnSpec{{Source: "x"}, {Source: "y"}, {Source: "z"}}, MetricColumn: "score"},
		Metadata: Metadata{Timestamp: "2026-01-01T00:00:00Z", System: &shared.Meta{OS: "test"}},
		Charts:   charts,
	})
	s.Len(dataset.Axes, 4)
	s.Equal("score", dataset.Axes[3].Label)
	s.Equal("2026-01-01T00:00:00Z", dataset.Timestamp)
	s.True(dataset.PreserveRows)
	s.Equal(&v, bar.ThreeD)
	s.Equal(&v, line.ThreeDVisualMap)
	s.Equal(&v, scatter.ThreeD)

	selectDataset := Assemble(AssembleInput{
		Points:   []shared.DataPoint{{XAxis: "west", YAxis: "12"}},
		Config:   parser.Config{Mode: parser.ModeValue, SelectViews: []parser.SelectView{{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x", AxisType: "category"}, {Source: "latency", AxisKey: "y", AxisType: "value"}}}}},
		Metadata: Metadata{}, Charts: []internalcharts.ChartConfig{&barchart.Config{Type: "bar"}},
	})
	s.Equal("region × latency", selectDataset.Name)

	multiDataset := Assemble(AssembleInput{Points: []shared.DataPoint{{Name: "west"}}, Config: parser.Config{Mode: parser.ModeMultiStat, SelectViews: []parser.SelectView{{Columns: []parser.ColumnSpec{{Source: "region"}, {Source: "latency"}}}}}})
	s.NotEmpty(multiDataset.Axes)

	multiViewDataset := Assemble(AssembleInput{
		Points: []shared.DataPoint{{XAxis: "west", YAxis: "12"}},
		Config: parser.Config{Mode: parser.ModeValue, SelectViews: []parser.SelectView{
			{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x", AxisType: "category"}, {Source: "latency", AxisKey: "y", AxisType: "value"}}},
			{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x", AxisType: "category"}, {Source: "throughput", AxisKey: "y", AxisType: "value"}}},
		}},
	})
	s.Len(multiViewDataset.Axes, 2)

	categoryDataset := Assemble(AssembleInput{Points: []shared.DataPoint{{XAxis: "west", YAxis: "2", ZAxis: "3"}}, Config: parser.Config{Axes: []parser.ColumnSpec{{Source: "x"}, {Source: "y"}, {Source: "z"}}}, Charts: []internalcharts.ChartConfig{&barchart.Config{Type: "bar"}}})
	s.Len(categoryDataset.Axes, 3)
}

func (s *CoreSuite) TestAssembleFallbacksAndHelpers() {
	grouped := Assemble(AssembleInput{Config: parser.Config{GroupPattern: "name,x"}})
	s.Len(grouped.Axes, 2)

	mixed := Assemble(AssembleInput{Config: parser.Config{Axes: []parser.ColumnSpec{
		{Source: "region", AxisKey: "x", AxisType: "category"},
		{Source: "latency", AxisKey: "y", AxisType: "value"},
	}}})
	s.Equal("x", mixed.Axes[0].Key)

	value := Assemble(AssembleInput{Config: parser.Config{
		Axes:    []parser.ColumnSpec{{Source: "x", AxisKey: "x", AxisType: "value"}},
		ColAxis: "y",
	}})
	s.Len(value.Axes, 2)

	axes := appendMetricAxis([]shared.Axis{{Key: "metric"}}, parser.Config{MetricColumn: "score"}, nil)
	s.Len(axes, 1)
	axes = appendMetricAxis(nil, parser.Config{}, []shared.DataPoint{{Metric: "1"}})
	s.Equal("value", axes[0].Label)

	bar := &barchart.Config{Type: "bar"}
	autoEnableValueMode3D([]internalcharts.ChartConfig{bar}, []shared.Axis{{Key: "x", Type: "value"}, {Key: "y", Type: "value"}}, false)
	s.Nil(bar.ThreeD)
}

func (s *CoreSuite) TestGenerateUIReportsMarshalErrors() {
	_, err := GenerateUI([]shared.Dataset{{Data: []shared.DataPoint{{Stats: []shared.Stat{{Value: shared.F64(math.NaN())}}}}}}, nil)
	s.ErrorContains(err, "marshal datasets")
}

func TestCoreSuite(t *testing.T) { suite.Run(t, new(CoreSuite)) }
