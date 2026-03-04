# IX-Agent-Notary — Architecture

## Problem statement
When an agent can call tools (code, CI/CD, cloud APIs, ticketing, secrets, production ops), the enterprise-grade question is:

**What exactly did the agent do, under what policy, with what approvals—and can we prove it?**

IX-Agent-Notary exists to make **policy enforcement + verifiable evidence** first-class.

---

## Design goals
1. **Tamper-evident receipts** for every meaningful action (inputs, policy decision, outputs, timing, identity, linkage).
2. **Independent verification**: any verifier can validate receipts without “trusting the agent.”
3. **PolicyGate enforcement**: tools don’t run unless allowed by policy (least privilege, allowlists).
4. **Small trusted computing base (TCB)**: keep the “trusted” runtime narrow and auditable.
5. **Correlation-friendly**: trace IDs and linkage across steps (receipt chains).

## Non-goals
- Building a full agent framework or chat UI.
- Claiming perfect prevention of all misuse (the goal is survivable integration with strong evidence).
- Replacing IAM; this complements IAM by producing **action-level** evidence.

---

## High-level components
- **Agent / Tool Caller (untrusted)**  
  Orchestrates tasks, requests tool usage, and consumes receipts.

- **IX-Agent-Notary Runtime (trusted boundary)**  
  Enforces policy, executes tools (or mediates execution), and emits signed receipts.

- **Tools / Systems (external)**  
  Git, CI jobs, cloud APIs, ticketing, etc.

- **Receipt Store (optional)**  
  Append-only store or log sink (filesystem, object storage, SIEM, etc.).

- **Verifier (trusted consumer)**  
  Validates signatures, schema, policy evidence, chain integrity, and red flags.

---

## Data flow (canonical)
```mermaid
flowchart LR
  subgraph AgentHost["Agent Host (untrusted logic)"]
    A["Agent / Tool Caller"]
  end

  subgraph Notary["IX-Agent-Notary (trusted runtime boundary)"]
    PG["PolicyGate (allow/deny + reason)"]
    EX["Tool Mediator/Executor"]
    RC["Receipt Composer"]
    SG["Signer"]
  end

  subgraph External["External Tools / Systems"]
    T["Tool/API Target"]
    ST["Receipt Store (optional)"]
    V["Verifier (CLI or service)"]
  end

  A -->|"Intent + context"| PG
  PG -->|"Allow/Deny (with evidence)"| EX
  EX -->|"Tool call"| T
  T -->|"Result"| EX
  EX --> RC
  RC --> SG
  SG -->|"Signed receipt"| ST
  ST --> V
  V -->|"Verify pass/fail + findings"| A
Trust boundaries (what must be trusted, what must not)
Trusted boundary (minimum viable TCB)

The following must be small, reviewable, and hardened:

Policy decision evaluator (PolicyGate)

Receipt construction + canonicalization (Receipt Composer)

Signing + key handling (Signer)

Receipt verification logic (Verifier)

Untrusted / assumed-compromisable

The agent orchestration logic itself

Prompting / LLM outputs

Any upstream “planner” code

Any tool response content (must be captured as evidence, not trusted as truth)

Principle: assume the agent is fallible or compromised; rely on enforcement + receipts, not “agent honesty.”

Receipt chain concept (why this matters)

Receipts should form a linked chain (like a log with parent pointers):

Each receipt references:

its parent receipt ID (if any),

a trace ID (shared across a workflow),

and hashes of relevant inputs/outputs.

This makes it difficult to “drop” or reorder steps without detection.

Key management (baseline posture)

Initial builds will support:

Dev keys for local testing (explicitly labeled, not production-safe)

Pluggable key sources for real deployments later (HSM/KMS integration path)

Rule: verification must be possible without secret material.

Deployment modes (what enterprises expect)

Local dev: agent + notary runtime + verifier on one machine.

CI/CD gate: notary runtime runs in pipeline; receipts shipped to artifacts / log store.

Control plane: notary runtime as a service that mediates tool calls and emits receipts to SIEM.

What makes this “enterprise-real”

Enforced policy before execution (not after-the-fact logs)

Signed evidence that survives disputes and audits

Verifier that a security team can run independently

Next docs (coming in future commits):

Receipt schema specification

Policy language + examples

Threat model + abuse cases

