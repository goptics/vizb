package shared

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SafeSuite struct {
	suite.Suite
}

func (s *SafeSuite) TestWithSafeNoPanic() {
	executed := false
	err := WithSafe("test function", func() {
		executed = true
	})
	s.True(executed)
	s.NoError(err)
}

func (s *SafeSuite) TestWithSafePanicRecovery() {
	err := WithSafe("test panic", func() {
		panic("test panic message")
	})
	s.Require().Error(err)
	s.EqualError(err, "panic recovered inside test panic: test panic message")
}

func (s *SafeSuite) TestWithSafePanicDifferentTypes() {
	err := WithSafe("string panic", func() {
		panic("string error")
	})
	s.Error(err)

	err = WithSafe("int panic", func() {
		panic(42)
	})
	s.Error(err)
}

func TestSafeSuite(t *testing.T) {
	suite.Run(t, new(SafeSuite))
}
