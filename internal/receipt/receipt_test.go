package receipt

import (
	"path/filepath"
	"testing"

	"ix-agent-notary/internal/testutil"
)

func TestGeneratedReceipts_StrictHashesPass(t *testing.T) {
	seedPath, _ := testutil.TempEd25519Keypair(t, receiptTestKeyID)

	cases := []struct {
		name       string
		targetPath string
	}{
		{name: "allow", targetPath: "docs/demo.txt"},
		{name: "deny", targetPath: ".env"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tc.name+".receipt.json")
			r := writeSignedTestReceipt(t, path, tc.targetPath, seedPath, receiptTestKeyID)

			if _, err := ValidateCoreHashes(r, HashValidationOptions{Strict: true}); err != nil {
				t.Fatalf("ValidateCoreHashes (strict): %v", err)
			}
		})
	}
}

func TestGeneratedReceipt_ExpectedComputedHashesMatch(t *testing.T) {
	seedPath, _ := testutil.TempEd25519Keypair(t, receiptTestKeyID)

	r := newSignedTestReceipt(t, "docs/demo.txt", seedPath, receiptTestKeyID)

	h, err := ComputeCoreHashes(r)
	if err != nil {
		t.Fatalf("ComputeCoreHashes: %v", err)
	}

	if h.ActionParametersExpected != h.ActionParametersComputed {
		t.Fatalf("parameters_hash mismatch expected=%q computed=%q", h.ActionParametersExpected, h.ActionParametersComputed)
	}
	if h.ResultOutputExpected != h.ResultOutputComputed {
		t.Fatalf("output_hash mismatch expected=%q computed=%q", h.ResultOutputExpected, h.ResultOutputComputed)
	}
}
