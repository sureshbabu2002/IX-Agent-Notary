package receipt

import (
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"strings"
	"testing"

	"ix-agent-notary/internal/testutil"
)

func TestApprovalSignature_RoundTrip_SignAndVerify(t *testing.T) {
	seedPath, _ := testutil.TempEd25519Keypair(t, receiptTestKeyID)

	approval := map[string]any{
		"approval_id": "appr-001",
		"type":        "ticket",
		"status":      "approved",
		"approver": map[string]any{
			"type":    "user",
			"id":      "you@example.com",
			"display": "You",
		},
		"scope": map[string]any{
			"kind":      "tool.invoke",
			"tool":      "filesystem",
			"operation": "file.write",
			"resource":  "docs/demo.txt",
		},
		"time": map[string]any{
			"requested_at": "2026-03-02T00:00:00Z",
			"decided_at":   "2026-03-02T00:00:01Z",
		},
		"notes": "demo",
		"signature": map[string]any{
			"alg":    "ed25519",
			"key_id": receiptTestKeyID,
		},
	}

	seedB, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatalf("read seed: %v", err)
	}
	seed, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(string(seedB)))
	if err != nil {
		t.Fatalf("decode seed: %v", err)
	}
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)

	payload, err := CanonicalizeApprovalForSigning(approval)
	if err != nil {
		t.Fatalf("CanonicalizeApprovalForSigning: %v", err)
	}

	sig := ed25519.Sign(priv, payload)
	approval["signature"].(map[string]any)["value"] = base64.RawURLEncoding.EncodeToString(sig)

	sigObj := approval["signature"].(map[string]any)
	sigB64 := sigObj["value"].(string)
	sigBytes, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		t.Fatalf("decode signature: %v", err)
	}

	if !ed25519.Verify(pub, payload, sigBytes) {
		t.Fatalf("expected signature to verify")
	}
}
