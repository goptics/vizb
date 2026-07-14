package json

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelectPath(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		path    string
		want    string // expected marshalled array; "" when wantErr
		wantErr string // substring expected in the error
	}{
		{
			name:  "nested keys",
			input: `{"data":{"results":[{"a":1},{"a":2}]}}`,
			path:  ".data.results",
			want:  `[{"a":1},{"a":2}]`,
		},
		{
			name:  "array index then key",
			input: `{"runs":[{"samples":[{"n":1}]}]}`,
			path:  ".runs[0].samples",
			want:  `[{"n":1}]`,
		},
		{
			name:  "trailing [] identity sugar",
			input: `{"items":[{"x":1}]}`,
			path:  ".items[]",
			want:  `[{"x":1}]`,
		},
		{
			name:  "leading dot optional",
			input: `{"rows":[{"x":1}]}`,
			path:  "rows",
			want:  `[{"x":1}]`,
		},
		{
			name:  "single object wraps into array",
			input: `{"only":{"x":3,"y":7}}`,
			path:  ".only",
			want:  `[{"x":3,"y":7}]`,
		},
		{
			name:    "missing key",
			input:   `{"data":[]}`,
			path:    ".nope",
			wantErr: "key 'nope' not found",
		},
		{
			name:    "index out of range",
			input:   `{"runs":[]}`,
			path:    ".runs[0]",
			wantErr: "out of range",
		},
		{
			name:    "key into non-object",
			input:   `{"n":5}`,
			path:    ".n.deeper",
			wantErr: "non-object",
		},
		{
			name:    "index into non-array",
			input:   `{"n":5}`,
			path:    ".n[0]",
			wantErr: "non-array",
		},
		{
			name:    "path resolves to scalar",
			input:   `{"n":5}`,
			path:    ".n",
			wantErr: "scalar",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := filepath.Join(t.TempDir(), "in.json")
			require.NoError(t, os.WriteFile(f, []byte(tc.input), 0o644))

			got, err := SelectPath(f, tc.path)
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.JSONEq(t, tc.want, string(got))
		})
	}
}

func TestSelectPathReadAndParseErrors(t *testing.T) {
	_, err := SelectPath(filepath.Join(t.TempDir(), "missing.json"), ".rows")
	require.Error(t, err)
	require.Contains(t, err.Error(), "error reading JSON")

	f := filepath.Join(t.TempDir(), "bad.json")
	require.NoError(t, os.WriteFile(f, []byte(`{"rows":`), 0o644))
	_, err = SelectPath(f, ".rows")
	require.Error(t, err)
	require.Contains(t, err.Error(), "error parsing JSON")
}

func TestTokenizeLenientMalformedSegments(t *testing.T) {
	require.Empty(t, tokenize("[]"))

	gapped := tokenize(".rows..items")
	require.Len(t, gapped, 2)
	require.Equal(t, "rows", gapped[0].key)
	require.Equal(t, "items", gapped[1].key)

	got := tokenize(".rows[].items[x][1]")
	require.Len(t, got, 2)
	require.Equal(t, "rows", got[0].key)
	require.Empty(t, got[0].indices)
	require.Equal(t, "items", got[1].key)
	require.Equal(t, []int{1}, got[1].indices)
}
