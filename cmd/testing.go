package cmd

import (
	"github.com/goptics/vizb/cmd/cli"
	"github.com/goptics/vizb/shared"
)

// ResetTestState restores root, ui, and merge flag globals to their defaults.
// Test-only entrypoint — production code must not call this.
func ResetTestState() {
	rootOpts = rootOptions{
		LinearOptions: cli.LinearOptions{
			CommonOptions: cli.CommonOptions{
				Parser:       "auto",
				GroupPattern: "x",
				MemUnit:      "B",
				TimeUnit:     "ns",
			},
		},
		Charts: append([]string(nil), defaultChartTypes...),
	}
	uiOpts = uiOptions{Charts: append([]string(nil), shared.DefaultChartTypes...)}
	mergeOpts = mergeOptions{TagAxis: "n"}
}
