package cmd

import (
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
)

func registerBenchmarkFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&shared.BenchSettings.Name, "name", "n", "Benchmarks", "Name of the benchmark")
	cmd.Flags().StringVarP(&shared.BenchSettings.Description, "description", "d", "", "Description of the benchmark")
	cmd.Flags().StringVarP(&shared.BenchSettings.MemUnit, "mem-unit", "m", "B", "Memory unit available: b, B, KB, MB, GB")
	cmd.Flags().StringVarP(&shared.BenchSettings.TimeUnit, "time-unit", "t", "ns", "Time unit available: ns, us, ms, s")
	cmd.Flags().StringVarP(&shared.BenchSettings.NumberUnit, "number-unit", "u", "", "Number unit available: K, M, B, T (default: as-is)")
	cmd.Flags().StringVarP(&shared.BenchSettings.Sort, "sort", "s", "", "Sort in asc or desc order (default: as-is)")
	cmd.Flags().StringSliceVarP(&shared.BenchSettings.Charts, "charts", "c", []string{"bar", "line", "pie"}, "Chart types to generate (bar, line, pie)")
	cmd.Flags().BoolVarP(&shared.BenchSettings.ShowLabels, "show-labels", "l", false, "Show labels on charts")
	cmd.Flags().StringVarP(&shared.BenchSettings.FilterRegex, "filter", "f", "", "Regex pattern to include only matching benchmark names")
	cmd.Flags().StringVarP(&shared.BenchSettings.Scale, "scale", "S", "linear", "Y-axis scale type (linear, log)")
}
