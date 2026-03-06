package simulate

import (
	"os"
	"path/filepath"
	"testing"

	"ix-agent-notary/internal/testutil"
	"ix-agent-notary/internal/verify"
)

func TestSimulate_WithApproval_ProducesStrictVerifiableReceipt(t *testing.T) {
	root := testutil.RepoRoot(t)
	seedPath, pubPath := testutil.TempEd25519Keypair(t, "test-key-001")

	out := filepath.Join(t.TempDir(), "approved.receipt.json")

	if err := Run(Options{
		PolicyPath:      filepath.Join(root, "policy", "demo.policy.json"),
		OutPath:         out,
		Kind:            "tool.invoke",
		Tool:            "filesystem",
		Operation:       "file.write",
		Path:            "docs/demo.txt",
		Bytes:           10,
		ActorID:         "agent:test",
		SessionID:       "sess-test-001",
		NotaryInst:      "notary-test-001",
		SignKeyPath:     seedPath,
		SignKeyID:       "test-key-001",
		IncludeApproval: true,
		ApproverID:      "user:test-approver",
		ApprovalType:    "human",
	}); err != nil {
		t.Fatalf("simulate run: %v", err)
	}

	if _, err := os.Stat(out); err != nil {
		t.Fatalf("expected receipt file to exist: %v", err)
	}

	if _, err := verify.Run(verify.Options{
		ReceiptPath:      out,
		SchemaPath:       filepath.Join(root, "spec", "receipt.schema.json"),
		StrictHashes:     true,
		StrictSignature:  true,
		StrictApprovals:  true,
		PublicKeyPathOpt: pubPath,
	}); err != nil {
		t.Fatalf("verify strict: %v", err)
	}
}
