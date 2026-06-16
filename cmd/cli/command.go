// Package cli holds the shared command machinery used by the root command and
// every chart subcommand: a self-registration registry, composable flag option
// groups (replacing the former global shared.FlagState), and the one linear
// pipeline that turns input into a chart HTML/JSON file.
package cli

import "github.com/spf13/cobra"

// CommandFactory builds a fresh *cobra.Command. Each chart subpackage registers
// one via init(), mirroring how pkg/parser parsers self-register.
type CommandFactory func() *cobra.Command

var registry []CommandFactory

// Register adds a chart command factory to the registry. Call from a chart
// subpackage's init(); the root command turns the registry into subcommands.
func Register(f CommandFactory) {
	registry = append(registry, f)
}

// Commands materialises every registered factory into a cobra command. The
// registry is chart-shape-agnostic: a future non-linear chart (graph, tree)
// registers the same way with its own RunE/pipeline.
func Commands() []*cobra.Command {
	cmds := make([]*cobra.Command, 0, len(registry))
	for _, f := range registry {
		cmds = append(cmds, f())
	}
	return cmds
}
