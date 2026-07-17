package core

import (
	"math"
	"strings"
	"testing"

	internalcharts "github.com/goptics/vizb/internal/charts"
	barchart "github.com/goptics/vizb/internal/charts/bar"
	linechart "github.com/goptics/vizb/internal/charts/line"
	scatterchart "github.com/goptics/vizb/internal/charts/scatter"
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
		{"unsupported parser", ConvertInput{Input: []byte("x,y\na,1\n"), Parser: "go", Charts: chart}, "does not support inline"},
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
	s.Equal("json", detectInlineParser([]byte(" [1]")))
	s.Equal("csv", detectInlineParser([]byte("x,y\n")))
}

func (s *CoreSuite) TestConvertReturnsReaderLookupError() {
	chart := []internalcharts.ChartConfig{&barchart.Config{Type: "bar", Scale: "linear"}}
	saved := parser.ReaderParsers["csv"]
	delete(parser.ReaderParsers, "csv")
	s.T().Cleanup(func() { parser.ReaderParsers["csv"] = saved })

	_, err := Convert(ConvertInput{Input: []byte("x,y\na,1\n"), Parser: "csv", Charts: chart})
	s.ErrorContains(err, "does not support in-memory input")
	parser.ReaderParsers["csv"] = saved
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
