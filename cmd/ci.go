package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/goptics/vizb/pkg/ci"
	"github.com/goptics/vizb/pkg/style"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
	"github.com/spf13/cobra"
)

var ciIdentify string

var ciCmd = &cobra.Command{
	Use:   "ci [input]",
	Short: "Run vizb in CI mode to capture benchmark history",
	Long:  `Parses Go benchmark output and outputs a standard vizb benchmark JSON with the identify value on the missing dimension.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runCI,
}

func init() {
	rootCmd.AddCommand(ciCmd)
	ciCmd.Flags().StringVar(&shared.ActionState.SHA, "sha", "", "Git commit SHA")
	ciCmd.Flags().StringVar(&shared.ActionState.Tag, "tag", "", "Git tag (for tag-triggered runs)")
	ciCmd.Flags().StringVar(&shared.ActionState.Branch, "branch", "", "Git branch name")
	ciCmd.Flags().StringVar(&shared.ActionState.Append, "append", "", "Path to existing benchmarks.json to append to")
	ciCmd.Flags().StringVarP(&shared.ActionState.Output, "output", "o", "benchmarks.json", "Output file path")
	ciCmd.Flags().IntVar(&shared.ActionState.Keep, "keep", 0, "Max number of tags/commits to keep (0 = unlimited)")
	ciCmd.Flags().StringVar(&shared.ActionState.GroupPattern, "group-pattern", "n/y", "Pattern to parse benchmark names in CI mode (n/y, n/x, x/y, etc.). The missing dimension gets the identify value.")
	ciCmd.Flags().StringVar(&shared.ActionState.GroupRegex, "group-regex", "", "Regex to parse benchmark names in CI mode. Overrides group-pattern.")
	ciCmd.Flags().StringVar(&ciIdentify, "identify", "tag", "Field to use as the identify value (tag or sha)")
	registerBenchmarkFlags(ciCmd)
}

func runCI(cmd *cobra.Command, args []string) {
	utils.ApplyValidationRules(sharedFlagValidationRules)

	input := "stdin"
	if len(args) > 0 {
		input = args[0]
	}

	shared.FlagState.GroupPattern = shared.ActionState.GroupPattern
	shared.FlagState.GroupRegex = shared.ActionState.GroupRegex

	identifyValue := shared.ActionState.Tag
	if ciIdentify == "sha" {
		identifyValue = shared.ActionState.SHA
	}

	opts := ci.ActionOpts{
		Input:         input,
		IdentifyValue: identifyValue,
		Date:          time.Now(),
		AppendFile:    shared.ActionState.Append,
		KeepCount:     shared.ActionState.Keep,
		GroupPattern:  shared.ActionState.GroupPattern,
		GroupRegex:    shared.ActionState.GroupRegex,
	}

	bench, err := ci.RunAction(opts)
	if err != nil {
		shared.ExitWithError("ci failed", err)
	}

	data, err := json.Marshal(bench)
	if err != nil {
		shared.ExitWithError("marshal output", err)
	}

	f := shared.MustCreateFile(shared.ActionState.Output)
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		shared.ExitWithError("write output", err)
	}

	fmt.Println(style.Success.Render(fmt.Sprintf("Generated benchmark with %d data points (identify: %s) to %s", len(bench.Data), identifyValue, shared.ActionState.Output)))
}
