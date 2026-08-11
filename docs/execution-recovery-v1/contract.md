# Execution recovery contract (normative)

Version: **1.1.0**  
Audience: KeeperHub CLI / MCP / HTTP adapter authors  
Published path: this file is synced to docs.keeperhub.com via `docs/execution-recovery.md`.

## Definitions

- **Write**: any API call that may create or re-drive an onchain side effect (execute transfer, contract call, workflow webhook).
- **Status read**: polling `GET /api/execute/{id}/status` (direct execution) or workflow run status (separate vocabulary).
- **Chain evidence**: at least one receipt with `verified=true` and `receiptStatus=success` for the expected predicate.
- **Idempotency key**: client-supplied `Idempotency-Key` header that must be byte-identical across retries of the same logical write.

## Status vocabularies (do not mix)

| Surface | Pending | Terminal |
| --- | --- | --- |
| Direct execution (`/api/execute/.../status`) | `pending`, `running`, `queued`, `unconfirmed`, transport `not_found` | `completed`, `failed` |
| Workflow run (`/api/workflows/executions/.../status`) | `pending`, `running` | `success`, `error`, `cancelled` |

`completed` is **not** proof of onchain success. Inspect `receipts[].receiptStatus`.

## Rules

### R1 — Unconfirmed → poll, do not resubmit

If status is `pending`, `running`, `queued`, or `unconfirmed`, continue status reads against the **same** execution ID. Do not issue a new write for the same logical intent while that execution ID remains durable.

**CLI conformance:** `kh ex transfer --wait` / `kh ex cc --wait` poll the same ID.

### R2 — Chain evidence / reverted receipts

1. If any receipt has `receiptStatus=reverted`, the outcome is **Failure** even when `status=completed` and `verified=true`.
2. When the caller requires payment/landing proof (`RequireChainEvidence` / `--require-verified`), `completed` without a verified successful receipt is **Failure**.

**CLI conformance:** wait paths fail closed on reverted receipts. Strict “no receipt ⇒ fail” is opt-in (see also PR discussion for `--require-verified`).

### R3 — Write retry → stable idempotency key

If a write is retried after transport failure (timeout, 5xx) before an execution ID is known, the client MUST reuse the same `Idempotency-Key`. After an execution ID is known, prefer R1.

**CLI conformance:** `kh ex transfer` and `kh ex cc` (writes) set `Idempotency-Key` once per invocation before `Do`, so HTTP-layer retries reuse it. Use `--idempotency-key` to pin a key across process restarts.

### R4 — Terminal failure / malformed

Statuses `failed` (direct) and unparseable/malformed bodies are terminal for that attempt. Missing `status` after a 200 decode is **malformed**, not success. Do not invent a success path from partial fields.

### R5 — Rate limit

HTTP `429` responses require backoff. They are not success. Preserve the idempotency key for the next write attempt of the same intent. The HTTP client does not auto-retry 429.

### R6 — Cold start `not_found`

A first status read that returns HTTP 404 / `not_found` immediately after submit may be transient. During `--wait`, poll briefly before treating not_found as terminal timeout.

## Fixture mapping + conformance

| Fixture | Rule | Expect | Consumed by |
| --- | --- | --- | --- |
| `queued.json` | R1 | pending | `TestFixtures_ClassifyTable` |
| `unconfirmed.json` | R1 | pending | `TestFixtures_ClassifyTable` |
| `completed_with_tx.json` | R2 | success (strict) | `TestFixtures_ClassifyTable` |
| `completed_without_tx.json` | R2 | failure (strict) | `TestFixtures_ClassifyTable` |
| `reverted.json` | R2 | failure | `TestFixtures_ClassifyTable` + `TestRevertedIsNeverSuccess` |
| `failed.json` | R4 | failure | `TestFixtures_ClassifyTable` |
| `malformed.json` | R4 | malformed | `TestFixtures_ClassifyTable` |
| `not_found.json` | R6 | pending | `TestFixtures_ClassifyTable` |
| `rate_limited.json` | R5 | rate_limited | `TestFixtures_ClassifyTable` |
| `cold_start.sequence.json` | R6 | 404→pending→success | `TestColdStartSequence_R6` |

Fixtures use the **flat** direct-execution wire shape (`executionId`, not nested `execution.id`) so `json.Unmarshal` into `DirectStatus` / `ExecStatusResponse` cannot silently zero-decode.
