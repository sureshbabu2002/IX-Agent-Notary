package verify

import (
	"fmt"

	"ix-agent-notary/internal/receipt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type ReceiptValidationOptions struct {
	StrictHashes    bool
	StrictSignature bool
	PublicKeyPath   string
}

// CompileSchema compiles the receipt JSON Schema once for reuse.
func CompileSchema(schemaPath string) (*jsonschema.Schema, error) {
	return compileSchema(schemaPath)
}

// ValidateReceiptObject validates a loaded receipt object (schema + optional strict hashes + optional strict signature).
func ValidateReceiptObject(r receipt.Receipt, schema *jsonschema.Schema, opts ReceiptValidationOptions) (receipt.HashCheck, receipt.SignatureCheck, error) {
	if schema == nil {
		return receipt.HashCheck{}, receipt.SignatureCheck{}, fmt.Errorf("schema is nil")
	}

	if err := schema.Validate(any(r)); err != nil {
		return receipt.HashCheck{}, receipt.SignatureCheck{}, fmt.Errorf("schema validation failed: %w", err)
	}

	hc, err := receipt.ValidateCoreHashes(r, receipt.HashValidationOptions{Strict: opts.StrictHashes})
	if err != nil {
		return receipt.HashCheck{}, receipt.SignatureCheck{}, err
	}

	sc, err := receipt.ValidateSignature(r, receipt.SignatureValidationOptions{
		Strict:        opts.StrictSignature,
		PublicKeyPath: opts.PublicKeyPath,
	})
	if err != nil {
		return receipt.HashCheck{}, receipt.SignatureCheck{}, err
	}

	return *hc, *sc, nil
}
