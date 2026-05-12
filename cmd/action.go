package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/goptics/vizb/pkg/ci"
	"github.com/goptics/vizb/pkg/style"
	"github.com/goptics/vizb/pkg/template"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
)

var actionCmd = &cobra.Command{
	Use:   "action [input]",
	Short: "Run vizb in CI mode to capture benchmark history",
	Long:  `Parses Go benchmark output and outputs a standard vizb benchmark JSON with tag on x-axis.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runAction,
}

func init() {
	rootCmd.AddCommand(actionCmd)
	actionCmd.Flags().StringVar(&shared.ActionState.SHA, "sha", "", "Git commit SHA")
	actionCmd.Flags().StringVar(&shared.ActionState.Tag, "tag", "", "Git tag (for tag-triggered runs)")
	actionCmd.Flags().StringVar(&shared.ActionState.Branch, "branch", "", "Git branch name")
	actionCmd.Flags().StringVar(&shared.ActionState.Merge, "merge", "", "Path to existing benchmarks.json to merge with")
	actionCmd.Flags().StringVarP(&shared.ActionState.Output, "output", "o", "benchmarks.json", "Output file path")
	actionCmd.Flags().BoolVar(&shared.ActionState.HTML, "html", false, "Also generate an index.html viewer")
	actionCmd.Flags().IntVar(&shared.ActionState.Keep, "keep", 0, "Max number of tags/commits to keep (0 = unlimited)")
}

func runAction(cmd *cobra.Command, args []string) {
	input := "stdin"
	if len(args) > 0 {
		input = args[0]
	}

	opts := ci.ActionOpts{
		Input:     input,
		Version:   shared.ActionState.SHA,
		Tag:       shared.ActionState.Tag,
		Branch:    shared.ActionState.Branch,
		Date:      time.Now(),
		MergeFile: shared.ActionState.Merge,
		Output:    shared.ActionState.Output,
		KeepCount: shared.ActionState.Keep,
	}

	bench, err := ci.RunAction(opts)
	if err != nil {
		shared.ExitWithError("action failed", err)
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

	fmt.Println(style.Success.Render(fmt.Sprintf("Generated benchmark with %d data points (tag: %s) to %s", len(bench.Data), shared.ActionState.Tag, shared.ActionState.Output)))

	if shared.ActionState.HTML {
		dataForHTML, err := json.Marshal([]shared.Benchmark{*bench})
		if err != nil {
			shared.ExitWithError("marshal html data", err)
		}
		htmlContent := template.GenerateHTMLBenchmarkUI(dataForHTML, template.VizbHTMLTemplate)
		htmlFile := shared.MustCreateFile("index.html")
		defer htmlFile.Close()
		if _, err := htmlFile.WriteString(htmlContent); err != nil {
			shared.ExitWithError("write html", err)
		}
		fmt.Println(style.Success.Render("Generated CI viewer index.html"))
	}
}
