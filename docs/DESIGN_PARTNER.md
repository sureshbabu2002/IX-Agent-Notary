# Design Partner (v0)

If you’re evaluating agent governance, this is the fastest way to turn IX-Agent-Notary into something deployable in your stack.

## What we want from a design partner
- A single real integration surface (one agent + one tool plane)
- A small allowlist policy requirement (what must be allowed vs denied)
- A receipt storage destination (directory, log pipeline, or SIEM ingest)
- 2–3 realistic incident/audit questions you need answered

## What you get
- A tailored proof-of-concept that emits verifiable receipts for your agent actions
- Policy + receipts that map to your audit story (SOC2/ISO-ish evidence)
- A clear “production requirements” list (IAM, KMS, immutability, monitoring)

## Engagement shape (practical)
- 1–2 weeks: POC scope + policy pack + receipt shape validation
- 2–4 weeks: enforce “no bypass” path + CI/SIEM ingestion + incident drill
- Outcome: deployable architecture + commercial licensing conversation

## How to start
Open a GitHub Issue titled:
**“Commercial licensing / design partner”**
and include:
- tool plane (what tools the agent touches)
- environment (CI, prod, sandbox)
- your must-have audit questions
