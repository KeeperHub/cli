# Proposal: optional `sbo3l_*` policy-receipt envelope fields on workflow webhook submit

> **Audience:** KeeperHub CLI / platform / docs maintainers.
> **Status:** Draft proposal (not yet implemented). Opening this
> doc-only PR to start a design conversation.
> **Author:** SBO3L community contributor (B2JK-Industry).
> **Reference implementation:** [`sbo3l-keeperhub-adapter`](https://crates.io/crates/sbo3l-keeperhub-adapter)
> on crates.io (v1.2.0 live).

## Summary

Propose adding a small set of **optional** `sbo3l_*`-prefixed
fields to the workflow webhook submit envelope so an upstream
caller (an autonomous agent or a policy-engine middleware) can
attach a signed policy receipt + audit-anchor commitment to a
workflow execution at submit time.

Fields are **optional** and **prefix-namespaced** so KeeperHub's
core surface is unchanged. Consumers that don't read them are
unaffected; consumers that do read them get a verifiable trust
profile attached to every workflow execution they submit.

## The shape

When submitting a workflow webhook (`POST` to the configured
webhook URL), the body MAY include the following optional fields
alongside the existing standard payload:

```json
{
  "<existing standard payload>": "...",

  "sbo3l_policy_hash": "<64-hex JCS+SHA-256 of the agent's active policy>",
  "sbo3l_audit_root":  "<64-hex commitment to the agent's audit chain>",
  "sbo3l_pubkey_ed25519": "<64-hex public key>",
  "sbo3l_receipt": "<base64-encoded signed-receipt struct>",
  "sbo3l_capsule_uri": "<https URL to the signed Passport capsule, OR data: URI>",
  "sbo3l_evidence_chain_position": "<integer — audit-chain block height the receipt was emitted at>"
}
```

Each field is independently meaningful:

| Field | Type | Purpose |
|---|---|---|
| `sbo3l_policy_hash` | 64-char hex | Pin to the policy snapshot the agent was running when this submit decision was made. KeeperHub-side verifier can re-fetch the snapshot from `policy_url` and recompute. |
| `sbo3l_audit_root` | 64-char hex | Append-only digest of the agent's audit chain at submit time. Tamper-resistant — silent retroactive history change shifts the digest. |
| `sbo3l_pubkey_ed25519` | 64-char hex | Verifying key for the receipt signature. |
| `sbo3l_receipt` | base64 string | Ed25519 signed receipt over `(policy_hash, audit_root, decision="allow", capsule_id)`. Receipt is the load-bearing trust artefact. |
| `sbo3l_capsule_uri` | URL or `data:` URI | Pointer to the signed Passport capsule (~50KB) — the full pre-execution evidence bundle. URI form keeps the webhook payload small; `data:` URI form embeds inline for sponsors that prefer self-contained submits. |
| `sbo3l_evidence_chain_position` | uint | Audit-chain block height the decision was anchored at. Lets a verifier replay the decision off-chain. |

## Why opt-in / why prefix

- **Opt-in**: KeeperHub doesn't depend on any of these fields. A
  webhook handler that ignores them is unchanged. A webhook
  handler that reads them gets a tamper-resistant trust profile
  attached to the workflow execution.
- **Prefix-namespaced**: `sbo3l_*` so the convention generalises.
  An adapter from another agent platform (e.g. LangChain-side
  policy boundary) can adopt the same shape with its own prefix
  (`langchain_*`, `crewai_*`, ...) — the prefix itself is the
  vendor token.

If the proposal lands, KeeperHub's webhook handler doesn't need
to validate the receipt cryptographically. The fields are
**advisory** at the platform level and **load-bearing** at the
adapter / consumer level: a sponsor receiving the workflow
execution can choose to verify-then-honour, or refuse the
execution if verification fails.

## Why this matters

Today, KeeperHub workflow submissions are **opaque** at the trust
layer. The submit envelope says "agent X requests workflow Y" —
no shared way to attach a signed receipt, no audit-chain
commitment, no pre-execution evidence bundle. Sponsors integrating
with KeeperHub either:

1. Trust the agent unconditionally (single-shot rubber-stamp).
2. Maintain their own bespoke pre-flight envelope (every project
   diverges).
3. Refuse autonomous-agent submissions entirely (the safe-but-
   limiting posture).

A standardised optional envelope shape means agentic platforms
can ship KeeperHub adapters with one consistent trust surface.
The reference implementation already exists:
[`sbo3l-keeperhub-adapter`](https://crates.io/crates/sbo3l-keeperhub-adapter)
v1.2.0 attaches these fields when submitting workflows on behalf
of SBO3L agents, and the receiving sponsor can verify-or-ignore.

## Reference implementation

`sbo3l-keeperhub-adapter` (crates.io v1.2.0) attaches these fields
today using a vendor-specific convention. Switching to a
KeeperHub-canonicalised prefix ("kh-recommended") would let
non-SBO3L adapters adopt the same shape without re-deriving it.

Consumer-shape changes for the adapter once this proposal lands
are documented as five doc-only PRs in the SBO3L repo:

- [PR #402](https://github.com/B2JK-Industry/SBO3L-ethglobal-openagents-2026/pull/402) — `KhErrorCode` enum split (companion to KH issue [#52](https://github.com/KeeperHub/cli/issues/52))
- [PR #403](https://github.com/B2JK-Industry/SBO3L-ethglobal-openagents-2026/pull/403) — adapter tests against `keeperhub-mock` (companion to KH issue [#53](https://github.com/KeeperHub/cli/issues/53))
- [PR #404](https://github.com/B2JK-Industry/SBO3L-ethglobal-openagents-2026/pull/404) — timeout SLO calibration (companion to KH issue [#54](https://github.com/KeeperHub/cli/issues/54))
- [PR #405](https://github.com/B2JK-Industry/SBO3L-ethglobal-openagents-2026/pull/405) — schema versioning headers (companion to KH issue [#55](https://github.com/KeeperHub/cli/issues/55))
- [PR #406](https://github.com/B2JK-Industry/SBO3L-ethglobal-openagents-2026/pull/406) — payload size + 413 envelope (companion to KH issue [#56](https://github.com/KeeperHub/cli/issues/56))

These show the **consumer-side adapter changes** the SBO3L
adapter would land once the upstream KH spec stabilises. Each
issue + companion PR documents exactly one friction point hit
during integration; the PR shows what the adapter diff looks like
once KH lands the proposal.

## Discussion / open questions

- Should the prefix be `sbo3l_*` (project-specific, generalisable
  to per-vendor) or `kh-recommended-*` (KH-canonicalised,
  centralised)?
- Should the receipt signature scheme be specified (Ed25519 in
  SBO3L's case) or left implementation-specific?
- Should KeeperHub's webhook handler optionally validate the
  receipt + reject mismatched ones, or stay agnostic (let the
  sponsor decide)?
- Does this compose with the proposed Webhook Schema Versioning
  ([KH#55](https://github.com/KeeperHub/cli/issues/55)) — i.e.
  bump the schema version when this lands?

## Compatibility

- **Backwards compatible.** Optional fields, no breaking changes
  to existing webhook handlers.
- **Forward compatible.** A KH version that defines additional
  trust-envelope fields can ship them under a different prefix
  without colliding with `sbo3l_*`.

## Proposal

Open this doc-only PR to start the design conversation. If
maintainers ack the direction, follow-up implementation PRs would:

1. Add an "Envelope" section to the workflow webhook schema docs
   noting the optional `sbo3l_*` fields (or `kh-recommended-*`,
   per discussion outcome).
2. Update example webhook payloads in this repo to demonstrate
   the optional fields.
3. (Optional) Add a server-side verifier helper as a Go module
   that platform consumers can import.

The reference adapter is already in production and shipping these
fields; canonicalising the convention upstream means the rest of
the agentic-DeFi ecosystem can converge.

🤖 Community-contributed proposal by an SBO3L (B2JK-Industry)
contributor. Happy to iterate on any of the above.
