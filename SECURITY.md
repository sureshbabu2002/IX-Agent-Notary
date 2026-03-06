# Security Policy

## Supported status

This project is currently **pre-1.0** and should be treated as **prototype / evaluation software** unless separately hardened and licensed for production use.

## Do not report security issues in public

Please do **not** open a normal public bug report for a security-sensitive issue.

Use one of these paths instead.

## Preferred reporting path

### 1) GitHub Private Vulnerability Reporting / Security Advisory

If private vulnerability reporting is enabled for this repository, use that first.

Include:

- affected component or file
- issue description
- impact
- reproduction steps
- conditions required to trigger it
- suggested fix, if you have one

## Fallback path when private reporting is unavailable

Open the issue template:

- `.github/ISSUE_TEMPLATE/security-private-channel.md`

Use the title:

- `Security (private channel requested)`

Keep the public issue minimal. Do **not** post exploit details publicly.

Include only:

- affected area at a high level
- severity estimate
- request for a private channel
- your preferred contact method

## What not to post publicly

Do not post any of the following in a public issue:

- exploit payloads
- private keys
- secret material
- customer or internal environment details
- step-by-step weaponized reproduction details

## What helps triage quickly

A high-quality private report usually includes:

- precise affected path or component
- attack preconditions
- realistic impact
- whether integrity, policy enforcement, receipt validity, or key handling is affected
- whether the issue is local-only or remotely triggerable
- whether a mitigation already exists

## Scope areas of particular interest

Security-sensitive areas in this repo include:

- policy enforcement boundary
- receipt canonicalization
- signature creation and verification
- approval-signature verification
- chain validation
- key resolution and trust anchoring
- storage integrity assumptions

## Coordination expectation

Please allow time for confirmation and remediation before public disclosure.

If you need to discuss coordinated disclosure timing, say so in the initial private report.

## Recommended starting documents

Before reporting, it may help to review:

- `docs/THREAT_MODEL.md`
- `docs/KEY_MANAGEMENT.md`
- `docs/POLICY_INTEGRITY.md`
- `docs/ARCHITECTURE.md`
