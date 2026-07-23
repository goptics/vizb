package scatter

import (
	"slices"

	"github.com/goptics/vizb/cmd/cli"
	"github.com/goptics/vizb/internal/charts"
	scatterchart "github.com/goptics/vizb/internal/charts/scatter"
)

func init() {
	charts.Register(charts.Spec{Type: "scatter", Factory: scatterchart.New})
	charts.SetFlags("scatter", append(slices.Clone(charts.BaseChartFlags),
		charts.ScaleFlag, charts.ThreeDFlag, charts.ThreeDRotateFlag, charts.ThreeDVisualMapFlag,
		charts.VisualMapFlag, charts.SymbolFlag, charts.SymbolSizeFlag,
	))
	cli.SetChartMeta(cli.ChartMeta{
		Type:  "scatter",
		Use:   "scatter [target]",
		Short: "Generate a scatter chart from data",
		Long: `Generate an interactive scatter chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.

Continuous coordinates (value axes) require solo --select, e.g.:
  vizb scatter data.csv --select x,y,z -o out.html

All-numeric files with no flags use auto col-axis x (column names as series),
not continuous coordinates. Use -A / --col-axis for series-on-axis without --select.`,
	})
}
