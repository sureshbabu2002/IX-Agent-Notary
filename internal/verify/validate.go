package verify

import (
	"fmt"

	"ix-agent-notary/internal/receipt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type ReceiptValidationOptions struct {
	StrictHashes    bool
	StrictSignature bool
	StrictApprovals bool
	PublicKeyPath   string
	PublicKeyDir    string
}

func CompileSchema(schemaPath string) (*jsonschema.Schema, error) {
	return compileSchema(schemaPath)
}

func ValidateReceiptObject(
	r receipt.Receipt,
	schema *jsonschema.Schema,
	opts ReceiptValidationOptions,
) (receipt.HashCheck, receipt.SignatureCheck, receipt.ApprovalSigCheck, error) {
	if schema == nil {
		return receipt.HashCheck{}, receipt.SignatureCheck{}, receipt.ApprovalSigCheck{}, fmt.Errorf("schema is nil")
	}

	if err := schema.Validate(any(r)); err != nil {
		return receipt.HashCheck{}, receipt.SignatureCheck{}, receipt.ApprovalSigCheck{}, fmt.Errorf("schema validation failed: %w", err)
	}

	hc, err := receipt.ValidateCoreHashes(r, receipt.HashValidationOptions{Strict: opts.StrictHashes})
	if err != nil {
		return receipt.HashCheck{}, receipt.SignatureCheck{}, receipt.ApprovalSigCheck{}, err
	}

	sc, err := receipt.ValidateSignature(r, receipt.SignatureValidationOptions{
		Strict:        opts.StrictSignature,
		PublicKeyPath: opts.PublicKeyPath,
		PublicKeyDir:  opts.PublicKeyDir,
	})
	if err != nil {
		return receipt.HashCheck{}, receipt.SignatureCheck{}, receipt.ApprovalSigCheck{}, err
	}

	ac, err := receipt.ValidateApprovalSignatures(r, receipt.ApprovalSigValidationOptions{
		Strict:        opts.StrictApprovals,
		PublicKeyPath: opts.PublicKeyPath,
		PublicKeyDir:  opts.PublicKeyDir,
	})
	if err != nil {
		return receipt.HashCheck{}, receipt.SignatureCheck{}, receipt.ApprovalSigCheck{}, err
	}

	return *hc, *sc, *ac, nil
}
