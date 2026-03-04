package simulate

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"ix-agent-notary/internal/canon"
	"ix-agent-notary/internal/hash"
	"ix-agent-notary/internal/id"
	"ix-agent-notary/internal/policy"
	"ix-agent-notary/internal/receipt"
	"ix-agent-notary/internal/sign"
)

type Options struct {
	PolicyPath string
	OutPath    string

	Kind      string
	Tool      string
	Operation string
	Path      string
	Bytes     int

	ActorID    string
	SessionID  string
	NotaryInst string

	SignKeyPath string
	SignKeyID   string

	// Optional governance signal:
	// If true, embed a single structured approval record in policy.approvals[].
	IncludeApproval bool
	ApproverID      string // e.g. user/email/IAM principal
	ApprovalType    string // human | ticket | breakglass (matches schema)
}

func Run(opts Options) error {
	if opts.PolicyPath == "" {
		return errors.New("PolicyPath is required")
	}
	if opts.OutPath == "" {
		return errors.New("OutPath is required")
	}
	if opts.Kind == "" {
		opts.Kind = "tool.invoke"
	}
	if opts.Tool == "" || opts.Operation == "" {
		return errors.New("Tool and Operation are required")
	}
	if opts.Path == "" {
		return errors.New("Path is required")
	}
	if opts.ActorID == "" {
		opts.ActorID = "agent:demo"
	}
	if opts.NotaryInst == "" {
		opts.NotaryInst = "notary-local-001"
	}
	if opts.SignKeyID == "" {
		opts.SignKeyID = "dev-key-001"
	}

	if opts.IncludeApproval {
		if strings.TrimSpace(opts.ApproverID) == "" {
			opts.ApproverID = "user:approver-demo"
		}
		if strings.TrimSpace(opts.ApprovalType) == "" {
			opts.ApprovalType = "human"
		}
		if !isValidApprovalType(opts.ApprovalType) {
			return fmt.Errorf("invalid ApprovalType %q (allowed: human|ticket|breakglass)", opts.ApprovalType)
		}
	}

	p, err := policy.Load(opts.PolicyPath)
	if err != nil {
		return err
	}

	dec := p.Evaluate(policy.Request{
		Kind:      opts.Kind,
		Tool:      opts.Tool,
		Operation: opts.Operation,
		Path:      opts.Path,
	})

	r, err := buildReceipt(opts, dec)
	if err != nil {
		return err
	}

	// If approvals are present, sign them (demo uses same key as receipt signing).
	if opts.IncludeApproval {
		pol, _ := r["policy"].(map[string]any)
		if pol != nil {
			apprs, _ := pol["approvals"].([]any)
			for i := range apprs {
				obj, ok := apprs[i].(map[string]any)
				if !ok || obj == nil {
					continue
				}
				if err := sign.SignApprovalInPlace(obj, opts.SignKeyPath, opts.SignKeyID); err != nil {
					return fmt.Errorf("sign approval: %w", err)
				}
			}
		}
	}

	// Sign and write (computes core hashes too).
	if err := sign.SignReceiptInPlace(r, opts.SignKeyPath, opts.SignKeyID); err != nil {
		return err
	}

	return receipt.Write(opts.OutPath, r)
}

func isValidApprovalType(t string) bool {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "human", "ticket", "breakglass":
		return true
	default:
		return false
	}
}

// buildReceipt constructs a receipt with policy decision + simulated result.
// Hashes/signature are added later by SignReceiptInPlace.
func buildReceipt(opts Options, dec policy.Decision) (receipt.Receipt, error) {
	now := time.Now().UTC()
	t := now.Format(time.RFC3339)

	receiptID, err := id.NewUUIDv4()
	if err != nil {
		return nil, err
	}
	traceID, err := id.NewUUIDv4()
	if err != nil {
		return nil, err
	}

	// Action parameters (simulated)
	params := map[string]any{
		"path":             opts.Path,
		"bytes":            opts.Bytes,
		"content_redacted": true,
	}

	// Result (simulated)
	status := "denied"
	summary := fmt.Sprintf("Denied %s %s (simulated).", opts.Tool, opts.Operation)
	output := map[string]any{
		"written": false,
		"denied":  true,
	}

	if dec.Decision == "allow" {
		status = "success"
		summary = fmt.Sprintf("Allowed %s %s (simulated).", opts.Tool, opts.Operation)
		output = map[string]any{
			"path":             opts.Path,
			"written":          true,
			"content_redacted": true,
		}
	}

	// Context hash (requested_path) — canonicalize the string value and hash it.
	reqPathHash, err := hashValueString(opts.Path, "base64url")
	if err != nil {
		return nil, err
	}

	// Policy rules evidence (schema expects policy.rules even if empty)
	rules := []map[string]any{}
	for _, mr := range dec.Matched {
		rules = append(rules, map[string]any{
			"rule_id":     mr.RuleID,
			"effect":      mr.Effect,
			"explanation": mr.Explanation,
		})
	}

	approvals := []any{}
	if opts.IncludeApproval {
		aid, err := id.NewUUIDv4()
		if err != nil {
			return nil, err
		}

		approvals = append(approvals, map[string]any{
			"approval_id": aid,
			"type":        strings.ToLower(strings.TrimSpace(opts.ApprovalType)),
			"status":      "approved",
			"approver": map[string]any{
				"type":    "user",
				"id":      strings.TrimSpace(opts.ApproverID),
				"display": "Demo Approver",
			},
			"scope": map[string]any{
				"kind":      opts.Kind,
				"tool":      opts.Tool,
				"operation": opts.Operation,
				"resource":  opts.Path,
			},
			"time": map[string]any{
				"requested_at": t,
				"decided_at":   t,
			},
			"notes": "Simulated approval record (demo).",
		})
	}

	policyObj := map[string]any{
		"policy_id": dec.PolicyID,
		"decision":  dec.Decision,
		"reason":    dec.Reason,
		"rules":     rules,
		"approvals": approvals,
		"context_hashes": map[string]any{
			"requested_path": reqPathHash,
		},
	}

	// Policy pack integrity (optional but recommended):
	// proves which exact policy pack produced the decision.
	if dec.PolicyHash != "" {
		policyObj["policy_hash"] = dec.PolicyHash
	}
	if dec.PolicySource != "" {
		policyObj["policy_source"] = dec.PolicySource
	}

	r := receipt.Receipt{
		"receipt_version": "0.1.0",
		"receipt_id":      receiptID,

		"time": map[string]any{
			"requested_at": t,
			"decided_at":   t,
			"completed_at": t,
		},

		"trace": map[string]any{
			"trace_id": traceID,
			"step":     1,
		},

		"actor": map[string]any{
			"type":       "agent",
			"id":         opts.ActorID,
			"display":    "Demo Agent",
			"session_id": opts.SessionID,
		},

		"notary": map[string]any{
			"runtime":     "IX-Agent-Notary",
			"version":     "0.1.0-dev",
			"instance_id": opts.NotaryInst,
			"environment": "local",
		},

		"action": map[string]any{
			"kind":            opts.Kind,
			"tool":            opts.Tool,
			"operation":       opts.Operation,
			"parameters":      params,
			"parameters_hash": "sha256:PLACEHOLDER_PARAMETERS_HASH",
		},

		"policy": policyObj,

		"result": map[string]any{
			"status":      status,
			"summary":     summary,
			"output":      output,
			"output_hash": "sha256:PLACEHOLDER_OUTPUT_HASH",
		},

		"integrity": map[string]any{
			"canonicalization": "RFC8785-JCS",
			"hash": map[string]any{
				"alg":      "sha-256",
				"encoding": "base64url",
			},
			"signature": map[string]any{
				"alg":    "ed25519",
				"key_id": opts.SignKeyID,
				"value":  "BASE64URL_SIGNATURE_PLACEHOLDER",
			},
		},
	}

	return r, nil
}

func hashValueString(s string, encoding string) (string, error) {
	enc, err := hash.ParseEncoding(encoding)
	if err != nil {
		return "", err
	}
	cbytes, err := canon.CanonicalizeRFC8785(s)
	if err != nil {
		return "", err
	}
	d := hash.Sha256Digest(cbytes)
	ds, err := hash.EncodeDigest(d, enc)
	if err != nil {
		return "", err
	}
	return "sha256:" + ds, nil
}
