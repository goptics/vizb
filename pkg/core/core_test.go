package core

import (
	"strings"
	"testing"

	internalcharts "github.com/goptics/vizb/internal/charts"
	barchart "github.com/goptics/vizb/internal/charts/bar"
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

func TestCoreSuite(t *testing.T) { suite.Run(t, new(CoreSuite)) }
