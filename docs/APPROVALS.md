# Approvals (v0)

Enterprises don’t just want “policy allowed it.” They want **governance evidence**:
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
  - `value` (string) — signature over canonical approval payload (RFC8785), excluding `signature.value`

---

## Approval signatures (implemented)

### Simulator emits signed approvals
When you run:

```bash
ix-an simulate ... --approve
