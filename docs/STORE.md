# Receipt Store (v0)

IX-Agent-Notary supports a minimal receipt-store pattern for keeping evidence.

The goal is simple:

- receipts should be easy to write
- easy to verify
- hard to silently tamper with

## 1) Directory store

Store receipts as individual `.json` files in a directory tree.

Verify the directory:

```bash
go run ./cmd/ix-an verify-dir --strict-approvals examples/receipts
```

What `verify-dir` enforces:

- schema validity
- strict core hash verification
- strict receipt signature verification
- optional strict approval verification when `--strict-approvals` is supplied
- chain verification by default across receipts found in the directory

That makes a plain directory of JSON files usable as an evaluation-grade evidence store.

## 2) Append-only log store (JSONL)

Receipts can also be ingested into an append-only JSON Lines log.

Append a receipt to the log:

```bash
go run ./cmd/ix-an store append --in /tmp/approved.receipt.json --log /tmp/receipts.jsonl
```

Important behavior:

- `store append` performs strict schema, hash, and signature validation before ingest
- approval signatures are enforced only if you add `--strict-approvals`
- invalid receipts are rejected before they ever enter the log

Verify the entire log:

```bash
go run ./cmd/ix-an store verify-log --log /tmp/receipts.jsonl
```

`store verify-log` checks:

- every log line is valid JSON
- every log entry is a receipt object
- receipt IDs are unique inside the log
- each receipt verifies strictly
- parent linkage is checked by default across receipts already in the log
- approval signatures are enforced if `--strict-approvals` is supplied

## 3) What this store is and is not

The directory store and JSONL log are intentionally simple patterns, not a claim of full immutable-storage infrastructure.

The trust model is:

- storage may be modified
- verification should detect unauthorized modification
- stronger deployments should add immutability controls around storage itself

## 4) Production guidance

For real deployments, put receipt storage behind stronger controls such as:

- WORM object storage
- S3 Object Lock style retention
- append-only databases
- tightly controlled archival pipelines

The signed receipt is the evidence object.  
Immutable storage increases confidence that evidence was not deleted or reordered.

## 5) Practical evaluation path

For a clean local evaluation:

```bash
bash scripts/gen_demo_assets.sh
go run ./cmd/ix-an verify-dir --strict-approvals examples/receipts
```

For a log-oriented evaluation:

```bash
go run ./cmd/ix-an store append --in examples/receipts/approved.receipt.json --log /tmp/receipts.jsonl
go run ./cmd/ix-an store verify-log --log /tmp/receipts.jsonl
```
