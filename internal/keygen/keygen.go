package keygen

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

type Options struct {
	OutSeedPath string
	OutPubPath  string
	Force       bool
}

func GenerateEd25519Keypair(opts Options) error {
	if opts.OutSeedPath == "" {
		return fmt.Errorf("keygen: OutSeedPath is required")
	}
	if opts.OutPubPath == "" {
		return fmt.Errorf("keygen: OutPubPath is required")
	}

	if !opts.Force {
		if _, err := os.Stat(opts.OutSeedPath); err == nil {
			return fmt.Errorf("keygen: seed already exists (use --force to overwrite): %s", opts.OutSeedPath)
		}
		if _, err := os.Stat(opts.OutPubPath); err == nil {
			return fmt.Errorf("keygen: pub already exists (use --force to overwrite): %s", opts.OutPubPath)
		}
	}

	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		return fmt.Errorf("keygen: read random seed: %w", err)
	}

	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)

	if err := os.MkdirAll(filepath.Dir(opts.OutSeedPath), 0o755); err != nil {
		return fmt.Errorf("keygen: mkdir seed dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(opts.OutPubPath), 0o755); err != nil {
		return fmt.Errorf("keygen: mkdir pub dir: %w", err)
	}

	seedB64 := base64.RawURLEncoding.EncodeToString(seed)
	pubB64 := base64.RawURLEncoding.EncodeToString(pub)

	if err := writeFileAtomic(opts.OutSeedPath, []byte(seedB64+"\n"), 0o600); err != nil {
		return err
	}
	if err := writeFileAtomic(opts.OutPubPath, []byte(pubB64+"\n"), 0o644); err != nil {
		return err
	}

	return nil
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-ix-key-*")
	if err != nil {
		return fmt.Errorf("keygen: create temp file: %w", err)
	}
	tmpName := tmp.Name()

	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()

	if err := tmp.Chmod(perm); err != nil {
		return fmt.Errorf("keygen: chmod temp file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("keygen: write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("keygen: sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("keygen: close temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("keygen: rename temp -> %s: %w", path, err)
	}
	return nil
}
