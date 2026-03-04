package main

import (
	"flag"
	"fmt"
	"os"

	"ix-agent-notary/internal/sign"
	"ix-agent-notary/internal/simulate"
	"ix-agent-notary/internal/verify"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "verify":
		verifyCmd(os.Args[2:])
	case "sign":
		signCmd(os.Args[2:])
	case "simulate":
		simulateCmd(os.Args[2:])
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func verifyCmd(args []string) {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	schemaPath := fs.String("schema", "", "path to receipt JSON Schema (default: spec/receipt.schema.json)")
	strictHashes := fs.Bool("strict-hashes", false, "fail if parameters_hash/output_hash are placeholders or missing")
	strictSig := fs.Bool("strict-signature", false, "fail if signature is missing/placeholder or public key can't be resolved")
	pubKeyPath := fs.String("pubkey", "", "optional path to an ed25519 public key (base64url). Overrides key lookup by key_id.")
	strictChain := fs.Bool("strict-chain", false, "verify parent_receipt_id chain (loads parent receipts from --chain-dir or receipt directory). Implies strict hashes+signature for the leaf.")
	chainDir := fs.String("chain-dir", "", "directory to search for parent receipts (default: directory containing the receipt)")

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "verify requires exactly 1 argument: <receipt.json>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Example:")
		fmt.Fprintln(os.Stderr, "  ix-an verify examples/receipts/denied.receipt.json --strict-chain")
		os.Exit(2)
	}

	receiptPath := fs.Arg(0)

	res, err := verify.Run(verify.Options{
		ReceiptPath:      receiptPath,
		SchemaPath:       *schemaPath,
		StrictHashes:     *strictHashes,
		StrictSignature:  *strictSig,
		PublicKeyPathOpt: *pubKeyPath,
		StrictChain:      *strictChain,
		ChainDir:         *chainDir,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	notes := []string{"schema ok"}

	if res.Hashes.Skipped {
		notes = append(notes, "hashes skipped")
	} else {
		notes = append(notes, "hashes ok")
	}

	if res.Signature.Skipped {
		notes = append(notes, "signature skipped")
	} else {
		notes = append(notes, "signature ok")
	}

	if res.Chain.Skipped {
		notes = append(notes, "chain skipped")
	} else {
		notes = append(notes, fmt.Sprintf("chain ok (depth=%d)", res.Chain.Depth))
	}

	fmt.Printf("OK: %s\n", joinNotes(notes))
}

func signCmd(args []string) {
	fs := flag.NewFlagSet("sign", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	inPath := fs.String("in", "", "input receipt JSON path")
	outPath := fs.String("out", "", "output receipt JSON path")
	keyPath := fs.String("key", "", "ed25519 private key seed path (32-byte seed base64url). Default: keys/dev/dev-key-001.seed")
	keyID := fs.String("key-id", "dev-key-001", "signature key_id to write into receipt (default: dev-key-001)")

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if *inPath == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "sign requires --in and --out")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Example:")
		fmt.Fprintln(os.Stderr, "  ix-an sign --in examples/receipts/minimal.receipt.json --out /tmp/minimal.signed.json --key keys/dev/dev-key-001.seed --key-id dev-key-001")
		os.Exit(2)
	}

	if err := sign.Run(sign.Options{
		InPath:  *inPath,
		OutPath: *outPath,
		KeyPath: *keyPath,
		KeyID:   *keyID,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("OK: wrote signed receipt:", *outPath)
}

func simulateCmd(args []string) {
	fs := flag.NewFlagSet("simulate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	policyPath := fs.String("policy", "policy/demo.policy.json", "path to policy JSON (default: policy/demo.policy.json)")
	outPath := fs.String("out", "", "output receipt JSON path")
	kind := fs.String("kind", "tool.invoke", "action kind (default: tool.invoke)")
	tool := fs.String("tool", "filesystem", "tool name (default: filesystem)")
	operation := fs.String("op", "file.write", "operation name (default: file.write)")
	path := fs.String("path", "", "target path for the simulated write (required)")
	bytes := fs.Int("bytes", 0, "byte count for the simulated write (optional)")
	actorID := fs.String("actor", "agent:demo", "actor id (default: agent:demo)")
	sessionID := fs.String("session", "sess-demo-001", "session id (default: sess-demo-001)")
	notaryInst := fs.String("notary", "notary-local-001", "notary instance id (default: notary-local-001)")
	keyPath := fs.String("key", "keys/dev/dev-key-001.seed", "ed25519 seed key path (default: keys/dev/dev-key-001.seed)")
	keyID := fs.String("key-id", "dev-key-001", "signature key_id to use (default: dev-key-001)")

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if *outPath == "" || *path == "" {
		fmt.Fprintln(os.Stderr, "simulate requires --out and --path")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  ix-an simulate --path docs/demo.txt --out /tmp/allow.receipt.json")
		fmt.Fprintln(os.Stderr, "  ix-an simulate --path .env        --out /tmp/deny.receipt.json")
		os.Exit(2)
	}

	if err := simulate.Run(simulate.Options{
		PolicyPath:  *policyPath,
		OutPath:     *outPath,
		Kind:        *kind,
		Tool:        *tool,
		Operation:   *operation,
		Path:        *path,
		Bytes:       *bytes,
		ActorID:     *actorID,
		SessionID:   *sessionID,
		NotaryInst:  *notaryInst,
		SignKeyPath: *keyPath,
		SignKeyID:   *keyID,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("OK: wrote simulated signed receipt:", *outPath)
}

func usage() {
	fmt.Fprintln(os.Stderr, "IX-Agent-Notary (ix-an)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  ix-an <command> [options]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  verify     Validate a receipt (schema + hashes + signature + optional chain)")
	fmt.Fprintln(os.Stderr, "  sign       Compute hashes + sign a receipt (ed25519)")
	fmt.Fprintln(os.Stderr, "  simulate   Simulate a tool action through PolicyGate and emit a signed receipt")
	fmt.Fprintln(os.Stderr, "  help       Show this help")
	fmt.Fprintln(os.Stderr)
}

func joinNotes(items []string) string {
	if len(items) == 0 {
		return ""
	}
	out := items[0]
	for i := 1; i < len(items); i++ {
		out += "; " + items[i]
	}
	return out
}
