# Policies

This folder contains **policy packs** for the PolicyGate evaluator.

- `demo.policy.json` is a deliberately small allowlist policy:
  - Denies writes to `.env`
  - Allows writes only under `docs/`
  - Default effect: **deny**

## Policy pack integrity (policy_hash)
When a receipt is emitted through PolicyGate, it can include:

- `policy.policy_hash` — a stable content hash of the policy pack (RFC8785 canonical JSON → sha-256 → base64url)
- `policy.policy_source` — where the policy was loaded from (file path in this demo)

This lets a verifier prove the decision came from a **specific policy version**, not just a policy ID string.
