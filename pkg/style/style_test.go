package style

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ThemeSuite struct {
	suite.Suite
}

func (s *ThemeSuite) TestValidateAndNormalizeTheme() {
	for _, value := range []string{"default", "Westeros", "#F00,#00ff00"} {
		s.NoError(ValidateTheme(NormalizeTheme(value)))
	}
	s.Equal("westeros", NormalizeTheme(" Westeros "))
	s.Equal("#F00,#00ff00", NormalizeTheme(" #F00, #00ff00 "))
}

func (s *ThemeSuite) TestRejectsInvalidTheme() {
	for _, value := range []string{"", "unknown", "#F00", "#F00,not-a-color"} {
		s.Error(ValidateTheme(NormalizeTheme(value)))
	}
}

func TestThemeSuite(t *testing.T) {
	suite.Run(t, new(ThemeSuite))
}
