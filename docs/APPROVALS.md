# Approvals (v0)

Enterprises don’t just want “policy allowed it.” They want **governance evidence**:
- who approved,
- what exactly was approved,
- when it was approved,
- and (optionally) a signature from the approver identity.

IX-Agent-Notary models this as structured objects inside:
`policy.approvals[]`

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
  - `alg` (string) — e.g. `ed25519`
  - `key_id` (string)
  - `value` (string) — signature over a canonicalized approval payload (future: normative spec)

---

## Why approvals matter (real buyer value)

Approvals turn receipts into **auditable governance artifacts**:
- SOC2 / ISO27001 evidence
- Change-management linkage (ticket approvals)
- Break-glass logging (incident time access)
- Least-privilege + “two person rule” patterns (future extension)

---

## Demo support
`ix-an simulate` can optionally embed an approval record using `--approve`.

This is intentionally minimal, but the structure is designed so a real deployment can plug in:
- Jira/ServiceNow change approvals
- GitHub PR approvals
- Cloud IAM justifications / access requests
- KMS/HSM-backed approval signatures
