# Enterprise Pilot Guide (v0)

This is the shortest “serious” path to evaluate IX-Agent-Notary in a real environment.

---

## Goal
Prove that agent/tool execution can be:
- **policy-enforced** (least privilege)
- **auditable** (machine-verifiable receipts)
- **tamper-evident** (hashes + signatures)
- **governance-aware** (approvals evidence)
- **incident-friendly** (trace + chain linkage)

---

## Recommended pilot architecture

### Minimal
Agent → Notary wrapper → Tool(s)  
Notary emits receipts → append-only store → CI/SIEM verifies receipts

### Better
Agent → Intent Router / Tool Gateway → Notary enforcement+signing → Tool(s)  
Receipts → WORM/immutable store + SIEM ingest (verify on ingest)

---

## Step-by-step pilot

### 1) Define a narrow allowlist policy
Start with “deny by default” plus a tiny allowlist.

Use `policy/demo.policy.json` as a pattern:
- allow only a safe prefix (like `docs/`)
- explicitly deny sensitive targets (`.env`, secrets paths, IAM changes)

### 2) Require receipts for all tool execution
The **hard requirement** for meaningful evaluation is architectural:
> Tools must not be reachable unless the call passes through Notary.

### 3) Run strict verification in CI
Gate merges on verification:
- schema validity
- strict hashes
- strict signature
- strict chain verification (when parent links exist)

This repo already runs strict checks in:
- `.github/workflows/ci.yml`
- `scripts/ci.sh`

### 4) Store receipts immutably (directory or JSONL)
- Directory mode:
  - `ix-an verify-dir <dir>`
- JSONL log mode:
  - `ix-an store append --in <receipt.json> --log receipts.jsonl`
  - `ix-an store verify-log --log receipts.jsonl`

In production: put logs behind immutability controls (WORM / Object Lock / append-only DB).

### 5) Validate “policy integrity”
Require `policy.policy_hash` in receipts so the decision is tied to an exact policy version.

### 6) Add approvals for high-risk actions
For actions outside the normal allowlist:
- require a ticket approval
- embed the structured approval object in `policy.approvals[]`

This repo supports demo approvals via:
- `ix-an simulate ... --approve --approver you@example.com --approval-type ticket`

---

## Success criteria (what a buyer cares about)
- You can **reconstruct** what happened from receipts alone
- Receipts fail verification if tampered
- Policies are enforceable and version-provable (`policy_hash`)
- Approvals create defensible governance evidence
- The system can be made “mandatory path” (no bypass)

---

## Next enterprise-grade extensions (roadmap)
- KMS/HSM signing + key rotation tooling
- Approval signatures (multi-party / quorum)
- Transparency log / immutable registry integration
- Tool adapters (HTTP, GitHub, cloud APIs) with scoped tokens
