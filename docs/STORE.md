# Receipt Store (v0)

This repo supports a minimal “receipt store” pattern.

## 1) Directory store
Receipts can be stored as individual JSON files in a directory.

Run:
```bash
ix-an verify-dir <dir>
This enforces:

schema validity

strict core hashes

strict signature verification

(default) chain verification across receipts in the directory

2) Append-only log store (JSONL)

Receipts can be ingested into an append-only JSON Lines log.

Append:
ix-an store append --in <receipt.json> --log /path/to/receipts.jsonl

Verify the log:
ix-an store verify-log --log /path/to/receipts.jsonl

Notes:

The log is “append-only” by convention; IX-Agent-Notary verifies tamper evidence via signatures.

A real deployment should put the log behind immutability controls (WORM / S3 Object Lock / append-only DB).
