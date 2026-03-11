# Approvals (v0)

Enterprises don’t just want “policy allowed it.” They want governance evidence:
- who approved,
- what exactly was approved,
- when it was approved,
- and (optionally) a signature from the approver identity.

IX-Agent-Notary models this as structured objects inside: `policy.approvals[]`.

---

## Approval object (schema-backed)

Each approval is a JSON object with these required fields:

- `approval_id` (string) — unique ID for the approval record
- `type` (enum) — `human | ticket | breakglass`
- `status` (enum) — `requested | approved | denied | expired | revoked`
- `approver` (object)
  - `type` (string) — e.g. `user`, `service`, `group`
  - `id` (string) — stable identifier (email, IAM principal, etc.)
  - `display` (string, optional)
- `scope` (object)
  - `kind` (string) — e.g. `tool.invoke`
  - `tool` (string)
  - `operation` (string)
  - `resource` (string, optional) — e.g. path, URL, ARN, ticket ID, etc.
- `time` (object)
  - `requested_at` (date-time)
  - `decided_at` (date-time)
  - `expires_at` (date-time, optional)

Optional fields:
- `notes` (string)
- `signature` (object)
  - `alg` (string) — `ed25519`
  - `key_id` (string)
  - `value` (string) — signature over canonical approval payload (RFC8785), including `signature.alg` and `signature.key_id`, but excluding `signature.value`

---

## Approval signatures (implemented)

### Simulator emits signed approvals
Run:

```bash
go run ./cmd/ix-an simulate --path docs/approved.txt --out /tmp/approved.receipt.json \
  --approve --approver you@example.com --approval-type ticket

The simulator signs the approval object (demo uses the same key as receipt signing).

Verifier can enforce signed approvals

Run:
go run ./cmd/ix-an verify /tmp/approved.receipt.json --strict-approvals --strict-hashes --strict-signature

Strict approvals means:

if approvals exist, each approval must include a signature

each signature must verify

Why approvals matter (buyer value)

Approvals turn receipts into auditable governance artifacts:

SOC2 / ISO27001 evidence

change-management linkage (ticket approvals)

break-glass logging (incident-time access)

least-privilege + “two-person rule” patterns (future extension)
