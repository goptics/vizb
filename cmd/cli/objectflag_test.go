package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ObjectFlagSuite covers the optional-value object flag binding (the --stat
// pattern applied to KindObject flags like --bg) and its arg rewrite.
type ObjectFlagSuite struct {
	suite.Suite
}

func (s *ObjectFlagSuite) TestObjectValueSetAndString() {
	t := s.T()
	var raw string
	ov := &objectValue{value: &raw}

	require.NoError(t, ov.Set(objectFlagOn))
	assert.Equal(t, objectFlagOn, raw)
	assert.Equal(t, objectFlagOn, ov.String())
	assert.Equal(t, "string", ov.Type())

	require.NoError(t, ov.Set("color=#fff;opacity=0.5"))
	assert.Equal(t, "color=#fff;opacity=0.5", ov.String())

	nilValue := &objectValue{}
	assert.Equal(t, "", nilValue.String())
}

func (s *ObjectFlagSuite) TestLooksLikeObjectValue() {
	t := s.T()
	assert.False(t, looksLikeObjectValue(""))
	assert.False(t, looksLikeObjectValue("-h"))
	assert.False(t, looksLikeObjectValue("--chart"))
	assert.True(t, looksLikeObjectValue(objectFlagOn))
	assert.True(t, looksLikeObjectValue("{color=#000}"))
	assert.False(t, looksLikeObjectValue("{color=#000"))
	assert.True(t, looksLikeObjectValue("color=rgba(1,2,3,0.2)"))
	assert.True(t, looksLikeObjectValue("color=#fff;borderColor=#000"))
	assert.True(t, looksLikeObjectValue("borderRadius=8,8,0,0"))
	// Unknown keys still look like a bag so --bg decal=1 is rewritten and
	// rejected by validation instead of being treated as a filename.
	assert.True(t, looksLikeObjectValue("decal=1"))
	assert.True(t, looksLikeObjectValue("color=#fff;decal=1"))
	assert.False(t, looksLikeObjectValue("color"))
	assert.False(t, looksLikeObjectValue("color=#fff;bare"))
}

func (s *ObjectFlagSuite) TestRewriteObjectArgJoinsSpaceForm() {
	args := []string{"vizb", "bar", "--bg", "color=rgba(180, 180, 180, 0.2);borderColor=#000", "data.csv"}
	got := RewriteObjectArg(args)
	s.Equal([]string{"vizb", "bar", "--bg=color=rgba(180, 180, 180, 0.2);borderColor=#000", "data.csv"}, got)
}

func (s *ObjectFlagSuite) TestRewriteObjectArgJoinsUnknownKey() {
	// Space-form unknown keys must reach FlagBag validation, not become a path.
	args := []string{"vizb", "bar", "--bg", "decal=1", "data.csv"}
	got := RewriteObjectArg(args)
	s.Equal([]string{"vizb", "bar", "--bg=decal=1", "data.csv"}, got)
}

func (s *ObjectFlagSuite) TestRewriteObjectArgLeavesOtherForms() {
	s.Run("equals form untouched", func() {
		args := []string{"vizb", "bar", "--bg=opacity=0.2", "data.csv"}
		s.Equal(args, RewriteObjectArg(args))
	})
	s.Run("bare flag untouched", func() {
		args := []string{"vizb", "bar", "--bg", "data.csv"}
		s.Equal(args, RewriteObjectArg(args))
	})
	s.Run("next flag untouched", func() {
		args := []string{"vizb", "bar", "--bg", "--3d", "data.csv"}
		s.Equal(args, RewriteObjectArg(args))
	})
	s.Run("unknown flag untouched", func() {
		args := []string{"vizb", "bar", "--nope", "color=#fff"}
		s.Equal(args, RewriteObjectArg(args))
	})
	s.Run("positional file not consumed", func() {
		file := s.T().TempDir() + "/data.csv"
		require.NoError(s.T(), os.WriteFile(file, []byte("a\n"), 0o600))
		args := []string{"vizb", "bar", "--bg", file}
		s.Equal(args, RewriteObjectArg(args))
	})
}

func TestObjectFlagSuite(t *testing.T) {
	suite.Run(t, new(ObjectFlagSuite))
}
