package simulate

import (
	"os"
	"path/filepath"
	"testing"

	"ix-agent-notary/internal/receipt"
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

func TestSimulate_WithExplicitChainFields_ProducesStrictChainChild(t *testing.T) {
	root := testutil.RepoRoot(t)
	seedPath, pubPath := testutil.TempEd25519Keypair(t, "test-key-001")
	dir := t.TempDir()

	parentPath := filepath.Join(dir, "chain.root.receipt.json")
	childPath := filepath.Join(dir, "chain.child.receipt.json")

	if err := Run(Options{
		PolicyPath:  filepath.Join(root, "policy", "demo.policy.json"),
		OutPath:     parentPath,
		Kind:        "tool.invoke",
		Tool:        "filesystem",
		Operation:   "file.write",
		Path:        "docs/chain-root.txt",
		Bytes:       10,
		ActorID:     "agent:test",
		SessionID:   "sess-test-001",
		NotaryInst:  "notary-test-001",
		SignKeyPath: seedPath,
		SignKeyID:   "test-key-001",
	}); err != nil {
		t.Fatalf("simulate parent: %v", err)
	}

	parent, err := receipt.Load(parentPath)
	if err != nil {
		t.Fatalf("load parent: %v", err)
	}

	parentID := mustTestStringField(t, parent, "receipt_id")
	parentTrace := mustTestObjectField(t, parent, "trace")
	parentTraceID := mustTestStringField(t, parentTrace, "trace_id")

	if err := Run(Options{
		PolicyPath:      filepath.Join(root, "policy", "demo.policy.json"),
		OutPath:         childPath,
		Kind:            "tool.invoke",
		Tool:            "filesystem",
		Operation:       "file.write",
		Path:            "docs/chain-child.txt",
		Bytes:           10,
		ActorID:         "agent:test",
		SessionID:       "sess-test-001",
		NotaryInst:      "notary-test-001",
		SignKeyPath:     seedPath,
		SignKeyID:       "test-key-001",
		TraceID:         parentTraceID,
		Step:            2,
		ParentReceiptID: parentID,
	}); err != nil {
		t.Fatalf("simulate child: %v", err)
	}

	res, err := verify.Run(verify.Options{
		ReceiptPath:      childPath,
		SchemaPath:       filepath.Join(root, "spec", "receipt.schema.json"),
		StrictHashes:     true,
		StrictSignature:  true,
		PublicKeyPathOpt: pubPath,
		StrictChain:      true,
		ChainDir:         dir,
	})
	if err != nil {
		t.Fatalf("verify strict chain: %v", err)
	}

	if res.Chain.Skipped {
		t.Fatalf("expected chain verification to run")
	}
	if res.Chain.Depth != 1 {
		t.Fatalf("expected chain depth=1, got %d", res.Chain.Depth)
	}
	if res.Chain.RootReceiptID != parentID {
		t.Fatalf("expected root receipt id %q, got %q", parentID, res.Chain.RootReceiptID)
	}
}

func mustTestObjectField(t *testing.T, obj map[string]any, key string) map[string]any {
	t.Helper()

	v, ok := obj[key].(map[string]any)
	if !ok || v == nil {
		t.Fatalf("field %q is missing or not an object", key)
	}
	return v
}

func mustTestStringField(t *testing.T, obj map[string]any, key string) string {
	t.Helper()

	v, ok := obj[key].(string)
	if !ok || v == "" {
		t.Fatalf("field %q is missing or not a non-empty string", key)
	}
	return v
}
