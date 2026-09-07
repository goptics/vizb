package pie

import (
	"slices"

	"github.com/goptics/vizb/cmd/cli"
	"github.com/goptics/vizb/internal/charts"
	piechart "github.com/goptics/vizb/internal/charts/pie"
)

func init() {
	charts.Register(charts.Spec{Type: "pie", Factory: piechart.New})
	charts.SetFlags("pie", append(slices.Clone(charts.BaseChartFlags), charts.DonutFlag))
	cli.SetChartMeta(cli.ChartMeta{
		Type:  "pie",
		Use:   "pie [target]",
		Short: "Generate a pie chart",
		Long:  "Generate an interactive pie chart (HTML or JSON) from CSV, JSON, or benchmark output.",
	})
}
