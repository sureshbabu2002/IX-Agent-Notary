package canon

import (
	"encoding/json"
	"fmt"

	"github.com/ucarion/jcs"
)

// CanonicalizeRFC8785 converts a JSON value into RFC 8785 (JCS) canonical bytes.
//
// Important:
//   - Pass native JSON-compatible Go values (string, float64, bool, nil, []any, map[string]any),
//     typically produced by json.Unmarshal.
//   - If you have raw JSON bytes, pass them as []byte and they will be decoded first.
func CanonicalizeRFC8785(v any) ([]byte, error) {
	// If caller passed raw JSON bytes, decode first.
	if t, ok := v.([]byte); ok {
		var vv any
		if err := json.Unmarshal(t, &vv); err != nil {
			return nil, fmt.Errorf("parse json bytes: %w", err)
		}
		v = vv
	}

	var buf []byte
	buf, err := jcs.Append(buf, v)
	if err != nil {
		return nil, fmt.Errorf("canonicalize (rfc8785): %w", err)
	}
	return buf, nil
}
