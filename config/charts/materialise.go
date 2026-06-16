package charts

// Defaults holds the chart-agnostic global defaults the root command
// supplies to every chart's Materialise (subcommand mode fills Materialise
// from the subcommand's own flags and leaves this unused). It captures the
// "global default" step in the 4-step precedence: override > flags > defaults
// > internal default. The root command reads these from the -s/--sort and
// -l/--show-labels flags and translates them into each chart's Flags value.
type Defaults struct {
	Sort       string
	ShowLabels bool
}
