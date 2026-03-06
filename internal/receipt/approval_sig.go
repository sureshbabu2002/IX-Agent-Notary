package receipt

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"ix-agent-notary/internal/crypto"
)

type ApprovalSigValidationOptions struct {
	Strict bool

	// PublicKeyPath is the highest-precedence verification source.
	// Expected format: base64url (no padding) encoded 32-byte ed25519 public key.
	PublicKeyPath string

	// PublicKeyDir is an optional directory containing <key_id>.pub files.
	// If unset, default lookup falls back to keys/ then keys/dev/ relative to cwd.
	PublicKeyDir string
}

type ApprovalSigCheck struct {
	Skipped  bool
	Total    int
	Verified int
}

func ValidateApprovalSignatures(r Receipt, opts ApprovalSigValidationOptions) (*ApprovalSigCheck, error) {
	pol, ok := r["policy"].(map[string]any)
	if !ok {
		if opts.Strict {
			return nil, errors.New("approval sig: missing policy object")
		}
		return &ApprovalSigCheck{Skipped: true}, nil
	}

	apprsAny, ok := pol["approvals"]
	if !ok {
		if opts.Strict {
			return nil, errors.New("approval sig: missing policy.approvals")
		}
		return &ApprovalSigCheck{Skipped: true}, nil
	}

	apprs, ok := apprsAny.([]any)
	if !ok {
		if opts.Strict {
			return nil, errors.New("approval sig: policy.approvals is not an array")
		}
		return &ApprovalSigCheck{Skipped: true}, nil
	}

	if len(apprs) == 0 {
		return &ApprovalSigCheck{Skipped: false, Total: 0, Verified: 0}, nil
	}

	check := &ApprovalSigCheck{Skipped: false, Total: len(apprs), Verified: 0}

	for i, a := range apprs {
		obj, ok := a.(map[string]any)
		if !ok {
			if opts.Strict {
				return nil, fmt.Errorf("approval sig: approvals[%d] is not an object", i)
			}
			continue
		}

		sigObjAny, hasSig := obj["signature"]
		if !hasSig || sigObjAny == nil {
			if opts.Strict {
				return nil, fmt.Errorf("approval sig: approvals[%d] missing signature", i)
			}
			continue
		}

		sigObj, ok := sigObjAny.(map[string]any)
		if !ok {
			if opts.Strict {
				return nil, fmt.Errorf("approval sig: approvals[%d].signature is not an object", i)
			}
			continue
		}

		alg, _ := sigObj["alg"].(string)
		keyID, _ := sigObj["key_id"].(string)
		val, _ := sigObj["value"].(string)

		alg = strings.ToLower(strings.TrimSpace(alg))
		keyID = strings.TrimSpace(keyID)
		val = strings.TrimSpace(val)

		if alg != "ed25519" {
			return nil, fmt.Errorf("approval sig: approvals[%d] unsupported alg %q", i, alg)
		}
		if keyID == "" {
			return nil, fmt.Errorf("approval sig: approvals[%d] missing signature.key_id", i)
		}
		if val == "" {
			return nil, fmt.Errorf("approval sig: approvals[%d] missing signature.value", i)
		}

		payload, err := CanonicalizeApprovalForSigning(obj)
		if err != nil {
			return nil, fmt.Errorf("approval sig: approvals[%d] canonicalize: %w", i, err)
		}

		sigBytes, err := base64.RawURLEncoding.DecodeString(val)
		if err != nil {
			return nil, fmt.Errorf("approval sig: approvals[%d] signature.value not base64url: %w", i, err)
		}
		if len(sigBytes) != ed25519.SignatureSize {
			return nil, fmt.Errorf("approval sig: approvals[%d] signature size invalid (got %d)", i, len(sigBytes))
		}

		pub, _, err := crypto.ResolveEd25519PublicKey(crypto.ResolvePublicKeyOptions{
			KeyID:         keyID,
			PublicKeyPath: opts.PublicKeyPath,
			SearchDirs:    receiptPublicKeySearchDirs(opts.PublicKeyDir),
		})
		if err != nil {
			return nil, fmt.Errorf("approval sig: approvals[%d] resolve pubkey: %w", i, err)
		}

		if !ed25519.Verify(pub, payload, sigBytes) {
			return nil, fmt.Errorf("approval sig: approvals[%d] invalid signature", i)
		}

		check.Verified++
	}

	if opts.Strict && check.Verified != check.Total {
		return nil, fmt.Errorf("approval sig: strict mode requires all approvals be signed (%d/%d verified)", check.Verified, check.Total)
	}

	return check, nil
}
