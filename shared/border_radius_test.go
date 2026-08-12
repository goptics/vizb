package shared

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

type BorderRadiusSuite struct {
	suite.Suite
}

func (s *BorderRadiusSuite) TestMarshalAlwaysArray() {
	raw, err := json.Marshal(BorderRadius{8})
	s.Require().NoError(err)
	s.Equal("[8]", string(raw))

	raw, err = json.Marshal(BorderRadius{8, 8, 0, 0})
	s.Require().NoError(err)
	s.Equal("[8,8,0,0]", string(raw))
}

func (s *BorderRadiusSuite) TestUnmarshalArrays() {
	cases := []struct {
		in   string
		want BorderRadius
	}{
		{"[8]", BorderRadius{8}},
		{"[8,4]", BorderRadius{8, 4}},
		{"[1,2,3]", BorderRadius{1, 2, 3}},
		{"[8,8,0,0]", BorderRadius{8, 8, 0, 0}},
	}
	for _, tc := range cases {
		var r BorderRadius
		s.Require().NoError(json.Unmarshal([]byte(tc.in), &r), tc.in)
		s.Equal(tc.want, r, tc.in)
	}
}

func (s *BorderRadiusSuite) TestUnmarshalNull() {
	var r BorderRadius = BorderRadius{1}
	s.Require().NoError(json.Unmarshal([]byte("null"), &r))
	s.Nil(r)
}

func (s *BorderRadiusSuite) TestUnmarshalRejectsInvalid() {
	invalid := []string{
		`8`, `""`, `"8"`, `[]`, `[1,2,3,4,5]`, `[-1]`, `[8.5]`,
		`[1,"x"]`, `true`, `{}`, `[1.5,2]`,
	}
	for _, in := range invalid {
		var r BorderRadius
		s.Error(json.Unmarshal([]byte(in), &r), in)
	}
}

func (s *BorderRadiusSuite) TestPointerFieldOmitEmpty() {
	type wrap struct {
		BorderRadius *BorderRadius `json:"borderRadius,omitempty"`
	}
	raw, err := json.Marshal(wrap{})
	s.Require().NoError(err)
	s.Equal("{}", string(raw))

	r := BorderRadius{8}
	raw, err = json.Marshal(wrap{BorderRadius: &r})
	s.Require().NoError(err)
	s.JSONEq(`{"borderRadius":[8]}`, string(raw))
}

func TestBorderRadiusSuite(t *testing.T) {
	suite.Run(t, new(BorderRadiusSuite))
}
