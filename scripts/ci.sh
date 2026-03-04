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
go run ./cmd/ix-an verify examples/receipts/denied.receipt.json  --strict-hashes --strict-signature --strict-chain
