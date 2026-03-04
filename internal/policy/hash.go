package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"ix-agent-notary/internal/canon"
	"ix-agent-notary/internal/hash"
)

func ComputePolicyHashFile(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("policy hash: path is empty")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("policy hash: read file: %w", err)
	}

	return ComputePolicyHashJSON(b)
}

// ComputePolicyHashJSON returns a stable content hash for a policy pack:
//
// Algorithm (normative for this repo):
// 1) Parse JSON into a native value
// 2) Canonicalize using RFC8785 (JCS)
// 3) sha-256 digest
// 4) base64url (no padding)
// 5) prefix with "sha256:"
func ComputePolicyHashJSON(policyJSON []byte) (string, error) {
	var v any
	if err := json.Unmarshal(policyJSON, &v); err != nil {
		return "", fmt.Errorf("policy hash: parse json: %w", err)
	}

	cbytes, err := canon.CanonicalizeRFC8785(v)
	if err != nil {
		return "", fmt.Errorf("policy hash: canonicalize: %w", err)
	}

	enc, err := hash.ParseEncoding("base64url")
	if err != nil {
		return "", fmt.Errorf("policy hash: encoding: %w", err)
	}

	d := hash.Sha256Digest(cbytes)
	ds, err := hash.EncodeDigest(d, enc)
	if err != nil {
		return "", fmt.Errorf("policy hash: encode digest: %w", err)
	}

	return "sha256:" + ds, nil
}
