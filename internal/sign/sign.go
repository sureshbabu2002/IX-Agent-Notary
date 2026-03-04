package sign

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"ix-agent-notary/internal/canon"
	"ix-agent-notary/internal/crypto"
	"ix-agent-notary/internal/receipt"
)

type Options struct {
	InPath  string
	OutPath string
	KeyPath string
	KeyID   string
}

// SignReceiptInPlace computes core hashes, ensures the signature envelope,
// canonicalizes (excluding signature.value), and writes integrity.signature.value.
func SignReceiptInPlace(r receipt.Receipt, keyPath string, keyID string) error {
	if keyID == "" {
		return errors.New("keyID is required")
	}
	if keyPath == "" {
		keyPath = filepath.Join("keys", "dev", "dev-key-001.seed")
	}

	// Always compute and write core hashes (removes placeholders).
	hc, err := receipt.ComputeCoreHashes(r)
	if err != nil {
		return err
	}
	if err := setCoreHashes(r, hc); err != nil {
		return err
	}

	// Ensure integrity fields exist and set signature metadata.
	if err := ensureSignatureEnvelope(r, keyID); err != nil {
		return err
	}

	// Canonicalize (excluding signature.value), sign, then write signature.value.
	msg, err := canonicalForSigning(r)
	if err != nil {
		return err
	}

	priv, err := crypto.LoadEd25519PrivateKeyFromSeedFile(keyPath)
	if err != nil {
		return err
	}

	sig := ed25519.Sign(priv, msg)
	sigB64 := crypto.EncodeBase64URLNoPad(sig)

	if err := setSignatureValue(r, sigB64); err != nil {
		return err
	}

	return nil
}

func Run(opts Options) error {
	if opts.InPath == "" || opts.OutPath == "" {
		return errors.New("sign requires InPath and OutPath")
	}
	if opts.KeyID == "" {
		return errors.New("sign requires KeyID")
	}
	if opts.KeyPath == "" {
		opts.KeyPath = filepath.Join("keys", "dev", "dev-key-001.seed")
	}

	r, err := receipt.Load(opts.InPath)
	if err != nil {
		return err
	}

	if err := SignReceiptInPlace(r, opts.KeyPath, opts.KeyID); err != nil {
		return err
	}

	return receipt.Write(opts.OutPath, r)
}

func setCoreHashes(r receipt.Receipt, hc *receipt.HashCheck) error {
	a, ok := r["action"].(map[string]any)
	if !ok {
		return errors.New("missing action object")
	}
	res, ok := r["result"].(map[string]any)
	if !ok {
		return errors.New("missing result object")
	}
	a["parameters_hash"] = hc.ActionParametersComputed
	res["output_hash"] = hc.ResultOutputComputed
	return nil
}

func ensureSignatureEnvelope(r receipt.Receipt, keyID string) error {
	integrity, ok := r["integrity"].(map[string]any)
	if !ok {
		integrity = map[string]any{}
		r["integrity"] = integrity
	}

	// Preserve existing hash settings if present; set safe defaults if absent.
	if _, ok := integrity["canonicalization"]; !ok {
		integrity["canonicalization"] = "RFC8785-JCS"
	}
	if _, ok := integrity["hash"]; !ok {
		integrity["hash"] = map[string]any{"alg": "sha-256", "encoding": "base64url"}
	}

	sigObj, ok := integrity["signature"].(map[string]any)
	if !ok {
		sigObj = map[string]any{}
		integrity["signature"] = sigObj
	}
	sigObj["alg"] = "ed25519"
	sigObj["key_id"] = keyID
	// value is set after signing
	return nil
}

func canonicalForSigning(r receipt.Receipt) ([]byte, error) {
	// Clone to avoid mutating the receipt during canonicalization.
	b, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("marshal receipt: %w", err)
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("unmarshal receipt: %w", err)
	}

	root, ok := v.(map[string]any)
	if !ok {
		return nil, errors.New("receipt root must be object")
	}

	if integrity, ok := root["integrity"].(map[string]any); ok {
		if sigObj, ok := integrity["signature"].(map[string]any); ok {
			delete(sigObj, "value")
		}
	}

	out, err := canon.CanonicalizeRFC8785(root)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func setSignatureValue(r receipt.Receipt, sig string) error {
	integrity, ok := r["integrity"].(map[string]any)
	if !ok {
		return errors.New("missing integrity object")
	}
	sigObj, ok := integrity["signature"].(map[string]any)
	if !ok {
		return errors.New("missing integrity.signature object")
	}
	sigObj["value"] = sig
	return nil
}
