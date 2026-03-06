# Dev keys (local evaluation)

This directory exists so evaluators have a standard place for **locally generated** dev keys.

Important rules:

- **No private keys are committed to this repo.**
- `*.seed` and `*.pub` under `keys/dev/` are gitignored by design.
- Anything created under `keys/dev/` is **demo-only** and should not be treated as production signing material.

## Recommended path

Generate local dev keys and local demo receipts:

```bash
bash scripts/gen_demo_assets.sh
```

That script will:

- generate a local ed25519 keypair
- generate example receipts under `examples/receipts/`
- strictly verify the generated results

Typical local outputs:

- `keys/dev/dev-key-001.seed`
- `keys/dev/dev-key-001.pub`
- `examples/receipts/*.json`

## Manual key generation

If you only want keys:

```bash
go run ./cmd/ix-an keygen --out-seed keys/dev/dev-key-001.seed --out-pub keys/dev/dev-key-001.pub
```

## Manual verification example

```bash
go run ./cmd/ix-an verify \
  --strict-hashes \
  --strict-signature \
  --pubkey keys/dev/dev-key-001.pub \
  /tmp/allow.receipt.json
```

## Reminder

Do not publish, reuse, or operationalize local demo keys as if they were production trust anchors.
