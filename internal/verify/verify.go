package verify

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"ix-agent-notary/internal/receipt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type Options struct {
	ReceiptPath      string
	SchemaPath       string
	StrictHashes     bool
	StrictSignature  bool
	StrictApprovals  bool
	PublicKeyPathOpt string
	PublicKeyDirOpt  string

	StrictChain bool
	ChainDir    string
}

type Result struct {
	ReceiptPath string
	SchemaPath  string
	Hashes      receipt.HashCheck
	Signature   receipt.SignatureCheck
	Approvals   receipt.ApprovalSigCheck
	Chain       receipt.ChainCheck
}

func Run(opts Options) (*Result, error) {
	if opts.ReceiptPath == "" {
		return nil, errors.New("receipt path is required")
	}
	if opts.SchemaPath == "" {
		opts.SchemaPath = filepath.Join("spec", "receipt.schema.json")
	}

	if opts.StrictChain {
		opts.StrictHashes = true
		opts.StrictSignature = true
	}

	schema, err := compileSchema(opts.SchemaPath)
	if err != nil {
		return nil, err
	}

	r, err := receipt.Load(opts.ReceiptPath)
	if err != nil {
		return nil, err
	}

	hc, sc, ac, err := ValidateReceiptObject(r, schema, ReceiptValidationOptions{
		StrictHashes:    opts.StrictHashes,
		StrictSignature: opts.StrictSignature,
		StrictApprovals: opts.StrictApprovals,
		PublicKeyPath:   opts.PublicKeyPathOpt,
		PublicKeyDir:    opts.PublicKeyDirOpt,
	})
	if err != nil {
		return nil, err
	}

	cc := receipt.ChainCheck{Skipped: true}

	if opts.StrictChain {
		dir := opts.ChainDir
		if dir == "" {
			dir = filepath.Dir(opts.ReceiptPath)
		}
		resolver, err := receipt.NewDirResolver(dir)
		if err != nil {
			return nil, err
		}

		validateParent := func(pr receipt.Receipt) error {
			_, _, _, err := ValidateReceiptObject(pr, schema, ReceiptValidationOptions{
				StrictHashes:    true,
				StrictSignature: true,
				StrictApprovals: opts.StrictApprovals,
				PublicKeyPath:   opts.PublicKeyPathOpt,
				PublicKeyDir:    opts.PublicKeyDirOpt,
			})
			return err
		}

		c, err := receipt.ValidateChain(r, resolver, validateParent, receipt.ChainValidationOptions{Strict: true})
		if err != nil {
			return nil, err
		}
		cc = *c
	}

	return &Result{
		ReceiptPath: opts.ReceiptPath,
		SchemaPath:  opts.SchemaPath,
		Hashes:      hc,
		Signature:   sc,
		Approvals:   ac,
		Chain:       cc,
	}, nil
}

func compileSchema(schemaPath string) (*jsonschema.Schema, error) {
	abs, err := filepath.Abs(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("resolve schema path: %w", err)
	}

	f, err := os.Open(abs)
	if err != nil {
		return nil, fmt.Errorf("open schema: %w", err)
	}
	defer f.Close()

	c := jsonschema.NewCompiler()

	const schemaURL = "https://ix-agent-notary.local/spec/receipt.schema.json"
	if err := c.AddResource(schemaURL, f); err != nil {
		return nil, fmt.Errorf("add schema resource: %w", err)
	}

	s, err := c.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}

	return s, nil
}
