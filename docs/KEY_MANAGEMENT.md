# Key Management (v0)

Receipts are only as trustworthy as the signing keys behind them.

This repo intentionally **does not** ship any private signing material. Private keys must never be committed to a public repository.

---

## Local evaluation (no secrets committed)

Generate a local dev keypair and example receipts (all **gitignored**):

```bash
bash scripts/gen_demo_assets.sh

This creates:

keys/dev/dev-key-001.seed — private ed25519 seed (0600, gitignored)

keys/dev/dev-key-001.pub — public key (gitignored)

examples/receipts/*.json — generated receipts (gitignored)

Then verify strictly:

go run ./cmd/ix-an verify-dir examples/receipts --strict-approvals

If you prefer manual steps:
go run ./cmd/ix-an keygen --out-seed keys/dev/dev-key-001.seed --out-pub keys/dev/dev-key-001.pub
go run ./cmd/ix-an simulate --path docs/demo.txt --out /tmp/allow.json --key keys/dev/dev-key-001.seed --key-id dev-key-001
go run ./cmd/ix-an verify /tmp/allow.json --strict-hashes --strict-signature

Production guidance (baseline posture)
1) Store signing keys in KMS/HSM

Keep private key material hardware-backed when possible (HSM / KMS / Vault with HSM-backed keys).

Limit permissions to “sign receipt” operations only.

Gate signing behind IAM authorization and change control in high-risk environments.

2) Publish a trusted public-key allowlist

Verification should only accept signatures from:

a curated set of trusted public keys,

mapped to known key_id values,

with an explicit revocation story (even if “manual list update” in v0).

3) Rotate keys without breaking verification

Receipts include:

integrity.signature.key_id

Recommended pattern:

treat key_id as immutable for a specific key version (e.g., notary-prod-2026-03)

rotate by issuing a new key and a new key_id

keep historical public keys available so old receipts remain verifiable

4) Separate domains (optional, but stronger)

For higher assurance:

use a distinct key for the notary’s receipt signing

and separate keys for human/ticket approvals (different trust domain)

Threats this mitigates

Receipt tampering (signature fails)

Receipt fabrication (unknown key_id / untrusted public key)

“audit theater” placeholders (strict verifier rejects)

Silent evidence drift (hash + signature binds the canonical receipt payload)

See also:

docs/THREAT_MODEL.md

docs/POLICY_INTEGRITY.md
