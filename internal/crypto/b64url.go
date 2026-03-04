package crypto

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func DecodeBase64URLNoPad(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, "=")

	// Accept both raw and padded base64url inputs.
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err == nil {
		return b, nil
	}

	// Fallback: add padding if needed and try URLEncoding.
	pad := len(s) % 4
	if pad != 0 {
		s = s + strings.Repeat("=", 4-pad)
	}
	b2, err2 := base64.URLEncoding.DecodeString(s)
	if err2 != nil {
		return nil, fmt.Errorf("decode base64url: %w", err)
	}
	return b2, nil
}

func EncodeBase64URLNoPad(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}
