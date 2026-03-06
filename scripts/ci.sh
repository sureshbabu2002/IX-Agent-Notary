#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

echo "== gofmt check =="
unformatted="$(gofmt -l .)"
if [[ -n "${unformatted}" ]]; then
  echo "gofmt required on:"
  echo "${unformatted}"
  exit 1
fi

echo "== go vet =="
go vet ./...

echo "== go build =="
go build ./cmd/ix-an

echo "== go test =="
go test ./...

echo "== generate demo assets (keys + receipts) =="
bash scripts/gen_demo_assets.sh

echo "== strict verify generated receipts individually =="
go run ./cmd/ix-an verify --strict-hashes --strict-signature examples/receipts/minimal.receipt.json
go run ./cmd/ix-an verify --strict-hashes --strict-signature examples/receipts/denied.receipt.json
go run ./cmd/ix-an verify --strict-hashes --strict-signature --strict-approvals examples/receipts/approved.receipt.json
go run ./cmd/ix-an verify --strict-hashes --strict-signature --strict-chain examples/receipts/chain.child.receipt.json

echo "== strict verify generated examples directory =="
go run ./cmd/ix-an verify-dir --strict-approvals --strict-chain examples/receipts

echo "OK: clean-clone CI pipeline passed"
