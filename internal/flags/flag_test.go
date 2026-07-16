package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlagSuite struct {
	suite.Suite
}

func (s *FlagSuite) TestFlagHelpers() {
	t := s.T()
	f := Flag{Name: "show-labels", Label: "labels", Key: "labels", JSONKey: "showLabels", ValidSet: []string{"a"}}
	assert.Equal(t, "labels", f.EffectiveLabel())
	assert.Equal(t, "labels", f.EffectiveKey())
	assert.True(t, f.IsChart())
	assert.True(t, f.IsSoft())

	plain := Flag{Name: "parser"}
	assert.Equal(t, "parser", plain.EffectiveLabel())
	assert.Equal(t, "parser", plain.EffectiveKey())
	assert.False(t, plain.IsChart())
	assert.False(t, plain.IsSoft())
}

func TestFlagSuite(t *testing.T) {
	suite.Run(t, new(FlagSuite))
}
