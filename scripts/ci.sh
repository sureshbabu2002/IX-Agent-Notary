#!/usr/bin/env bash
set -euo pipefail

echo "== gofmt check =="
unformatted="$(gofmt -l .)"
if [[ -n "${unformatted}" ]]; then
  echo "gofmt required on:"
  echo "${unformatted}"
  exit 1
fi

echo "== go vet =="
go vet ./...

echo "== go test =="
go test ./...

echo "== verify examples (strict) =="
go run ./cmd/ix-an verify examples/receipts/minimal.receipt.json --strict-hashes --strict-signature
go run ./cmd/ix-an verify examples/receipts/denied.receipt.json  --strict-chain

echo "== verify examples directory (strict-chain default on) =="
go run ./cmd/ix-an verify-dir examples/receipts

echo "== simulate approvals + strict verify =="
tmpdir="$(mktemp -d)"
go run ./cmd/ix-an simulate --path docs/demo.txt --out "$tmpdir/approved.receipt.json" \
  --approve --approver you@example.com --approval-type ticket
go run ./cmd/ix-an verify "$tmpdir/approved.receipt.json" --strict-hashes --strict-signature --strict-approvals
