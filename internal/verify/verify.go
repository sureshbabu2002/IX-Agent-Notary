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
	PublicKeyPathOpt string

	StrictChain bool
	ChainDir    string // if empty and StrictChain=true, defaults to directory containing ReceiptPath
}

type Result struct {
	ReceiptPath string
	SchemaPath  string
	Hashes      receipt.HashCheck
	Signature   receipt.SignatureCheck
	Chain       receipt.ChainCheck
}

func Run(opts Options) (*Result, error) {
	if opts.ReceiptPath == "" {
		return nil, errors.New("receipt path is required")
	}
	if opts.SchemaPath == "" {
		opts.SchemaPath = filepath.Join("spec", "receipt.schema.json")
	}

	// If we're doing strict chain validation, leaf must be strict too.
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

	if err := schema.Validate(any(r)); err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}

	hc, err := receipt.ValidateCoreHashes(r, receipt.HashValidationOptions{Strict: opts.StrictHashes})
	if err != nil {
		return nil, err
	}

	sc, err := receipt.ValidateSignature(r, receipt.SignatureValidationOptions{
		Strict:        opts.StrictSignature,
		PublicKeyPath: opts.PublicKeyPathOpt,
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

		// Validate each parent strictly: schema + hashes + signature.
		validateParent := func(pr receipt.Receipt) error {
			if err := schema.Validate(any(pr)); err != nil {
				return fmt.Errorf("schema validation failed: %w", err)
			}
			if _, err := receipt.ValidateCoreHashes(pr, receipt.HashValidationOptions{Strict: true}); err != nil {
				return err
			}
			if _, err := receipt.ValidateSignature(pr, receipt.SignatureValidationOptions{Strict: true, PublicKeyPath: opts.PublicKeyPathOpt}); err != nil {
				return err
			}
			return nil
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
		Hashes:      *hc,
		Signature:   *sc,
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
