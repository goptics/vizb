package json

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SelectPath reads the JSON file, navigates to the node named by a jq-like dot
// path, and returns it re-marshalled as a JSON array ready for ParseJSON.
//
// Path grammar (a deliberate subset of jq): an optional leading '.', then
// segments separated by '.'. A segment is an object key, optionally followed by
// one or more '[n]' array indices; a bare trailing '[]' is identity sugar for
// "the array itself". Examples: ".data.results", ".runs[0].samples", "results[]".
//
// The resolved node is coerced to an array: an array is used as-is; a single
// object is wrapped into a one-element array (so a path to one object still
// charts); anything else is an error.
//
// ponytail: whole-file load is opt-in (--json-path only); the default JSON path
// stays streaming. Upgrade to a streaming seek only if huge enveloped files
// become a real case.
func SelectPath(filename, path string) ([]byte, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON: %w", err)
	}
	return SelectBytes(raw, path)
}

// SelectBytes applies a json path to request-scoped JSON without reading from
// the filesystem. It is the safe counterpart to SelectPath.
func SelectBytes(raw []byte, path string) ([]byte, error) {

	var root any
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	node, err := navigate(root, path)
	if err != nil {
		return nil, err
	}

	switch v := node.(type) {
	case []any:
		return json.Marshal(v)
	case map[string]any:
		return json.Marshal([]any{v})
	default:
		return nil, fmt.Errorf("--json-path '%s' resolves to a scalar, not an array or object", path)
	}
}

// navigate walks node following the path segments.
func navigate(node any, path string) (any, error) {
	for _, seg := range tokenize(path) {
		var err error
		if node, err = step(node, seg); err != nil {
			return nil, err
		}
	}
	return node, nil
}

// pathSeg is one key plus any array indices that follow it. A leading index
// (key == "" with indices) handles a path that starts at an array, and the
// "[]" identity sugar yields a key-less, index-less no-op.
type pathSeg struct {
	key     string
	indices []int
}

// tokenize splits a dot path into segments. It is lenient: a leading '.' and a
// trailing "[]" are stripped, and empty pieces are skipped.
func tokenize(path string) []pathSeg {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, ".")

	var segs []pathSeg
	for _, piece := range strings.Split(path, ".") {
		if piece == "" {
			continue
		}

		key := piece
		var indices []int
		if i := strings.IndexByte(piece, '['); i >= 0 {
			key = piece[:i]
			rest := piece[i:]
			for rest != "" {
				end := strings.IndexByte(rest, ']')
				if end <= 1 { // "[]" identity or malformed → skip
					rest = rest[min(end+1, len(rest)):]
					continue
				}
				if n, err := strconv.Atoi(rest[1:end]); err == nil {
					indices = append(indices, n)
				}
				rest = rest[end+1:]
			}
		}

		if key == "" && len(indices) == 0 {
			continue
		}
		segs = append(segs, pathSeg{key: key, indices: indices})
	}
	return segs
}

// step applies a single segment (key lookup then any array indices).
func step(node any, seg pathSeg) (any, error) {
	if seg.key != "" {
		obj, ok := node.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("--json-path: cannot read key '%s' from a non-object", seg.key)
		}
		v, ok := obj[seg.key]
		if !ok {
			return nil, fmt.Errorf("--json-path: key '%s' not found", seg.key)
		}
		node = v
	}

	for _, idx := range seg.indices {
		arr, ok := node.([]any)
		if !ok {
			return nil, fmt.Errorf("--json-path: cannot index [%d] into a non-array", idx)
		}
		if idx < 0 || idx >= len(arr) {
			return nil, fmt.Errorf("--json-path: index [%d] out of range (len %d)", idx, len(arr))
		}
		node = arr[idx]
	}
	return node, nil
}
