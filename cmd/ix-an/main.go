package main

import (
	"flag"
	"fmt"
	"os"

	"ix-agent-notary/internal/receipt"
	"ix-agent-notary/internal/sign"
	"ix-agent-notary/internal/simulate"
	"ix-agent-notary/internal/store"
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
	case "verify-dir":
		verifyDirCmd(os.Args[2:])
	case "sign":
		signCmd(os.Args[2:])
	case "simulate":
		simulateCmd(os.Args[2:])
	case "store":
		storeCmd(os.Args[2:])
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

func verifyDirCmd(args []string) {
	fs := flag.NewFlagSet("verify-dir", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	schemaPath := fs.String("schema", "", "path to receipt JSON Schema (default: spec/receipt.schema.json)")
	pubKeyPath := fs.String("pubkey", "", "optional path to an ed25519 public key (base64url). Overrides key lookup by key_id.")
	strictChain := fs.Bool("strict-chain", true, "verify parent_receipt_id linkage for all receipts found (default: true)")

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "verify-dir requires exactly 1 argument: <dir>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Example:")
		fmt.Fprintln(os.Stderr, "  ix-an verify-dir examples/receipts")
		os.Exit(2)
	}

	dir := fs.Arg(0)

	res, err := verify.VerifyDir(verify.DirOptions{
		Dir:             dir,
		SchemaPath:      *schemaPath,
		PublicKeyPath:   *pubKeyPath,
		StrictHashes:    true,
		StrictSignature: true,
		StrictChain:     *strictChain,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		for _, f := range res.Failures {
			fmt.Fprintf(os.Stderr, "  - %s\n", f)
		}
		os.Exit(1)
	}

	fmt.Printf("OK: verified %d receipts (%d ok)\n", res.Total, res.OK)
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

	approve := fs.Bool("approve", false, "embed a single approval record in policy.approvals[] (demo governance evidence)")
	approver := fs.String("approver", "user:approver-demo", "approver id used when --approve is set")
	approvalType := fs.String("approval-type", "human", "approval type used when --approve is set: human|ticket|breakglass")

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if *outPath == "" || *path == "" {
		fmt.Fprintln(os.Stderr, "simulate requires --out and --path")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  ix-an simulate --path docs/demo.txt --out /tmp/allow.receipt.json")
		fmt.Fprintln(os.Stderr, "  ix-an simulate --path .env        --out /tmp/deny.receipt.json")
		fmt.Fprintln(os.Stderr, "  ix-an simulate --path docs/demo.txt --out /tmp/approved.receipt.json --approve --approver you@example.com --approval-type ticket")
		os.Exit(2)
	}

	if err := simulate.Run(simulate.Options{
		PolicyPath:       *policyPath,
		OutPath:          *outPath,
		Kind:             *kind,
		Tool:             *tool,
		Operation:        *operation,
		Path:             *path,
		Bytes:            *bytes,
		ActorID:          *actorID,
		SessionID:        *sessionID,
		NotaryInst:       *notaryInst,
		SignKeyPath:      *keyPath,
		SignKeyID:        *keyID,
		IncludeApproval:  *approve,
		ApproverID:       *approver,
		ApprovalType:     *approvalType,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("OK: wrote simulated signed receipt:", *outPath)
}

func storeCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "store requires a subcommand: append | verify-log")
		os.Exit(2)
	}

	switch args[0] {
	case "append":
		storeAppendCmd(args[1:])
	case "verify-log":
		storeVerifyLogCmd(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown store subcommand: %s\n", args[0])
		os.Exit(2)
	}
}

func storeAppendCmd(args []string) {
	fs := flag.NewFlagSet("store append", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	inPath := fs.String("in", "", "input receipt JSON path (required)")
	logPath := fs.String("log", "", "append-only JSONL log path (required)")
	schemaPath := fs.String("schema", "", "path to receipt JSON Schema (default: spec/receipt.schema.json)")
	pubKeyPath := fs.String("pubkey", "", "optional path to an ed25519 public key (base64url). Overrides key lookup by key_id.")

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if *inPath == "" || *logPath == "" {
		fmt.Fprintln(os.Stderr, "store append requires --in and --log")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Example:")
		fmt.Fprintln(os.Stderr, "  ix-an store append --in examples/receipts/denied.receipt.json --log /tmp/receipts.jsonl")
		os.Exit(2)
	}

	// Strictly verify before ingest.
	if _, err := verify.Run(verify.Options{
		ReceiptPath:      *inPath,
		SchemaPath:       *schemaPath,
		StrictHashes:     true,
		StrictSignature:  true,
		PublicKeyPathOpt: *pubKeyPath,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: receipt did not verify strictly; not ingesting: %v\n", err)
		os.Exit(1)
	}

	r, err := receipt.Load(*inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	if err := store.AppendJSONL(*logPath, r); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("OK: appended to log:", *logPath)
}

func storeVerifyLogCmd(args []string) {
	fs := flag.NewFlagSet("store verify-log", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	logPath := fs.String("log", "", "append-only JSONL log path (required)")
	schemaPath := fs.String("schema", "", "path to receipt JSON Schema (default: spec/receipt.schema.json)")
	pubKeyPath := fs.String("pubkey", "", "optional path to an ed25519 public key (base64url). Overrides key lookup by key_id.")
	strictChain := fs.Bool("strict-chain", true, "verify parent_receipt_id linkage within the log (default: true)")

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if *logPath == "" {
		fmt.Fprintln(os.Stderr, "store verify-log requires --log")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Example:")
		fmt.Fprintln(os.Stderr, "  ix-an store verify-log --log /tmp/receipts.jsonl")
		os.Exit(2)
	}

	if *schemaPath == "" {
		*schemaPath = "spec/receipt.schema.json"
	}

	schema, err := verify.CompileSchema(*schemaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	recs, err := store.ReadAllJSONL(*logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	byID := map[string]receipt.Receipt{}
	for i, r := range recs {
		rid, _ := r["receipt_id"].(string)
		if rid == "" {
			fmt.Fprintf(os.Stderr, "FAIL: log entry %d missing receipt_id\n", i+1)
			os.Exit(1)
		}
		if _, exists := byID[rid]; exists {
			fmt.Fprintf(os.Stderr, "FAIL: duplicate receipt_id in log: %s\n", rid)
			os.Exit(1)
		}
		byID[rid] = r
	}

	// Strictly validate all receipts.
	for rid, r := range byID {
		_, _, verr := verify.ValidateReceiptObject(r, schema, verify.ReceiptValidationOptions{
			StrictHashes:    true,
			StrictSignature: true,
			PublicKeyPath:   *pubKeyPath,
		})
		if verr != nil {
			fmt.Fprintf(os.Stderr, "FAIL: receipt %s failed verify: %v\n", rid, verr)
			os.Exit(1)
		}
	}

	if *strictChain {
		resolver, err := receipt.NewMapResolver(byID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
			os.Exit(1)
		}

		validateParent := func(r receipt.Receipt) error {
			_, _, verr := verify.ValidateReceiptObject(r, schema, verify.ReceiptValidationOptions{
				StrictHashes:    true,
				StrictSignature: true,
				PublicKeyPath:   *pubKeyPath,
			})
			return verr
		}

		for rid, r := range byID {
			_, cerr := receipt.ValidateChain(r, resolver, validateParent, receipt.ChainValidationOptions{Strict: true})
			if cerr != nil {
				fmt.Fprintf(os.Stderr, "FAIL: chain verify failed for receipt %s: %v\n", rid, cerr)
				os.Exit(1)
			}
		}
	}

	fmt.Printf("OK: verified log (%d receipts)\n", len(byID))
}

func usage() {
	fmt.Fprintln(os.Stderr, "IX-Agent-Notary (ix-an)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  ix-an <command> [options]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  verify        Validate a receipt (schema + hashes + signature + optional chain)")
	fmt.Fprintln(os.Stderr, "  verify-dir    Validate all receipts in a directory (strict by default)")
	fmt.Fprintln(os.Stderr, "  sign          Compute hashes + sign a receipt (ed25519)")
	fmt.Fprintln(os.Stderr, "  simulate      Simulate a tool action through PolicyGate and emit a signed receipt")
	fmt.Fprintln(os.Stderr, "  store         Append receipts to an append-only JSONL log and verify logs")
	fmt.Fprintln(os.Stderr, "  help          Show this help")
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
