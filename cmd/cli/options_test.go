package cli

import (
	"testing"

	"github.com/goptics/vizb/testutil"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/suite"
)

// OptionsSuite covers option validation, parser-config mapping, and chart
// selection assembly.
type OptionsSuite struct {
	suite.Suite
}

func (s *OptionsSuite) TestCommonValidateNormalisesUnits() {
	o := &CommonOptions{MemUnit: "kb", TimeUnit: "ns", NumberUnit: "m", GroupPattern: "x", Parser: "auto"}
	o.Validate()
	s.Equal("KB", o.MemUnit)
	s.Equal("M", o.NumberUnit)
}

func (s *OptionsSuite) TestCommonValidateWarnsAndDefaultsInvalid() {
	o := &CommonOptions{MemUnit: "invalid", TimeUnit: "ns", GroupPattern: "x", Parser: "auto"}
	out := testutil.CaptureStderr(func() { o.Validate() })
	s.Equal("B", o.MemUnit)
	s.Contains(out, "Invalid memory unit")
}

func (s *OptionsSuite) TestCommonValidateRejectsUnknownParser() {
	o := &CommonOptions{TimeUnit: "ns", MemUnit: "B", GroupPattern: "x", Parser: "nope"}
	testutil.CaptureStderr(func() { o.Validate() })
	s.Equal("auto", o.Parser)
}

func (s *OptionsSuite) TestLinearValidateNormalisesSort() {
	o := &LinearOptions{Sort: "ASC"}
	o.CommonOptions = CommonOptions{TimeUnit: "ns", MemUnit: "B", GroupPattern: "x", Parser: "auto"}
	o.Validate()
	s.Equal("asc", o.Sort)
}

func (s *OptionsSuite) TestParseConfigMapsCols() {
	o := &CommonOptions{
		GroupPattern: "x",
		Cols:         "price{Unit price},count",
	}
	cfg := o.ParseConfig()
	s.Require().Len(cfg.Cols, 2)
	s.Equal("price", cfg.Cols[0].Source)
	s.Equal("Unit price", cfg.Cols[0].Label)
	s.Equal("count", cfg.Cols[1].Source)
}

func (s *OptionsSuite) TestParseConfigRejectsColsGroupOverlap() {
	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	o := &CommonOptions{
		GroupPattern: "x",
		Group:        []string{"date"},
		Cols:         "price,date",
	}
	s.Panics(func() { o.ParseConfig() })
	s.True(*exitCalled)
}

func (s *OptionsSuite) TestParseConfigMapsFields() {
	o := &CommonOptions{
		GroupPattern: "n/x", GroupRegex: "re", Group: []string{"a", "b"},
		Filter: "keep", MemUnit: "KB", TimeUnit: "us", NumberUnit: "M",
	}
	cfg := o.ParseConfig()
	s.Equal("n/x", cfg.GroupPattern)
	s.Equal("re", cfg.GroupRegex)
	s.Equal([]string{"a", "b"}, cfg.Group)
	s.Equal("keep", cfg.Filter)
	s.Equal("KB", cfg.MemUnit)
	s.Equal("us", cfg.TimeUnit)
	s.Equal("M", cfg.NumberUnit)
}

func (s *OptionsSuite) TestLinearOptionsBind() {
	var common CommonOptions
	commonFS := pflag.NewFlagSet("common", pflag.ContinueOnError)
	common.Bind(commonFS)
	s.NotNil(commonFS.Lookup("name"))
	s.NotNil(commonFS.Lookup("parser"))
	s.NotNil(commonFS.Lookup("group-pattern"))

	var linear LinearOptions
	linearFS := pflag.NewFlagSet("linear", pflag.ContinueOnError)
	linear.Bind(linearFS)
	s.NotNil(linearFS.Lookup("sort"))
	s.NotNil(linearFS.Lookup("show-labels"))
	s.NotNil(linearFS.Lookup("name"))

	var chart ChartOptions
	chartFS := pflag.NewFlagSet("chart", pflag.ContinueOnError)
	chart.Bind(chartFS)
	s.NotNil(chartFS.Lookup("swap"))
	s.NotNil(chartFS.Lookup("sort"))
}

func (s *OptionsSuite) TestValidateParserInvalid() {
	err := validateParser("nope")
	s.Error(err)
	s.Contains(err.Error(), "unknown parser")
}

func (s *OptionsSuite) TestValidateScale() {
	s.Run("log is accepted", func() {
		scale := "LOG"
		ValidateScale(&scale)
		s.Equal("log", scale)
	})
	s.Run("invalid falls back to linear", func() {
		scale := "bogus"
		testutil.CaptureStderr(func() { ValidateScale(&scale) })
		s.Equal("linear", scale)
	})
}

func TestOptionsSuite(t *testing.T) {
	suite.Run(t, new(OptionsSuite))
}
