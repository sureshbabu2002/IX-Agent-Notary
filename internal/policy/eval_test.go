package policy

import (
	"path/filepath"
	"testing"

	"ix-agent-notary/internal/testutil"
)

func TestPolicyDemo_AllowsDocsWrite(t *testing.T) {
	root := testutil.RepoRoot(t)
	p, err := Load(filepath.Join(root, "policy", "demo.policy.json"))
	if err != nil {
		t.Fatalf("Load policy: %v", err)
	}

	dec := p.Evaluate(Request{
		Kind:      "tool.invoke",
		Tool:      "filesystem",
		Operation: "file.write",
		Path:      "docs/demo.txt",
	})

	if dec.Decision != "allow" {
		t.Fatalf("expected allow, got %q (reason=%q)", dec.Decision, dec.Reason)
	}
	if len(dec.Matched) != 1 || dec.Matched[0].RuleID != "fs-write-docs-only" {
		t.Fatalf("expected matched rule fs-write-docs-only, got %+v", dec.Matched)
	}
}

func TestPolicyDemo_DeniesDotEnv(t *testing.T) {
	root := testutil.RepoRoot(t)
	p, err := Load(filepath.Join(root, "policy", "demo.policy.json"))
	if err != nil {
		t.Fatalf("Load policy: %v", err)
	}

	dec := p.Evaluate(Request{
		Kind:      "tool.invoke",
		Tool:      "filesystem",
		Operation: "file.write",
		Path:      ".env",
	})

	if dec.Decision != "deny" {
		t.Fatalf("expected deny, got %q (reason=%q)", dec.Decision, dec.Reason)
	}
	if len(dec.Matched) != 1 || dec.Matched[0].RuleID != "deny-dotenv" {
		t.Fatalf("expected matched rule deny-dotenv, got %+v", dec.Matched)
	}
}

func TestPolicyDemo_DefaultDeny(t *testing.T) {
	root := testutil.RepoRoot(t)
	p, err := Load(filepath.Join(root, "policy", "demo.policy.json"))
	if err != nil {
		t.Fatalf("Load policy: %v", err)
	}

	dec := p.Evaluate(Request{
		Kind:      "tool.invoke",
		Tool:      "filesystem",
		Operation: "file.write",
		Path:      "tmp/anything.txt",
	})

	if dec.Decision != "deny" {
		t.Fatalf("expected deny (default), got %q (reason=%q)", dec.Decision, dec.Reason)
	}
	if len(dec.Matched) != 0 {
		t.Fatalf("expected no matched rules, got %+v", dec.Matched)
	}
}
