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
		Long:  "Generate an interactive scatter chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
	})
}
