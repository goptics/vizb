package cli

import (
	"slices"
	"testing"

	internal_charts "github.com/goptics/vizb/internal/charts"
	"github.com/goptics/vizb/internal/flags"
	"github.com/goptics/vizb/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

// FlagBagSuite covers the unified flag bag: binding, warn-and-default / fatal
// validation, parser-config mapping, and chart-seed assembly.
type FlagBagSuite struct {
	suite.Suite
}

// newCmdBag builds a bag over fl, bound to a fresh cobra command's flag set.
func (s *FlagBagSuite) newCmdBag(fl []flags.Flag) (*cobra.Command, *FlagBag) {
	bag := NewFlagBag(fl)
	cmd := &cobra.Command{Use: "t"}
	bag.Bind(cmd.Flags())
	return cmd, bag
}

func (s *FlagBagSuite) TestValidateNormalisesUnits() {
	cmd, bag := s.newCmdBag(slices.Clone(DataFlags))
	s.Require().NoError(cmd.Flags().Set("mem-unit", "kb"))
	s.Require().NoError(cmd.Flags().Set("number-unit", "m"))
	bag.Validate(cmd)
	s.Equal("KB", bag.String("mem-unit"))
	s.Equal("M", bag.String("number-unit"))
}

func (s *FlagBagSuite) TestValidateWarnsAndDefaultsInvalid() {
	cmd, bag := s.newCmdBag(slices.Clone(DataFlags))
	s.Require().NoError(cmd.Flags().Set("mem-unit", "invalid"))
	out := testutil.CaptureStderr(func() { bag.Validate(cmd) })
	s.Equal("B", bag.String("mem-unit"))
	s.Contains(out, "Invalid memory unit")
}

func (s *FlagBagSuite) TestValidateRejectsUnknownParser() {
	cmd, bag := s.newCmdBag(slices.Clone(DataFlags))
	s.Require().NoError(cmd.Flags().Set("parser", "nope"))
	testutil.CaptureStderr(func() { bag.Validate(cmd) })
	s.Equal("auto", bag.String("parser"))
}

func (s *FlagBagSuite) TestValidateNormalisesSort() {
	cmd, bag := s.newCmdBag(append(slices.Clone(DataFlags), internal_charts.SortFlag))
	s.Require().NoError(cmd.Flags().Set("sort", "ASC"))
	bag.Validate(cmd)
	s.Equal("asc", bag.String("sort"))
}

func (s *FlagBagSuite) TestScaleWarnDefault() {
	s.Run("LOG is normalised", func() {
		cmd, bag := s.newCmdBag(append(slices.Clone(DataFlags), internal_charts.ScaleFlag))
		s.Require().NoError(cmd.Flags().Set("scale", "LOG"))
		bag.Validate(cmd)
		s.Equal("log", bag.String("scale"))
	})
	s.Run("invalid falls back to linear", func() {
		cmd, bag := s.newCmdBag(append(slices.Clone(DataFlags), internal_charts.ScaleFlag))
		s.Require().NoError(cmd.Flags().Set("scale", "bogus"))
		testutil.CaptureStderr(func() { bag.Validate(cmd) })
		s.Equal("linear", bag.String("scale"))
	})
}

func (s *FlagBagSuite) TestParseConfigMapsSelectGrouped() {
	cmd, bag := s.newCmdBag(slices.Clone(DataFlags))
	s.Require().NoError(cmd.Flags().Set("group", "date"))
	s.Require().NoError(cmd.Flags().Set("select", "price{Unit price},count"))
	cfg := bag.ParseConfig()
	s.Require().Len(cfg.Select, 2)
	s.Empty(cfg.SelectViews)
	s.Equal("price", cfg.Select[0].Source)
	s.Equal("Unit price", cfg.Select[0].Label)
	s.Equal("count", cfg.Select[1].Source)
}

func (s *FlagBagSuite) TestParseConfigMapsSelectSoloAxisMode() {
	cmd, bag := s.newCmdBag(slices.Clone(DataFlags))
	s.Require().NoError(cmd.Flags().Set("select", "region,latency"))
	cfg := bag.ParseConfig()
	s.Empty(cfg.Select)
	s.Require().Len(cfg.SelectViews, 1)
	s.Require().Len(cfg.SelectViews[0].Columns, 2)
	s.Equal("region", cfg.SelectViews[0].Columns[0].Source)
	s.Equal("x", cfg.SelectViews[0].Columns[0].AxisKey)
	s.Equal("latency", cfg.SelectViews[0].Columns[1].Source)
	s.Equal("y", cfg.SelectViews[0].Columns[1].AxisKey)
}

func (s *FlagBagSuite) TestParseConfigMapsSelectParenTypeLabel() {
	cmd, bag := s.newCmdBag(slices.Clone(DataFlags))
	s.Require().NoError(cmd.Flags().Set("select", "region{Region},latency (Latency by Region)"))
	cfg := bag.ParseConfig()
	s.Require().Len(cfg.SelectViews, 1)
	s.Equal("Latency by Region", cfg.SelectViews[0].TypeLabel)
	s.Equal("Region", cfg.SelectViews[0].Columns[0].Label)
	s.Equal("latency", cfg.SelectViews[0].Columns[1].Source)
}

func (s *FlagBagSuite) TestParseConfigMapsRepeatableSelectSolo() {
	cmd, bag := s.newCmdBag(slices.Clone(DataFlags))
	s.Require().NoError(cmd.Flags().Set("select", "region,latency"))
	s.Require().NoError(cmd.Flags().Set("select", "region,sales"))
	cfg := bag.ParseConfig()
	s.Require().Len(cfg.SelectViews, 2)
	s.Len(cfg.SelectViews[0].Columns, 2)
	s.Len(cfg.SelectViews[1].Columns, 2)
}

func (s *FlagBagSuite) TestParseConfigMergesRepeatableSelectGrouped() {
	cmd, bag := s.newCmdBag(slices.Clone(DataFlags))
	s.Require().NoError(cmd.Flags().Set("group", "date"))
	s.Require().NoError(cmd.Flags().Set("select", "price"))
	s.Require().NoError(cmd.Flags().Set("select", "count"))
	cfg := bag.ParseConfig()
	s.Require().Len(cfg.Select, 2)
	s.Equal("price", cfg.Select[0].Source)
	s.Equal("count", cfg.Select[1].Source)
}

func (s *FlagBagSuite) TestParseConfigRejectsSoloSelectArity() {
	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	cmd, bag := s.newCmdBag(slices.Clone(DataFlags))
	s.Require().NoError(cmd.Flags().Set("select", "region"))
	s.Panics(func() { bag.ParseConfig() })
	s.True(*exitCalled)
}

func (s *FlagBagSuite) TestParseConfigRejectsSelectGroupOverlap() {
	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	cmd, bag := s.newCmdBag(slices.Clone(DataFlags))
	s.Require().NoError(cmd.Flags().Set("group", "date"))
	s.Require().NoError(cmd.Flags().Set("select", "price,date"))
	s.Panics(func() { bag.ParseConfig() })
	s.True(*exitCalled)
}

func (s *FlagBagSuite) TestParseConfigRejectsInvalidSelect() {
	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	cmd, bag := s.newCmdBag(slices.Clone(DataFlags))
	s.Require().NoError(cmd.Flags().Set("select", "price{unclosed"))
	s.Panics(func() { bag.ParseConfig() })
	s.True(*exitCalled)
}

func (s *FlagBagSuite) TestParseConfigMapsFields() {
	cmd, bag := s.newCmdBag(slices.Clone(DataFlags))
	s.Require().NoError(cmd.Flags().Set("group-pattern", "n/x"))
	s.Require().NoError(cmd.Flags().Set("group-regex", "re"))
	s.Require().NoError(cmd.Flags().Set("group", "a,b"))
	s.Require().NoError(cmd.Flags().Set("filter", "keep"))
	s.Require().NoError(cmd.Flags().Set("mem-unit", "KB"))
	s.Require().NoError(cmd.Flags().Set("time-unit", "us"))
	s.Require().NoError(cmd.Flags().Set("number-unit", "M"))
	cfg := bag.ParseConfig()
	s.Equal("n/x", cfg.GroupPattern)
	s.Equal("re", cfg.GroupRegex)
	s.Equal([]string{"a", "b"}, cfg.Group)
	s.Equal("keep", cfg.Filter)
	s.Equal("KB", cfg.MemUnit)
	s.Equal("us", cfg.TimeUnit)
	s.Equal("M", cfg.NumberUnit)
}

func (s *FlagBagSuite) TestBindRegistersFlags() {
	fl := append(slices.Clone(DataFlags), internal_charts.BaseChartFlags...)
	fl = append(fl, internal_charts.ScaleFlag)
	cmd, _ := s.newCmdBag(fl)
	for _, name := range []string{"name", "parser", "group-pattern", "sort", "show-labels", "swap", "scale", "stat"} {
		s.NotNil(cmd.Flags().Lookup(name), "missing --%s", name)
	}
	s.Nil(cmd.Flags().Lookup("axes"))
}

func (s *FlagBagSuite) TestChartSeedTriStateStatAndScale() {
	fl := append(slices.Clone(DataFlags), internal_charts.BaseChartFlags...)
	fl = append(fl, internal_charts.ScaleFlag)

	s.Run("unset: stat omitted, scale defaulted", func() {
		cmd, bag := s.newCmdBag(fl)
		seed := bag.ChartSeed(cmd)
		s.NotContains(seed, "stat")
		s.Equal("linear", seed["scale"])
	})
	s.Run("bare --stat enables all categories", func() {
		cmd, bag := s.newCmdBag(fl)
		s.Require().NoError(cmd.Flags().Set("stat", "all"))
		seed := bag.ChartSeed(cmd)
		s.Equal(map[string]any{"enabled": true, "math": []string{}}, seed["stat"])
	})
	s.Run("--sort encodes to map; --show-labels keys to showLabels", func() {
		cmd, bag := s.newCmdBag(fl)
		s.Require().NoError(cmd.Flags().Set("sort", "asc"))
		s.Require().NoError(cmd.Flags().Set("show-labels", "true"))
		seed := bag.ChartSeed(cmd)
		s.Equal(map[string]any{"enabled": true, "order": "asc"}, seed["sort"])
		s.Equal(true, seed["showLabels"])
	})
}

func (s *FlagBagSuite) TestValidateParserInvalid() {
	err := validateParser("nope")
	s.Error(err)
	s.Contains(err.Error(), "unknown parser")
}

func (s *FlagBagSuite) TestReadersAndRefs() {
	fl := append(slices.Clone(DataFlags), internal_charts.BaseChartFlags...)
	fl = append(fl, internal_charts.ScaleFlag, internal_charts.SymbolSizeFlag)
	cmd, bag := s.newCmdBag(fl)

	s.Require().NoError(cmd.Flags().Set("show-labels", "true"))
	s.Require().NoError(cmd.Flags().Set("symbol-size", "8"))
	s.Require().NoError(cmd.Flags().Set("group", "a,b"))
	s.True(bag.Bool("show-labels"))
	s.Equal(8.0, bag.Float("symbol-size"))
	s.Equal([]string{"a", "b"}, bag.StringSlice("group"))
	s.Require().NotNil(bag.StringSliceRef("group"))
}

func (s *FlagBagSuite) TestValidateRejectsInvalidSymbolSize() {
	fl := append(slices.Clone(DataFlags), internal_charts.SymbolSizeFlag)
	cmd, bag := s.newCmdBag(fl)
	s.Require().NoError(cmd.Flags().Set("symbol-size", "0"))

	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()
	s.Panics(func() { bag.Validate(cmd) })
	s.True(*exitCalled)
}

func TestFlagBagSuite(t *testing.T) {
	suite.Run(t, new(FlagBagSuite))
}
