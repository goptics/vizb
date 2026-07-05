// Package bar registers the bar chart type: it plugs the typed Config factory
// into the charts registry, stores the flag descriptors, and advertises cobra
// metadata for the `vizb bar` subcommand.
package bar

import (
	"slices"

	"github.com/goptics/vizb/cmd/cli"
	"github.com/goptics/vizb/internal/charts"
	barchart "github.com/goptics/vizb/internal/charts/bar"
)

func init() {
	charts.Register(charts.Spec{Type: "bar", Factory: barchart.New})
	charts.SetFlags("bar", append(slices.Clone(charts.BaseChartFlags),
		charts.ScaleFlag, charts.ThreeDFlag, charts.ThreeDRotateFlag, charts.ThreeDVisualMapFlag,
		charts.HorizontalFlag,
	))
	cli.SetChartMeta(cli.ChartMeta{
		Type:  "bar",
		Use:   "bar [target]",
		Short: "Generate a bar chart from data",
		Long:  "Generate an interactive bar chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
	})
}
