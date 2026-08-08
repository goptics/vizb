package charts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BorderRadiusValidationSuite struct {
	suite.Suite
}

func (s *BorderRadiusValidationSuite) TestValidateBorderRadiusValue() {
	tests := []struct {
		name      string
		value     string
		expectErr bool
	}{
		{"Valid positive integer", "8", false},
		{"Valid zero", "0", false},
		{"Valid large integer", "100", false},
		{"Invalid negative", "-5", true},
		{"Invalid non-integer", "abc", true},
		{"Invalid float", "8.5", true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := ValidateBorderRadiusValue(tt.value)
			if tt.expectErr {
				assert.Error(s.T(), err)
			} else {
				assert.NoError(s.T(), err)
			}
		})
	}
}

func TestBorderRadiusValidationSuite(t *testing.T) {
	suite.Run(t, new(BorderRadiusValidationSuite))
}