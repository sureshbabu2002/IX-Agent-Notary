package verify

import (
	"path/filepath"
	"testing"

	"ix-agent-notary/internal/receipt"
	"ix-agent-notary/internal/sign"
	"ix-agent-notary/internal/simulate"
	"ix-agent-notary/internal/testutil"
)

const batchTestKeyID = "test-key-001"

func TestVerifyDir_GeneratedReceipts_StrictChain(t *testing.T) {
	root := testutil.RepoRoot(t)
	dir, pubPath := buildVerifyDirChainFixture(t, root)

	_, err := VerifyDir(DirOptions{
		Dir:             dir,
		SchemaPath:      filepath.Join(root, "spec", "receipt.schema.json"),
		PublicKeyDir:    filepath.Dir(pubPath),
		StrictHashes:    true,
		StrictSignature: true,
		StrictApprovals: false,
		StrictChain:     true,
	})
	if err != nil {
		t.Fatalf("VerifyDir failed: %v", err)
	}
}

func buildVerifyDirChainFixture(t *testing.T, root string) (dir string, pubPath string) {
	t.Helper()

	seedPath, pubPath := testutil.TempEd25519Keypair(t, batchTestKeyID)
	dir = t.TempDir()

	parentPath := filepath.Join(dir, "parent.receipt.json")
	childPath := filepath.Join(dir, "child.receipt.json")

	writeVerifyDirReceipt(t, root, parentPath, "docs/parent.txt", seedPath)
	writeVerifyDirReceipt(t, root, childPath, "docs/child.txt", seedPath)

	parent, err := receipt.Load(parentPath)
	if err != nil {
		t.Fatalf("Load parent: %v", err)
	}
	child, err := receipt.Load(childPath)
	if err != nil {
		t.Fatalf("Load child: %v", err)
	}

	parentID := mustVerifyDirString(t, parent, "receipt_id")
	parentTrace := mustVerifyDirObject(t, parent, "trace")
	parentTraceID := mustVerifyDirString(t, parentTrace, "trace_id")

	childTrace := mustVerifyDirObject(t, child, "trace")
	childTrace["trace_id"] = parentTraceID
	childTrace["step"] = 2
	childTrace["parent_receipt_id"] = parentID

	if err := sign.SignReceiptInPlace(child, seedPath, batchTestKeyID); err != nil {
		t.Fatalf("SignReceiptInPlace child: %v", err)
	}
	if err := receipt.Write(childPath, child); err != nil {
		t.Fatalf("Write child: %v", err)
	}

	return dir, pubPath
}

func writeVerifyDirReceipt(t *testing.T, root string, outPath string, targetPath string, seedPath string) {
	t.Helper()

	if err := simulate.Run(simulate.Options{
		PolicyPath:  filepath.Join(root, "policy", "demo.policy.json"),
		OutPath:     outPath,
		Kind:        "tool.invoke",
		Tool:        "filesystem",
		Operation:   "file.write",
		Path:        targetPath,
		Bytes:       10,
		ActorID:     "agent:test",
		SessionID:   "sess-test-001",
		NotaryInst:  "notary-test-001",
		SignKeyPath: seedPath,
		SignKeyID:   batchTestKeyID,
	}); err != nil {
		t.Fatalf("simulate.Run %s: %v", outPath, err)
	}
}

func mustVerifyDirObject(t *testing.T, obj map[string]any, key string) map[string]any {
	t.Helper()

	v, ok := obj[key].(map[string]any)
	if !ok || v == nil {
		t.Fatalf("field %q is missing or not an object", key)
	}
	return v
}

func mustVerifyDirString(t *testing.T, obj map[string]any, key string) string {
	t.Helper()

	v, ok := obj[key].(string)
	if !ok || v == "" {
		t.Fatalf("field %q is missing or not a non-empty string", key)
	}
	return v
}
