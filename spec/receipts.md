# IX-Agent-Notary Receipt Specification (Draft)

Status: **Draft** (pre-alpha)  
Current spec version: **0.1.0**

This document defines the **receipt** format emitted by IX-Agent-Notary. A receipt is a structured, tamper-evident record of a tool/action request, the **policy decision** that governed it, and the resulting effects—plus cryptographic integrity so third parties can verify it.

---

## 1) Goals

A receipt must allow an independent verifier to answer:

1. **What happened?** (action/tool, parameters summary, timing)
2. **Who requested it?** (agent identity)
3. **Who executed/mediated it?** (notary runtime identity)
4. **Was it allowed? Why?** (policy decision + evidence)
5. **What changed / what was produced?** (results + hashes)
6. **Can I trust this record?** (canonicalization + signature)
7. **Can I chain steps?** (trace + parent linkage)

---

## 2) Terminology

- **Agent**: the orchestration logic requesting actions. Treat as *untrusted / fallible*.
- **Notary runtime**: the enforcement boundary that evaluates policy and emits signed receipts.
- **Tool**: any invoked capability (CLI, API call, CI job, filesystem action, etc.).
- **Receipt**: a signed JSON document.
- **Verifier**: software that checks schema, hashes, chain linkage, and signatures.

---

## 3) Canonicalization & signing rules (normative)

Receipts are JSON objects that MUST be signed over a **canonical byte representation**.

### 3.1 Canonical JSON
- Canonicalization algorithm: **JCS (JSON Canonicalization Scheme, RFC 8785)**.
- The canonical JSON bytes are the input to hashing/signing.

### 3.2 Hashing
- Default hash: `SHA-256`
- Hash fields use lowercase hex or base64url (must be stated in `integrity.hash.encoding`).

### 3.3 Signature
Receipts include a signature envelope under `integrity.signature`.

- `integrity.signature.alg` MUST specify the algorithm, e.g.:
  - `ed25519`
  - `ecdsa-p256-sha256`
- `integrity.signature.value` is the signature over:
  - canonicalized JSON of the receipt **excluding** the `integrity.signature.value` field itself.

**Rule:** Verifiers MUST re-canonicalize and verify the signature, not trust stored bytes.

---

## 4) Receipt object (normative)

A receipt is a JSON object with these top-level fields:

- `receipt_version` (string, required) — semantic version of this receipt format (e.g., `"0.1.0"`).
- `receipt_id` (string, required) — globally unique identifier (UUID recommended).
- `time` (object, required) — timestamps for request/decision/execution.
- `trace` (object, required) — trace linkage for multi-step workflows.
- `actor` (object, required) — who requested the action (agent identity).
- `notary` (object, required) — who enforced policy/emitted receipt (runtime identity).
- `action` (object, required) — what was attempted.
- `policy` (object, required) — allow/deny decision + evidence.
- `result` (object, required) — outcome and artifacts/hashes.
- `integrity` (object, required) — canonicalization + hash + signature envelope.
- `redaction` (object, optional) — notes about removed/sensitive fields.

### 4.1 `time`
Required:
- `time.requested_at` — RFC3339 timestamp string
- `time.decided_at` — RFC3339 timestamp string
- `time.completed_at` — RFC3339 timestamp string (for denied actions, can equal `decided_at`)

### 4.2 `trace`
Required:
- `trace.trace_id` — stable ID for a workflow (UUID recommended)
- `trace.step` — integer step index (starts at 1)
Optional:
- `trace.parent_receipt_id` — receipt_id of the prior step (for chains)

### 4.3 `actor`
Required:
- `actor.type` — e.g. `"agent"`, `"service"`, `"user"`
- `actor.id` — stable identifier (string)
Optional:
- `actor.display` — human-friendly label
- `actor.session_id` — session identifier if applicable

### 4.4 `notary`
Required:
- `notary.runtime` — e.g. `"IX-Agent-Notary"`
- `notary.version` — runtime version string
- `notary.instance_id` — unique instance identifier
Optional:
- `notary.environment` — e.g. `"local"`, `"ci"`, `"prod-control-plane"`

### 4.5 `action`
Required:
- `action.kind` — e.g. `"tool.invoke"`, `"file.write"`, `"api.call"`
- `action.tool` — tool name (string) or `"N/A"` for non-tool actions
- `action.operation` — operation name (string), e.g. `"git.commit"`, `"github.create_issue"`
- `action.parameters` — object describing parameters **OR** a redacted placeholder
- `action.parameters_hash` — hash of canonicalized `action.parameters` (so you can redact but still prove stability)

Optional:
- `action.input_artifacts` — array of artifact refs/hashes used as inputs

**Rule:** If parameters contain secrets, store a redacted form in `action.parameters` and put the full detail only in a secure store; keep `parameters_hash` for integrity linkage.

### 4.6 `policy`
Required:
- `policy.policy_id` — identifier (string)
- `policy.decision` — `"allow"` or `"deny"`
- `policy.reason` — short human-readable reason
- `policy.rules` — array of matched rules with `rule_id`, `effect`, and `explanation`
- `policy.approvals` — array (can be empty). Each approval includes:
  - `approver_type` (`"user"|"service"`)
  - `approver_id`
  - `approved_at`
  - `scope` (what was approved)
Optional:
- `policy.context_hashes` — hashes of context inputs used for policy evaluation

### 4.7 `result`
Required:
- `result.status` — `"success"|"failure"|"denied"`
- `result.summary` — short human-readable summary
- `result.output` — object describing outputs (or redacted placeholder)
- `result.output_hash` — hash of canonicalized `result.output`

Optional:
- `result.artifacts` — array of `{ type, uri, hash, hash_alg }`
- `result.error` — structured error (for failures), with `code`, `message`, `details` (redactable)

### 4.8 `integrity`
Required:
- `integrity.canonicalization` — MUST be `"RFC8785-JCS"`
- `integrity.hash` — object:
  - `integrity.hash.alg` (e.g. `"sha-256"`)
  - `integrity.hash.encoding` (`"hex"` or `"base64url"`)
- `integrity.signature` — object:
  - `integrity.signature.alg` (e.g. `"ed25519"`)
  - `integrity.signature.key_id` — identifier for public key lookup
  - `integrity.signature.value` — signature string (base64url recommended)

Optional:
- `integrity.public_key` — embedded public key (discouraged for production, okay for demos)

---

## 5) Versioning rules (normative)
- Backward-compatible additions: bump MINOR (e.g., `0.1.0` → `0.2.0`)
- Breaking changes/removals: bump MAJOR (e.g., `0.x` may still break; once `1.0.0`, MAJOR indicates breaking)
- Receipts MUST include `receipt_version` so verifiers can enforce compatibility.

---

## 6) Minimal receipt example (informative)
See: `examples/receipts/minimal.receipt.json`

## 7) Denied receipt example (informative)
See: `examples/receipts/denied.receipt.json`

## 8) Notes
This spec intentionally separates:
- **policy evidence** (why it was allowed/denied)
- **action evidence** (what was attempted)
- **result evidence** (what happened)
- **cryptographic integrity** (why the record is believable)

Implementation will arrive in future commits (signing, verification CLI, policy evaluator, demos).
