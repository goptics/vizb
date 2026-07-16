package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StatFlagSuite struct {
	suite.Suite
}

func (s *StatFlagSuite) TestStatValueSetAndString() {
	t := s.T()
	var math []string
	sv := &statValue{value: &math}

	require.NoError(t, sv.Set(statFlagAll))
	assert.Equal(t, []string{}, math)
	assert.Equal(t, "", sv.String())
	assert.Equal(t, "string", sv.Type())

	require.NoError(t, sv.Set("counts,center"))
	assert.Equal(t, []string{"counts", "center"}, math)
	assert.Equal(t, "counts,center", sv.String())
}

func (s *StatFlagSuite) TestLooksLikeStatValue() {
	t := s.T()
	assert.False(t, looksLikeStatValue(""))
	assert.False(t, looksLikeStatValue("-h"))
	assert.True(t, looksLikeStatValue("all"))
	assert.True(t, looksLikeStatValue("counts,center"))
	assert.False(t, looksLikeStatValue("counts,bogus"))
}

func (s *StatFlagSuite) TestRewriteStatArg() {
	t := s.T()
	assert.Equal(t, []string{"--stat=counts"}, RewriteStatArg([]string{"--stat", "counts"}))
	assert.Equal(t, []string{"--stat", "-o"}, RewriteStatArg([]string{"--stat", "-o"}))

	dir := t.TempDir()
	file := filepath.Join(dir, "bench.json")
	require.NoError(t, os.WriteFile(file, []byte("{}"), 0644))
	assert.Equal(t, []string{"--stat", file}, RewriteStatArg([]string{"--stat", file}))
}

func TestStatFlagSuite(t *testing.T) {
	suite.Run(t, new(StatFlagSuite))
}
