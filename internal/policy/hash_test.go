package policy

import (
	"path/filepath"
	"strings"
	"testing"

	"ix-agent-notary/internal/testutil"
)

func TestPolicyHash_IsStableAcrossLoads(t *testing.T) {
	root := testutil.RepoRoot(t)
	p1, err := Load(filepath.Join(root, "policy", "demo.policy.json"))
	if err != nil {
		t.Fatalf("Load #1: %v", err)
	}
	p2, err := Load(filepath.Join(root, "policy", "demo.policy.json"))
	if err != nil {
		t.Fatalf("Load #2: %v", err)
	}

	if p1.PolicyHash == "" || !strings.HasPrefix(p1.PolicyHash, "sha256:") {
		t.Fatalf("expected policy hash to start with sha256:, got %q", p1.PolicyHash)
	}
	if p1.PolicyHash != p2.PolicyHash {
		t.Fatalf("expected stable hash across loads; got %q vs %q", p1.PolicyHash, p2.PolicyHash)
	}
	if p1.SourcePath == "" {
		t.Fatalf("expected SourcePath to be set")
	}
}
