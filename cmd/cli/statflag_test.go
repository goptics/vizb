package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatValueSetAndString(t *testing.T) {
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

func TestLooksLikeStatValue(t *testing.T) {
	assert.False(t, looksLikeStatValue(""))
	assert.False(t, looksLikeStatValue("-h"))
	assert.True(t, looksLikeStatValue("all"))
	assert.True(t, looksLikeStatValue("counts,center"))
	assert.False(t, looksLikeStatValue("counts,bogus"))
}

func TestRewriteStatArg(t *testing.T) {
	assert.Equal(t, []string{"--stat=counts"}, RewriteStatArg([]string{"--stat", "counts"}))
	assert.Equal(t, []string{"--stat", "-o"}, RewriteStatArg([]string{"--stat", "-o"}))

	dir := t.TempDir()
	file := filepath.Join(dir, "bench.json")
	require.NoError(t, os.WriteFile(file, []byte("{}"), 0644))
	assert.Equal(t, []string{"--stat", file}, RewriteStatArg([]string{"--stat", file}))
}
