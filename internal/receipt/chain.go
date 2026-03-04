package receipt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type ChainValidationOptions struct {
	// Strict makes missing parents, step mismatches, trace mismatches, cycles, etc. hard errors.
	Strict bool

	// MaxDepth prevents infinite walks on cycles or pathological chains.
	// If <= 0, defaults to 64.
	MaxDepth int
}

type ChainCheck struct {
	Skipped       bool
	Depth         int // number of parent links traversed
	RootReceiptID string
}

type ChainResolver interface {
	Resolve(receiptID string) (Receipt, string, error)
}

// DirResolver indexes receipts by receipt_id from a directory tree (recursively).
type DirResolver struct {
	dir  string
	byID map[string]string
}

func NewDirResolver(dir string) (*DirResolver, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, errors.New("dir resolver: dir is empty")
	}

	byID := map[string]string{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			return nil
		}

		rid, ok := tryExtractReceiptID(path)
		if !ok {
			return nil
		}
		// First wins; deterministic enough for a demo.
		if _, exists := byID[rid]; !exists {
			byID[rid] = path
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("dir resolver walk: %w", err)
	}

	return &DirResolver{dir: dir, byID: byID}, nil
}

func (dr *DirResolver) Resolve(receiptID string) (Receipt, string, error) {
	receiptID = strings.TrimSpace(receiptID)
	if receiptID == "" {
		return nil, "", errors.New("resolve: receiptID is empty")
	}

	p, ok := dr.byID[receiptID]
	if !ok {
		return nil, "", fmt.Errorf("resolve: receipt_id %q not found under %s", receiptID, dr.dir)
	}

	r, err := Load(p)
	if err != nil {
		return nil, "", err
	}
	return r, p, nil
}

// ValidateChain validates trace linkage by walking parent_receipt_id pointers until the root.
//
// validateFn is called on each resolved parent receipt (recommended: schema + hashes + signature).
// The leaf receipt is assumed already validated by the caller.
func ValidateChain(leaf Receipt, resolver ChainResolver, validateFn func(Receipt) error, opts ChainValidationOptions) (*ChainCheck, error) {
	if resolver == nil {
		return nil, errors.New("chain validation requires a resolver")
	}
	if opts.MaxDepth <= 0 {
		opts.MaxDepth = 64
	}

	leafID, err := getReceiptID(leaf)
	if err != nil {
		return nil, err
	}

	traceID, step, parentID, err := getTrace(leaf)
	if err != nil {
		return nil, err
	}

	visited := map[string]bool{leafID: true}
	depth := 0

	// No parent: enforce root step expectations in strict mode.
	if strings.TrimSpace(parentID) == "" {
		if opts.Strict && step != 1 {
			return nil, fmt.Errorf("chain: root receipt step must be 1 (got %d) for receipt_id=%s", step, leafID)
		}
		return &ChainCheck{Skipped: false, Depth: 0, RootReceiptID: leafID}, nil
	}

	childID := leafID
	childTraceID := traceID
	childStep := step
	curParentID := parentID

	for strings.TrimSpace(curParentID) != "" {
		if depth >= opts.MaxDepth {
			return nil, fmt.Errorf("chain: exceeded max depth (%d) starting at receipt_id=%s", opts.MaxDepth, leafID)
		}
		if visited[curParentID] {
			return nil, fmt.Errorf("chain: cycle detected at receipt_id=%s (starting leaf=%s)", curParentID, leafID)
		}
		visited[curParentID] = true

		parentR, _, err := resolver.Resolve(curParentID)
		if err != nil {
			if opts.Strict {
				return nil, err
			}
			return &ChainCheck{Skipped: true, Depth: depth, RootReceiptID: ""}, nil
		}

		if validateFn != nil {
			if err := validateFn(parentR); err != nil {
				return nil, fmt.Errorf("chain: parent receipt %s failed validation: %w", curParentID, err)
			}
		}

		parentIDActual, err := getReceiptID(parentR)
		if err != nil {
			return nil, err
		}
		if parentIDActual != curParentID {
			return nil, fmt.Errorf("chain: resolved parent receipt_id mismatch: expected %s got %s", curParentID, parentIDActual)
		}

		parentTraceID, parentStep, parentParentID, err := getTrace(parentR)
		if err != nil {
			return nil, err
		}

		if parentTraceID != childTraceID {
			return nil, fmt.Errorf("chain: trace_id mismatch: child(%s)=%s parent(%s)=%s", childID, childTraceID, parentIDActual, parentTraceID)
		}

		if parentStep != (childStep - 1) {
			return nil, fmt.Errorf("chain: step mismatch: child(%s).step=%d parent(%s).step=%d (expected %d)",
				childID, childStep, parentIDActual, parentStep, childStep-1)
		}

		// Advance up the chain.
		depth++
		childID = parentIDActual
		childStep = parentStep
		curParentID = parentParentID
	}

	// We reached a root (no parent).
	if opts.Strict && childStep != 1 {
		return nil, fmt.Errorf("chain: root receipt step must be 1 (got %d) for root receipt_id=%s", childStep, childID)
	}

	return &ChainCheck{Skipped: false, Depth: depth, RootReceiptID: childID}, nil
}

func getReceiptID(r Receipt) (string, error) {
	s, _ := r["receipt_id"].(string)
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("missing receipt_id")
	}
	return s, nil
}

func getTrace(r Receipt) (traceID string, step int, parentID string, err error) {
	t, ok := r["trace"].(map[string]any)
	if !ok {
		return "", 0, "", errors.New("missing trace object")
	}

	traceID, _ = t["trace_id"].(string)
	traceID = strings.TrimSpace(traceID)
	if traceID == "" {
		return "", 0, "", errors.New("missing trace.trace_id")
	}

	stepAny, ok := t["step"]
	if !ok {
		return "", 0, "", errors.New("missing trace.step")
	}

	step, err = toInt(stepAny)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid trace.step: %w", err)
	}

	parentID, _ = t["parent_receipt_id"].(string)
	parentID = strings.TrimSpace(parentID)

	return traceID, step, parentID, nil
}

func toInt(v any) (int, error) {
	switch t := v.(type) {
	case float64:
		// json.Unmarshal uses float64 for numbers
		if t != float64(int(t)) {
			return 0, fmt.Errorf("not an integer: %v", t)
		}
		return int(t), nil
	case int:
		return t, nil
	default:
		return 0, fmt.Errorf("unsupported number type: %T", v)
	}
}

func tryExtractReceiptID(path string) (string, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	var tmp struct {
		ReceiptID string `json:"receipt_id"`
	}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return "", false
	}

	rid := strings.TrimSpace(tmp.ReceiptID)
	if rid == "" {
		return "", false
	}
	return rid, true
}
