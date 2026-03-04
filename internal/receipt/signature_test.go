package receipt

import (
	"path/filepath"
	"testing"

	"ix-agent-notary/internal/testutil"
)

func TestExamples_StrictSignaturesPass(t *testing.T) {
	root := testutil.RepoRoot(t)

	cases := []struct {
		name string
		path string
	}{
		{"minimal", filepath.Join(root, "examples", "receipts", "minimal.receipt.json")},
		{"denied", filepath.Join(root, "examples", "receipts", "denied.receipt.json")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Load(tc.path)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}

			if _, err := ValidateSignature(r, SignatureValidationOptions{Strict: true}); err != nil {
				t.Fatalf("ValidateSignature (strict): %v", err)
			}
		})
	}
}
