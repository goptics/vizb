package pie

import (
	"slices"

	"github.com/goptics/vizb/cmd/cli"
	"github.com/goptics/vizb/config/charts"
	piechart "github.com/goptics/vizb/config/charts/pie"
)

func init() {
	charts.Register(charts.Spec{Type: "pie", Factory: piechart.New})
	charts.SetFlags("pie", slices.Clone(charts.BaseChartFlags))
	cli.SetChartMeta(cli.ChartMeta{
		Type:  "pie",
		Use:   "pie [target]",
		Short: "Generate a pie chart from data",
		Long:  "Generate an interactive pie chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
	})
}
