package heatmap

import (
	"slices"

	"github.com/goptics/vizb/cmd/cli"
	"github.com/goptics/vizb/internal/charts"
	heatmapchart "github.com/goptics/vizb/internal/charts/heatmap"
)

func init() {
	charts.Register(charts.Spec{Type: "heatmap", Factory: heatmapchart.New})
	charts.SetFlags("heatmap", slices.Clone(charts.BaseChartFlags))
	cli.SetChartMeta(cli.ChartMeta{
		Type:  "heatmap",
		Use:   "heatmap [target]",
		Short: "Generate a heatmap chart",
		Long:  "Generate an interactive heatmap chart (HTML or JSON) from CSV, JSON, or benchmark output.",
	})
}
