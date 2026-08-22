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

// --- brace-delimited object values ---

func (s *SpecParseSuite) TestParseObjectValueSemicolonFields() {
	spec, err := Parse("bar:bg={color=rgba(180, 180, 180, 0.2);borderColor=#000;borderWidth=0}", Options{AllowBareKeys: true})
	s.Require().NoError(err)
	s.Require().Len(spec.Props, 1)
	prop := spec.Props[0]
	s.Equal("bg", prop.Key)
	s.True(prop.HasValue)
	s.Empty(prop.Value)
	s.Require().NotNil(prop.Object)
	s.Equal([]Prop{
		{Key: "color", Value: "rgba(180, 180, 180, 0.2)", HasValue: true},
		{Key: "borderColor", Value: "#000", HasValue: true},
		{Key: "borderWidth", Value: "0", HasValue: true},
	}, prop.Object)
}

func (s *SpecParseSuite) TestParseObjectValueCommaStaysLiteral() {
	// No top-level ';': legacy comma mode must not split inside the braces.
	spec, err := Parse("bar:horizontal,bg={opacity=0.2;shadowColor=rgba(0, 0, 0, 0.5)}", Options{AllowBareKeys: true})
	s.Require().NoError(err)
	s.Require().Len(spec.Props, 2)
	s.Equal(Prop{Key: "horizontal", HasValue: false}, spec.Props[0])
	object := spec.Props[1].Object
	s.Require().NotNil(object)
	s.Len(object, 2)
	s.Equal("rgba(0, 0, 0, 0.5)", object[1].Value)
}

func (s *SpecParseSuite) TestParseObjectValueCommaListField() {
	spec, err := Parse("bar:bg={borderRadius=8,8,0,0}", Options{AllowBareKeys: true})
	s.Require().NoError(err)
	s.Require().Len(spec.Props, 1)
	s.Require().NotNil(spec.Props[0].Object)
	s.Require().Len(spec.Props[0].Object, 1)
	s.Equal("8,8,0,0", spec.Props[0].Object[0].Value)
}

func (s *SpecParseSuite) TestParseObjectValueEmptyBag() {
	spec, err := Parse("bar:bg={}", Options{AllowBareKeys: true})
	s.Require().NoError(err)
	s.Require().Len(spec.Props, 1)
	s.True(spec.Props[0].HasValue)
	s.NotNil(spec.Props[0].Object)
	s.Empty(spec.Props[0].Object)
}

func (s *SpecParseSuite) TestParseObjectValueWithSiblingAfterSemicolon() {
	spec, err := Parse("bar:sort=asc;bg={opacity=0.2};labels", Options{AllowBareKeys: true})
	s.Require().NoError(err)
	s.Require().Len(spec.Props, 3)
	s.Equal("sort", spec.Props[0].Key)
	s.Equal("opacity", spec.Props[1].Object[0].Key)
	s.Equal("labels", spec.Props[2].Key)
}

func (s *SpecParseSuite) TestParseSemicolonSkipsEmptySegments() {
	spec, err := Parse("bar:sort=asc;;labels", Options{AllowBareKeys: true})
	s.Require().NoError(err)
	s.Require().Len(spec.Props, 2)
	s.Equal("sort", spec.Props[0].Key)
	s.Equal("labels", spec.Props[1].Key)
}

func (s *SpecParseSuite) TestParseSemicolonRejectsEmptyFieldKey() {
	_, err := Parse("bar:sort=asc;=bad", Options{})
	s.Require().Error(err)
	s.Contains(err.Error(), "key")
}

func (s *SpecParseSuite) TestParseScalarPropsHaveNilObject() {
	spec, err := Parse("bar:swap=yxn,sort=asc", Options{})
	s.Require().NoError(err)
	s.Require().Len(spec.Props, 2)
	s.Nil(spec.Props[0].Object)
	s.Nil(spec.Props[1].Object)
}

func (s *SpecParseSuite) TestParseObjectValueErrors() {
	s.Run("unmatched open brace", func() {
		_, err := Parse("bar:bg={color=#000", Options{AllowBareKeys: true})
		s.Require().Error(err)
		s.Contains(err.Error(), "unmatched '{'")
	})
	s.Run("unmatched close brace", func() {
		_, err := Parse("bar:bg=color=#000}", Options{AllowBareKeys: true})
		s.Require().Error(err)
		s.Contains(err.Error(), "unmatched '}'")
	})
	s.Run("trailing text after object", func() {
		_, err := Parse("bar:bg={color=#000} trailing", Options{AllowBareKeys: true})
		s.Require().Error(err)
		s.Contains(err.Error(), "malformed object value")
	})
	s.Run("empty object field key", func() {
		_, err := Parse("bar:bg={=bad}", Options{AllowBareKeys: true})
		s.Require().Error(err)
		s.Contains(err.Error(), "key")
	})
	s.Run("nested object field value", func() {
		_, err := Parse("bar:bg={color={x=y}}", Options{AllowBareKeys: true})
		s.Require().Error(err)
		s.Contains(err.Error(), "nested braces")
	})
}

func (s *SpecParseSuite) TestParseObjectValueSkipsEmptyFields() {
	spec, err := Parse("bar:bg={color=#000;;opacity=0.5}", Options{})
	s.Require().NoError(err)
	s.Require().Len(spec.Props, 1)
	s.Equal([]Prop{
		{Key: "color", Value: "#000", HasValue: true},
		{Key: "opacity", Value: "0.5", HasValue: true},
	}, spec.Props[0].Object)
}

func (s *SpecParseSuite) TestParseBag() {
	props, err := ParseBag("color=rgba(180, 180, 180, 0.2);borderColor=#000;borderWidth=0")
	s.Require().NoError(err)
	s.Require().Len(props, 3)
	s.Equal(Prop{Key: "color", Value: "rgba(180, 180, 180, 0.2)", HasValue: true}, props[0])
	s.Equal(Prop{Key: "borderColor", Value: "#000", HasValue: true}, props[1])
	s.Equal(Prop{Key: "borderWidth", Value: "0", HasValue: true}, props[2])

	props, err = ParseBag("borderRadius=8,8,0,0")
	s.Require().NoError(err)
	s.Require().Len(props, 1)
	s.Equal("8,8,0,0", props[0].Value)
}

func (s *SpecParseSuite) TestParseBagRejectsBraces() {
	_, err := ParseBag("{color=#000}")
	s.Require().Error(err)
	s.Contains(err.Error(), "braces are not allowed")

	_, err = ParseBag("color=#000}")
	s.Require().Error(err)
}

func (s *SpecParseSuite) TestParseBagEmptyAndWhitespace() {
	props, err := ParseBag("")
	s.Require().NoError(err)
	s.Empty(props)

	props, err = ParseBag(" ; ")
	s.Require().NoError(err)
	s.Empty(props)
}

func (s *SpecParseSuite) TestParseBagRejectsInvalidFields() {
	_, err := ParseBag("=bad")
	s.Require().Error(err)
	s.Contains(err.Error(), "key")

	_, err = ParseBag("bare")
	s.Require().Error(err)
	s.Contains(err.Error(), "bare")
}
