package canon

import (
	"encoding/json"
	"fmt"

	"github.com/ucarion/jcs"
)

// CanonicalizeRFC8785 converts a JSON value into RFC 8785 (JCS) canonical bytes.
//
// Input constraints:
// - Values must be JSON-compatible types as produced by json.Unmarshal:
//   bool, float64, string, []any, map[string]any, nil.
func CanonicalizeRFC8785(v any) ([]byte, error) {
	// If caller passed raw JSON bytes/string, decode first.
	switch t := v.(type) {
	case []byte:
		var vv any
		if err := json.Unmarshal(t, &vv); err != nil {
			return nil, fmt.Errorf("parse json bytes: %w", err)
		}
		v = vv
	case string:
		var vv any
		if err := json.Unmarshal([]byte(t), &vv); err != nil {
			return nil, fmt.Errorf("parse json string: %w", err)
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
