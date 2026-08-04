package specparse

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SpecParseSuite struct {
	suite.Suite
}

func TestSpecParseSuite(t *testing.T) {
	suite.Run(t, new(SpecParseSuite))
}

func (s *SpecParseSuite) TestParseLegacyCommaTwoProps() {
	spec, err := Parse("bar:swap=yxn,sort=asc", Options{})
	s.Require().NoError(err)
	s.Equal("bar", spec.Prefix)
	s.Require().Len(spec.Props, 2)
	s.Equal(Prop{Key: "swap", Value: "yxn", HasValue: true}, spec.Props[0])
	s.Equal(Prop{Key: "sort", Value: "asc", HasValue: true}, spec.Props[1])
}

func (s *SpecParseSuite) TestParseSemicolonKeepsCommasInValues() {
	spec, err := Parse("bar:stat=center,spread;labels", Options{AllowBareKeys: true})
	s.Require().NoError(err)
	s.Equal("bar", spec.Prefix)
	s.Require().Len(spec.Props, 2)
	s.Equal(Prop{Key: "stat", Value: "center,spread", HasValue: true}, spec.Props[0])
	s.Equal(Prop{Key: "labels", Value: "", HasValue: false}, spec.Props[1])
}

func (s *SpecParseSuite) TestParseOceanColorsSemicolon() {
	spec, err := Parse("ocean:colors=#0ff,#00f;visualMapColors=#111,#222", Options{})
	s.Require().NoError(err)
	s.Equal("ocean", spec.Prefix)
	s.Require().Len(spec.Props, 2)
	s.Equal(Prop{Key: "colors", Value: "#0ff,#00f", HasValue: true}, spec.Props[0])
	s.Equal(Prop{Key: "visualMapColors", Value: "#111,#222", HasValue: true}, spec.Props[1])
}

func (s *SpecParseSuite) TestParseBareKeyAllowed() {
	spec, err := Parse("bar:labels", Options{AllowBareKeys: true})
	s.Require().NoError(err)
	s.Equal("bar", spec.Prefix)
	s.Require().Len(spec.Props, 1)
	s.Equal(Prop{Key: "labels", Value: "", HasValue: false}, spec.Props[0])
}

func (s *SpecParseSuite) TestParseBareKeyErrorsWithoutAllowBareKeys() {
	_, err := Parse("bar:labels", Options{})
	s.Require().Error(err)
	s.Contains(err.Error(), "bare")
}

func (s *SpecParseSuite) TestParseMultiValueWithKnownKeys() {
	opts := Options{
		MultiValueKeys: map[string]struct{}{"stat": {}},
		KnownKeys:      map[string]struct{}{"stat": {}, "labels": {}},
		AllowBareKeys:  true,
	}
	spec, err := Parse("bar:stat=center,spread,labels", opts)
	s.Require().NoError(err)
	s.Equal("bar", spec.Prefix)
	s.Require().Len(spec.Props, 2)
	s.Equal(Prop{Key: "stat", Value: "center,spread", HasValue: true}, spec.Props[0])
	s.Equal(Prop{Key: "labels", Value: "", HasValue: false}, spec.Props[1])
}

func (s *SpecParseSuite) TestParseMultiValueAlone() {
	opts := Options{
		MultiValueKeys: map[string]struct{}{"stat": {}},
	}
	spec, err := Parse("bar:stat=center,spread", opts)
	s.Require().NoError(err)
	s.Equal("bar", spec.Prefix)
	s.Require().Len(spec.Props, 1)
	s.Equal(Prop{Key: "stat", Value: "center,spread", HasValue: true}, spec.Props[0])
}

func (s *SpecParseSuite) TestParseMultiValueAloneContinuesThroughBareTokens() {
	// Without KnownKeys, multi-value consumption only stops at key= form,
	// so a trailing bare token is absorbed into the multi-value.
	opts := Options{
		MultiValueKeys: map[string]struct{}{"stat": {}},
	}
	spec, err := Parse("bar:stat=center,spread,labels", opts)
	s.Require().NoError(err)
	s.Require().Len(spec.Props, 1)
	s.Equal(Prop{Key: "stat", Value: "center,spread,labels", HasValue: true}, spec.Props[0])
}

func (s *SpecParseSuite) TestParseMultiValueStopsAtNextKeyEquals() {
	opts := Options{
		MultiValueKeys: map[string]struct{}{"stat": {}},
	}
	spec, err := Parse("bar:stat=center,spread,sort=asc", opts)
	s.Require().NoError(err)
	s.Require().Len(spec.Props, 2)
	s.Equal(Prop{Key: "stat", Value: "center,spread", HasValue: true}, spec.Props[0])
	s.Equal(Prop{Key: "sort", Value: "asc", HasValue: true}, spec.Props[1])
}

func (s *SpecParseSuite) TestParseErrors() {
	s.Run("empty", func() {
		_, err := Parse("", Options{})
		s.Require().Error(err)
	})
	s.Run("no colon", func() {
		_, err := Parse("bar", Options{})
		s.Require().Error(err)
		s.Contains(err.Error(), ":")
	})
	s.Run("empty prefix", func() {
		_, err := Parse(":swap=yxn", Options{})
		s.Require().Error(err)
		s.Contains(err.Error(), "prefix")
	})
	s.Run("whitespace prefix", func() {
		_, err := Parse("  :swap=yxn", Options{})
		s.Require().Error(err)
	})
	s.Run("require props empty rest", func() {
		_, err := Parse("bar:", Options{RequireProps: true})
		s.Require().Error(err)
	})
	s.Run("require props whitespace rest", func() {
		_, err := Parse("bar:   ", Options{RequireProps: true})
		s.Require().Error(err)
	})
	s.Run("empty key", func() {
		_, err := Parse("bar:=value", Options{})
		s.Require().Error(err)
		s.Contains(err.Error(), "key")
	})
	s.Run("empty key after separator", func() {
		_, err := Parse("bar:swap=yxn,=bad", Options{})
		s.Require().Error(err)
	})
}

func (s *SpecParseSuite) TestParseEmptyPropsAllowedWithoutRequire() {
	spec, err := Parse("bar:", Options{})
	s.Require().NoError(err)
	s.Equal("bar", spec.Prefix)
	s.Empty(spec.Props)

	spec, err = Parse("bar:   ", Options{})
	s.Require().NoError(err)
	s.Equal("bar", spec.Prefix)
	s.Empty(spec.Props)
}

func (s *SpecParseSuite) TestParsePreservesOrderAndKeyCasing() {
	spec, err := Parse("ocean:colors=#0ff;visualMapColors=#111;swap=yxn", Options{})
	s.Require().NoError(err)
	s.Require().Len(spec.Props, 3)
	s.Equal("colors", spec.Props[0].Key)
	s.Equal("visualMapColors", spec.Props[1].Key)
	s.Equal("swap", spec.Props[2].Key)
}

func (s *SpecParseSuite) TestParseTrimsKeyAndValue() {
	spec, err := Parse(" bar : swap = yxn , sort = asc ", Options{})
	s.Require().NoError(err)
	s.Equal("bar", spec.Prefix)
	s.Require().Len(spec.Props, 2)
	s.Equal(Prop{Key: "swap", Value: "yxn", HasValue: true}, spec.Props[0])
	s.Equal(Prop{Key: "sort", Value: "asc", HasValue: true}, spec.Props[1])
}

func (s *SpecParseSuite) TestParseEmptyValueWithEquals() {
	spec, err := Parse("bar:labels=", Options{})
	s.Require().NoError(err)
	s.Require().Len(spec.Props, 1)
	s.Equal(Prop{Key: "labels", Value: "", HasValue: true}, spec.Props[0])
}

func (s *SpecParseSuite) TestParseMultiValueCaseSensitiveKeys() {
	opts := Options{
		MultiValueKeys: map[string]struct{}{"stat": {}},
		KnownKeys:      map[string]struct{}{"stat": {}, "Stat": {}, "labels": {}},
		AllowBareKeys:  true,
	}
	// "Stat" is not in MultiValueKeys (case-sensitive), so no multi-value join.
	spec, err := Parse("bar:Stat=center,spread,labels", opts)
	s.Require().NoError(err)
	// Without multi-value, each comma segment is a prop start attempt.
	// "spread" is bare and not allowed... actually AllowBareKeys is true.
	// Stat=center, then bare spread, then bare labels.
	s.Require().Len(spec.Props, 3)
	s.Equal(Prop{Key: "Stat", Value: "center", HasValue: true}, spec.Props[0])
	s.Equal(Prop{Key: "spread", Value: "", HasValue: false}, spec.Props[1])
	s.Equal(Prop{Key: "labels", Value: "", HasValue: false}, spec.Props[2])
}

func (s *SpecParseSuite) TestParseKnownKeysStopsOnlyMatchingKeys() {
	opts := Options{
		MultiValueKeys: map[string]struct{}{"stat": {}},
		KnownKeys:      map[string]struct{}{"stat": {}, "labels": {}},
		AllowBareKeys:  true,
	}
	// "other" is not known, so absorbed into multi-value even as bare token.
	spec, err := Parse("bar:stat=center,other,labels", opts)
	s.Require().NoError(err)
	s.Require().Len(spec.Props, 2)
	s.Equal(Prop{Key: "stat", Value: "center,other", HasValue: true}, spec.Props[0])
	s.Equal(Prop{Key: "labels", Value: "", HasValue: false}, spec.Props[1])
}

func (s *SpecParseSuite) TestSplitList() {
	s.Equal([]string{"a", "b", "c"}, SplitList("a,b,c"))
	s.Equal([]string{"a", "b", "c"}, SplitList(" a , b , c "))
	s.Equal([]string{"center", "spread"}, SplitList("center,spread"))
	s.Empty(SplitList(""))
	s.Empty(SplitList("   "))
	s.Empty(SplitList(",,,"))
	s.Equal([]string{"a", "c"}, SplitList("a,,c"))
	s.Equal([]string{"#0ff", "#00f"}, SplitList("#0ff,#00f"))
}
