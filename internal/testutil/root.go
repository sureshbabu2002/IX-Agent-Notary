package testutil

import (
	"path/filepath"
	"runtime"
	"testing"

	"ix-agent-notary/internal/keygen"
)

func RepoRoot(t *testing.T) string {
	t.Helper()

	// This file lives at: <root>/internal/testutil/root.go
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TempEd25519Keypair(t *testing.T, keyID string) (seedPath string, pubPath string) {
	t.Helper()

	if keyID == "" {
		keyID = "test-key-001"
	}

	dir := t.TempDir()
	seedPath = filepath.Join(dir, keyID+".seed")
	pubPath = filepath.Join(dir, keyID+".pub")

	if err := keygen.GenerateEd25519Keypair(keygen.Options{
		OutSeedPath: seedPath,
		OutPubPath:  pubPath,
		Force:       true,
	}); err != nil {
		t.Fatalf("GenerateEd25519Keypair: %v", err)
	}

	return seedPath, pubPath
}
