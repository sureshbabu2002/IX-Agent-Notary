package receipt

import (
	"encoding/json"
	"errors"
	"fmt"

	"ix-agent-notary/internal/canon"
)

// CanonicalizeApprovalForSigning prepares a policy.approvals[] object for signing
// by removing approval.signature.value (if present), then RFC8785-canonicalizing.
//
// This is intentionally a small helper so future work can add:
// - approval signature verification
// - multi-party approvals / quorum rules
func CanonicalizeApprovalForSigning(approval any) ([]byte, error) {
	// Deep-copy via JSON roundtrip so we can safely mutate.
	b, err := json.Marshal(approval)
	if err != nil {
		return nil, fmt.Errorf("marshal approval: %w", err)
	}

	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("unmarshal approval: %w", err)
	}

	obj, ok := v.(map[string]any)
	if !ok {
		return nil, errors.New("approval must be a JSON object")
	}

	// Remove signature.value so it is not self-referential.
	if sig, ok := obj["signature"].(map[string]any); ok {
		delete(sig, "value")
	}

	out, err := canon.CanonicalizeRFC8785(obj)
	if err != nil {
		return nil, err
	}
	return out, nil
}
