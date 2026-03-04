# Policy Pack Integrity (policy_hash)

Enterprises don’t just want “policy_id = X” in a receipt — they want to know **exactly which policy version** produced the decision.

IX-Agent-Notary supports this with:

- `policy.policy_hash` (recommended)
- `policy.policy_source` (optional metadata)

## What policy_hash means
`policy_hash` is a stable, verifiable content hash of the policy pack.

**Normative algorithm (this repo):**
1) Parse the policy JSON
2) Canonicalize using RFC8785 (JCS)
3) Compute SHA-256 digest
4) Encode digest as base64url (no padding)
5) Prefix with `sha256:`

That’s it. Same policy content → same hash, regardless of whitespace/key order.

## Why it matters
- Prevents “policy drift” ambiguity during incident response
- Makes approvals/audits defensible (“this is the exact policy we ran”)
- Enables storing receipts in append-only logs where the policy pack itself may live elsewhere (Git, S3, policy registry)

## Operational guidance (real deployments)
- Treat policy packs as versioned artifacts (Git commit hash, signed release, OCI artifact, etc.)
- Optionally embed additional attestations:
  - policy registry URI
  - policy signer identity
  - policy signature / transparency log reference

Those are future extensions; `policy_hash` is the minimal, high-value core.
