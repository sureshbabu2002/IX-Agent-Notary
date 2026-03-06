package receipt

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"strings"

	"ix-agent-notary/internal/canon"
	"ix-agent-notary/internal/crypto"
)

type SignatureValidationOptions struct {
	Strict        bool
	PublicKeyPath string // optional exact public key file (base64url)
	PublicKeyDir  string // optional directory containing <key_id>.pub
}

type SignatureCheck struct {
	Skipped bool
	Alg     string
	KeyID   string
}

func ValidateSignature(r Receipt, opts SignatureValidationOptions) (*SignatureCheck, error) {
	integrity, ok := r["integrity"].(map[string]any)
	if !ok {
		return nil, errors.New("missing integrity object")
	}

	sigObj, ok := integrity["signature"].(map[string]any)
	if !ok {
		return nil, errors.New("missing integrity.signature object")
	}

	alg, _ := sigObj["alg"].(string)
	keyID, _ := sigObj["key_id"].(string)
	val, _ := sigObj["value"].(string)

	alg = strings.ToLower(strings.TrimSpace(alg))
	keyID = strings.TrimSpace(keyID)
	val = strings.TrimSpace(val)

	if isPlaceholder(val) {
		if opts.Strict {
			return nil, fmt.Errorf("signature is missing/placeholder")
		}
		return &SignatureCheck{Skipped: true, Alg: alg, KeyID: keyID}, nil
	}

	if alg != "ed25519" {
		return nil, fmt.Errorf("unsupported signature alg: %q (only ed25519 supported right now)", alg)
	}
	if keyID == "" {
		return nil, fmt.Errorf("missing integrity.signature.key_id")
	}

	pub, _, err := crypto.ResolveEd25519PublicKey(crypto.ResolvePublicKeyOptions{
		KeyID:         keyID,
		PublicKeyPath: opts.PublicKeyPath,
		SearchDirs:    receiptPublicKeySearchDirs(opts.PublicKeyDir),
	})
	if err != nil {
		if opts.Strict {
			return nil, err
		}
		return &SignatureCheck{Skipped: true, Alg: alg, KeyID: keyID}, nil
	}

	sigBytes, err := crypto.DecodeBase64URLNoPad(val)
	if err != nil {
		return nil, fmt.Errorf("decode signature value: %w", err)
	}
	if len(sigBytes) != ed25519.SignatureSize {
		return nil, fmt.Errorf("invalid ed25519 signature length: %d", len(sigBytes))
	}

	msg, err := canonicalBytesForSignature(r)
	if err != nil {
		return nil, err
	}

	if !ed25519.Verify(pub, msg, sigBytes) {
		return nil, fmt.Errorf("signature verification failed (key_id=%s)", keyID)
	}

	return &SignatureCheck{Skipped: false, Alg: alg, KeyID: keyID}, nil
}

func canonicalBytesForSignature(r Receipt) ([]byte, error) {
	cloned, err := cloneReceipt(r)
	if err != nil {
		return nil, err
	}

	if integrity, ok := cloned["integrity"].(map[string]any); ok {
		if sigObj, ok := integrity["signature"].(map[string]any); ok {
			delete(sigObj, "value")
		}
	}

	b, err := canon.CanonicalizeRFC8785(cloned)
	if err != nil {
		return nil, fmt.Errorf("canonicalize receipt for signature: %w", err)
	}
	return b, nil
}

func cloneReceipt(r Receipt) (map[string]any, error) {
	b, err := marshalJSON(r)
	if err != nil {
		return nil, err
	}
	return unmarshalJSONObject(b)
}

func receiptPublicKeySearchDirs(dir string) []string {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil
	}
	return []string{dir}
}
