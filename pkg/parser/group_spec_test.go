package parser

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GroupSpecSuite struct {
	suite.Suite
}

func (s *GroupSpecSuite) TestParseGroupSpecFlatSlice() {
	spec, err := parseGroupSpec([]string{"name", "category", "region"}, "", []string{",", ","})
	s.Require().NoError(err)
	s.Equal([]string{"name", "category", "region"}, spec.Columns)
	s.False(spec.Structured)
}

func (s *GroupSpecSuite) TestParseGroupSpecSpaceSingleValue() {
	spec, err := parseGroupSpec([]string{"name category region"}, "x n y", []string{" ", " "})
	s.Require().NoError(err)
	s.Equal([]string{"name", "category", "region"}, spec.Columns)
	s.Equal([]string{" ", " "}, spec.Separators)
}

func (s *GroupSpecSuite) TestParseGroupSpecStructured() {
	spec, err := parseGroupSpec([]string{"name", "category/region"}, "", nil)
	s.Require().NoError(err)
	s.True(spec.Structured)
	s.Equal([]string{"name", "category", "region"}, spec.Columns)
	s.Equal([]string{",", "/"}, spec.Separators)
}

func (s *GroupSpecSuite) TestResolveGroupConfigSpacePattern() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"name category region"},
		GroupPattern: "x n y",
	})
	s.Require().NoError(err)
	s.Equal([]string{"name", "category", "region"}, cfg.GroupColumns)
	s.Equal([]string{" ", " "}, cfg.LabelSeparators)
}

func (s *GroupSpecSuite) TestResolveGroupConfigCommaPatternFlat() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"region", "product"},
		GroupPattern: "y,x",
	})
	s.Require().NoError(err)
	s.Equal([]string{"region", "product"}, cfg.GroupColumns)
	s.Equal([]string{","}, cfg.LabelSeparators)
}

func (s *GroupSpecSuite) TestGroupPatternSeparatorMismatch() {
	cases := []struct {
		name    string
		group   []string
		pattern string
		want    string
	}{
		{
			name:    "comma_group_slash_pattern",
			group:   []string{"product", "category", "region"},
			pattern: "x/y/z",
			want:    `--group "product,category,region" and --group-pattern "x/y/z" separators do not match (expected ", ,", got "/ /")`,
		},
		{
			name:    "slash_group_mixed_pattern",
			group:   []string{"product/category/region"},
			pattern: "x,y/z",
			want:    `--group "product/category/region" and --group-pattern "x,y/z" separators do not match (expected "/ /", got ", /")`,
		},
		{
			name:    "hash_group_mixed_pattern",
			group:   []string{"product#category#region"},
			pattern: "x#y,z",
			want:    `--group "product#category#region" and --group-pattern "x#y,z" separators do not match (expected "# #", got "# ,")`,
		},
		{
			name:    "structured_multi_arg",
			group:   []string{"name", "category/region"},
			pattern: "x/y/z",
			want:    `--group "name,category/region" and --group-pattern "x/y/z" separators do not match (expected ", /", got "/ /")`,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			cfg, err := ResolveGroupConfig(Config{
				Group:        tc.group,
				GroupPattern: tc.pattern,
			})
			if err == nil {
				err = ValidateTabularGroupAlignment(cfg)
			}
			s.Require().Error(err)
			s.Equal(tc.want, err.Error())
		})
	}
}

func (s *GroupSpecSuite) TestValidateTabularGroupAcceptsCommaPattern() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"region", "product", "month"},
		GroupPattern: "x,y,z",
	})
	s.Require().NoError(err)
	s.Require().NoError(ValidateTabularGroupAlignment(cfg))
}

func (s *GroupSpecSuite) TestJoinLabelPartsSpaces() {
	got := JoinLabelParts([]string{"alpha", "beta", "gamma"}, []string{" ", " "})
	s.Equal("alpha beta gamma", got)
}

func (s *GroupSpecSuite) TestParseBenchmarkNameSpacePattern() {
	got, err := ParseBenchmarkNameToGroups("alpha beta gamma", "x n y")
	s.Require().NoError(err)
	s.Equal("alpha", got["xAxis"])
	s.Equal("beta", got["name"])
	s.Equal("gamma", got["yAxis"])
}

func TestGroupSpecSuite(t *testing.T) {
	suite.Run(t, new(GroupSpecSuite))
}
