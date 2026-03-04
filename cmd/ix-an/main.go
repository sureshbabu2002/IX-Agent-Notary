package main

import (
	"flag"
	"fmt"
	"os"

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

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "verify requires exactly 1 argument: <receipt.json>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Example:")
		fmt.Fprintln(os.Stderr, "  ix-an verify examples/receipts/minimal.receipt.json")
		os.Exit(2)
	}

	receiptPath := fs.Arg(0)

	_, err := verify.Run(verify.Options{
		ReceiptPath: receiptPath,
		SchemaPath:  *schemaPath,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("OK: receipt matches schema (signature verification not yet implemented)")
}

func usage() {
	fmt.Fprintln(os.Stderr, "IX-Agent-Notary (ix-an)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  ix-an <command> [options]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  verify   Validate a receipt against the JSON Schema")
	fmt.Fprintln(os.Stderr, "  help     Show this help")
	fmt.Fprintln(os.Stderr)
}
