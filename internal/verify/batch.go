package verify

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"ix-agent-notary/internal/receipt"
)

type DirOptions struct {
	Dir             string
	SchemaPath      string
	PublicKeyPath   string
	StrictHashes    bool
	StrictSignature bool
	StrictApprovals bool
	StrictChain     bool
}

type DirResult struct {
	Dir      string
	Total    int
	OK       int
	Fail     int
	Failures []string
}

func VerifyDir(opts DirOptions) (*DirResult, error) {
	dir := strings.TrimSpace(opts.Dir)
	if dir == "" {
		return nil, fmt.Errorf("dir is required")
	}
	if opts.SchemaPath == "" {
		opts.SchemaPath = filepath.Join("spec", "receipt.schema.json")
	}

	// For “enterprise serious” posture, chain implies strict leaf validation.
	if opts.StrictChain {
		opts.StrictHashes = true
		opts.StrictSignature = true
	}

	schema, err := CompileSchema(opts.SchemaPath)
	if err != nil {
		return nil, err
	}

	// Build resolver once (used for chain checks).
	resolver, err := receipt.NewDirResolver(dir)
	if err != nil {
		return nil, err
	}

	// Collect JSON files deterministically.
	var files []string
	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			files = append(files, path)
		}
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk dir: %w", walkErr)
	}
	sort.Strings(files)

	res := &DirResult{Dir: dir, Total: len(files)}

	validateStrict := func(r receipt.Receipt) error {
		_, _, _, err := ValidateReceiptObject(r, schema, ReceiptValidationOptions{
			StrictHashes:    opts.StrictHashes,
			StrictSignature: opts.StrictSignature,
			StrictApprovals: opts.StrictApprovals,
			PublicKeyPath:   opts.PublicKeyPath,
		})
		return err
	}

	// Parent validator is always strict when StrictChain is on (and includes approvals if requested).
	validateParentStrict := func(r receipt.Receipt) error {
		_, _, _, err := ValidateReceiptObject(r, schema, ReceiptValidationOptions{
			StrictHashes:    true,
			StrictSignature: true,
			StrictApprovals: opts.StrictApprovals,
			PublicKeyPath:   opts.PublicKeyPath,
		})
		return err
	}

	for _, p := range files {
		r, err := receipt.Load(p)
		if err != nil {
			res.Fail++
			res.Failures = appendFailure(res.Failures, fmt.Sprintf("%s: load failed: %v", p, err))
			continue
		}

		if err := validateStrict(r); err != nil {
			res.Fail++
			res.Failures = appendFailure(res.Failures, fmt.Sprintf("%s: verify failed: %v", p, err))
			continue
		}

		if opts.StrictChain {
			_, cerr := receipt.ValidateChain(
				r,
				resolver,
				validateParentStrict,
				receipt.ChainValidationOptions{Strict: true},
			)
			if cerr != nil {
				res.Fail++
				res.Failures = appendFailure(res.Failures, fmt.Sprintf("%s: chain failed: %v", p, cerr))
				continue
			}
		}

		res.OK++
	}

	if res.Fail > 0 {
		return res, fmt.Errorf("verify-dir failed: %d/%d receipts failed", res.Fail, res.Total)
	}

	return res, nil
}

func appendFailure(f []string, s string) []string {
	const max = 20
	if len(f) < max {
		return append(f, s)
	}
	return f
}
