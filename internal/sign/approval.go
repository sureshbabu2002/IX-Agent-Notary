package sign

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"ix-agent-notary/internal/receipt"
)

// SignApprovalInPlace signs a single approval object and writes/overwrites approval.signature.
//
// Signature payload = RFC8785 canonical JSON of the approval object excluding signature.value,
// but INCLUDING signature.alg and signature.key_id (so verifiers can bind metadata to the signature).
func SignApprovalInPlace(approval map[string]any, seedPath string, keyID string) error {
	if approval == nil {
		return errors.New("approval sign: approval is nil")
	}
	seedPath = strings.TrimSpace(seedPath)
	if seedPath == "" {
		return errors.New("approval sign: seedPath is empty")
	}
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		return errors.New("approval sign: keyID is empty")
	}

	seedB64, err := os.ReadFile(seedPath)
	if err != nil {
		return fmt.Errorf("approval sign: read seed: %w", err)
	}

	seed, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(string(seedB64)))
	if err != nil {
		return fmt.Errorf("approval sign: decode seed (base64url): %w", err)
	}
	if len(seed) != ed25519.SeedSize {
		return fmt.Errorf("approval sign: seed must be %d bytes (got %d)", ed25519.SeedSize, len(seed))
	}

	priv := ed25519.NewKeyFromSeed(seed)

	// IMPORTANT: include signature metadata in the signed payload, excluding only signature.value.
	approval["signature"] = map[string]any{
		"alg":    "ed25519",
		"key_id": keyID,
	}

	payload, err := receipt.CanonicalizeApprovalForSigning(approval)
	if err != nil {
		return fmt.Errorf("approval sign: canonicalize: %w", err)
	}

	sig := ed25519.Sign(priv, payload)
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	approval["signature"] = map[string]any{
		"alg":    "ed25519",
		"key_id": keyID,
		"value":  sigB64,
	}

	return nil
}
