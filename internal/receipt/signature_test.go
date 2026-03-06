package receipt

import (
	"path/filepath"
	"testing"

	"ix-agent-notary/internal/testutil"
)

func TestGeneratedReceipts_StrictSignaturesPass(t *testing.T) {
	seedPath, pubPath := testutil.TempEd25519Keypair(t, receiptTestKeyID)

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

			if _, err := ValidateSignature(r, SignatureValidationOptions{
				Strict:       true,
				PublicKeyDir: filepath.Dir(pubPath),
			}); err != nil {
				t.Fatalf("ValidateSignature (strict): %v", err)
			}
		})
	}
}
