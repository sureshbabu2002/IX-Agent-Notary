package receipt

import (
	"path/filepath"
	"testing"

	"ix-agent-notary/internal/testutil"
)

func TestDeniedReceipt_ChainVerifies(t *testing.T) {
	root := testutil.RepoRoot(t)
	dir := filepath.Join(root, "examples", "receipts")

	leaf, err := Load(filepath.Join(dir, "denied.receipt.json"))
	if err != nil {
		t.Fatalf("Load leaf: %v", err)
	}

	resolver, err := NewDirResolver(dir)
	if err != nil {
		t.Fatalf("NewDirResolver: %v", err)
	}

	validateParent := func(r Receipt) error {
		if _, err := ValidateCoreHashes(r, HashValidationOptions{Strict: true}); err != nil {
			return err
		}
		if _, err := ValidateSignature(r, SignatureValidationOptions{Strict: true}); err != nil {
			return err
		}
		return nil
	}

	cc, err := ValidateChain(leaf, resolver, validateParent, ChainValidationOptions{Strict: true})
	if err != nil {
		t.Fatalf("ValidateChain: %v", err)
	}

	if cc.Skipped {
		t.Fatalf("expected chain not skipped")
	}
	if cc.Depth != 1 {
		t.Fatalf("expected depth=1, got %d", cc.Depth)
	}
	if cc.RootReceiptID == "" {
		t.Fatalf("expected non-empty root receipt id")
	}
}
