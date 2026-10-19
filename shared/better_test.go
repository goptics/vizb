package shared

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

type BetterSuite struct {
	suite.Suite
}

func TestBetterSuite(t *testing.T) {
	suite.Run(t, new(BetterSuite))
}

func (s *BetterSuite) TestStatJSON() {
	s.Run("set", func() {
		b, err := json.Marshal(Stat{Type: "Latency", Better: BetterLow})
		s.Require().NoError(err)
		s.JSONEq(`{"type":"Latency","better":"low"}`, string(b))
	})
	s.Run("omitted when empty", func() {
		b, err := json.Marshal(Stat{Type: "Latency"})
		s.Require().NoError(err)
		s.JSONEq(`{"type":"Latency"}`, string(b))
	})
	s.Run("legacy json loads", func() {
		var st Stat
		s.Require().NoError(json.Unmarshal([]byte(`{"type":"Latency","value":1}`), &st))
		s.Empty(st.Better)
	})
}

func (s *BetterSuite) TestParseBetterFlag() {
	s.Run("empty", func() {
		rules, err := ParseBetterFlag("  ")
		s.Require().NoError(err)
		s.Nil(rules)
	})
	s.Run("splits on last colon and trims", func() {
		rules, err := ParseBetterFlag("rps:high, Memory Usage : LOW ,req/s:high,a:b:low")
		s.Require().NoError(err)
		s.Equal([]BetterRule{
			{Key: "rps", Dir: BetterHigh},
			{Key: "Memory Usage", Dir: BetterLow},
			{Key: "req/s", Dir: BetterHigh},
			{Key: "a:b", Dir: BetterLow},
		}, rules)
	})
	for _, bad := range []string{"rps:up", "rps", ":low", "rps:high,RPS:low", "rps:high,,x:low"} {
		s.Run("rejects "+bad, func() {
			_, err := ParseBetterFlag(bad)
			s.Error(err)
		})
	}
}

func (s *BetterSuite) TestApplyBetter() {
	data := func() []DataPoint {
		return []DataPoint{{Stats: []Stat{
			{Type: "Execution Time (ns/op)", Better: BetterLow},
			{Type: "Metric (widgets/op)"},
			{Type: "Allocations/op", Better: BetterLow},
			{Type: "Req/s"},
			{Type: "Iterations"},
		}}}
	}

	s.Run("matches base name, unit, full type and label case-insensitively", func() {
		d := data()
		warnings := ApplyBetter(d, []BetterRule{
			{Key: "execution time", Dir: BetterHigh},
			{Key: "widgets/op", Dir: BetterHigh},
			{Key: "allocations/op", Dir: BetterHigh},
			{Key: "req/s", Dir: BetterLow},
		})
		s.Empty(warnings)
		got := []string{}
		for _, st := range d[0].Stats {
			got = append(got, st.Better)
		}
		s.Equal([]string{BetterHigh, BetterHigh, BetterHigh, BetterLow, ""}, got)
	})

	s.Run("unit match is whole-unit", func() {
		d := data()
		warnings := ApplyBetter(d, []BetterRule{{Key: "op", Dir: BetterHigh}})
		s.Equal([]string{`better key "op" matched no metric; ignoring`}, warnings)
		s.Equal(BetterLow, d[0].Stats[0].Better)
	})

	s.Run("warns per unmatched key", func() {
		s.Equal([]string{`better key "nosuch" matched no metric; ignoring`}, ApplyBetter(data(), []BetterRule{{Key: "nosuch", Dir: BetterLow}}))
	})
}
