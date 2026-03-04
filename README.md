# IX-Agent-Notary
Proof-carrying agent/tool actions: **policy enforcement + cryptographically signed receipts** you can verify in CI/SIEM.

## Why this exists
As soon as an AI agent can touch real systems (repos, CI/CD, cloud APIs, ticketing, secrets, production ops), the enterprise question becomes:

**What exactly did the agent do, under what policy, with what approvals — and can we prove it?**

IX-Agent-Notary is a small “trust layer” that makes that answer **machine-verifiable**.

## What it does (plain English)
1) A tool action is evaluated by **PolicyGate** (allow/deny, least privilege).  
2) The notary emits a **receipt**: who/what/when, the exact policy decision, hashes, and optional approvals.  
3) The receipt is canonicalized (RFC 8785 / JCS) and **signed** (ed25519 in v0).  
4) A verifier independently checks: schema, hashes, signatures, approvals, and optional chain linkage.

This is not “trust me bro” logging — it’s evidence you can verify.

## 10-minute evaluation (recommended)
Run the repo CI script locally (this is also what GitHub Actions runs):

```bash
bash scripts/ci.sh
What that does:

enforces gofmt, go vet, go test

generates local dev keys + demo receipts (gitignored)

strictly verifies the generated receipts directory

Generate demo assets only
bash scripts/gen_demo_assets.sh
go run ./cmd/ix-an verify-dir examples/receipts --strict-approvals

Note: This repo intentionally ships no private keys and no pre-generated receipts.
Demo keys/receipts are generated locally and are gitignored by design.

Core capabilities (v0)

Receipt schema (spec/receipt.schema.json) and draft spec (spec/receipts.md)

PolicyGate (demo policy pack pattern under policy/)

Signing (ed25519; canonical JSON via RFC 8785 / JCS)

Strict verification (schema + hashes + signature + approvals + optional chain)

Approvals as governance evidence (docs/APPROVALS.md)

Receipt storage patterns: directory store + append-only JSONL log (docs/STORE.md)

CLI (ix-an)

All commands run via Go (no install required):

Verify one receipt

go run ./cmd/ix-an verify <receipt.json> --strict-hashes --strict-signature

Verify a directory (strict by default)
go run ./cmd/ix-an verify-dir <dir> --strict-approvals

Simulate a tool action → emit a signed receipt
go run ./cmd/ix-an simulate --path docs/demo.txt --out /tmp/allow.receipt.json
go run ./cmd/ix-an verify /tmp/allow.receipt.json --strict-hashes --strict-signature

Approvals demo (signed governance evidence in the receipt)
go run ./cmd/ix-an simulate --path docs/approved.txt --out /tmp/approved.receipt.json \
  --approve --approver you@example.com --approval-type ticket

go run ./cmd/ix-an verify /tmp/approved.receipt.json --strict-hashes --strict-signature --strict-approvals

Append-only JSONL log (ingest strictly, then verify log)
go run ./cmd/ix-an store append --in /tmp/approved.receipt.json --log /tmp/receipts.jsonl
go run ./cmd/ix-an store verify-log --log /tmp/receipts.jsonl

Where this fits (buyer mental model)

IX-Agent-Notary is the enforcement + evidence layer that sits between “agents” and “tools”:

prevents unsafe calls via allow/deny policy

produces verifiable receipts for audit/compliance and incident response

makes agent integrations survivable in regulated environments

It is intentionally not a full agent framework and not a SIEM. It’s the part you want to trust.

Docs (start here)

Architecture: docs/ARCHITECTURE.md

Threat model: docs/THREAT_MODEL.md

Key management: docs/KEY_MANAGEMENT.md

Approvals: docs/APPROVALS.md

Receipt store: docs/STORE.md

Policy integrity notes: docs/POLICY_INTEGRITY.md

Enterprise pilot guide: docs/ENTERPRISE_PILOT.md

Design partner notes: docs/DESIGN_PARTNER.md

License / commercial use

IX-Agent-Notary is source-available for evaluation under LICENSE.

If you want to use this in production or any commercial context, you need a separate commercial license.
See COMMERCIAL.md for the exact trigger conditions and contact path.

Security

Please report security issues per SECURITY.md.

If you’re evaluating agent governance and want receipts that your security/compliance teams can actually verify, start with docs/ENTERPRISE_PILOT.md.
