# Execution recovery contract (normative)

Version: **1.4.0**  
Audience: KeeperHub CLI / MCP / HTTP adapter authors  
Published path: this file is synced to docs.keeperhub.com via `docs/execution-recovery.md`.

This contract covers **direct execution**:

- write: `POST /api/execute/transfer`, `POST /api/execute/contract-call`
- status: `GET /api/execute/{id}/status`

It does **not** cover `POST /api/workflows/<id>/webhook`.

## Definitions

- **Write**: `POST /api/execute/transfer` or `POST /api/execute/contract-call`.
- **Status read**: polling `GET /api/execute/{id}/status`.
- **Successful receipt**: `receipts[]` entry with `receiptStatus=success`.
- **Idempotency key**: client-supplied `Idempotency-Key` header that must be byte-identical across retries of the same logical write.

## Status vocabularies (do not mix)

"Pending" and "Terminal" below are **client wait semantics**: keep polling, or stop waiting and report. Terminal is not a claim that the server will never change the row again.

| Surface | Pending | Terminal |
| --- | --- | --- |
| Direct execution (`GET /api/execute/{id}/status`) | `pending`, `running` | `unconfirmed`, `completed`, `failed` |
| Workflow run (`GET /api/workflows/executions/{id}/status`) | `pending`, `running` | `success`, `error`, `cancelled` |

Server enum (`app/api/execute/_lib/types.ts`): `pending | running | unconfirmed | completed | failed`. There is no `queued` value.

HTTP 404 on the status route is `{ "error": "Execution not found" }` with no `status` field (missing id, or an execution that belongs to another organization). It is not a status string.

`completed` is **not** by itself proof of on-chain success. Inspect `receipts[].receiptStatus`.

Receipt statuses (`lib/web3/verify-receipt.ts`): `success | reverted | not_found | timeout | safe_inner_failure`.
`reverted` and `safe_inner_failure` are conclusive. `not_found` and `timeout` mean the receipt could not be read; `completeExecution` settles those rows as `unconfirmed`, not `failed`.

## Rules

### R1 — Poll the same ID, never resubmit

If status is `pending` or `running`, continue status reads against the **same** execution ID. Do not issue a new write for the same logical intent while that execution ID remains durable.

`unconfirmed` means the transaction was broadcast but no receipt could be read yet. The server keeps that row open (`completedAt` stays null) and a reconciliation sweep settles it to `completed` or `failed` once the chain answers. A waiting client must **stop** there and report it, rather than poll to an expired budget and exit non-zero: a non-zero exit invites a re-run, and re-running an intent that may already be on chain is the double-spend this contract exists to prevent. Read the settled status later against the same execution ID.

**CLI conformance:** `kh ex transfer --wait` / `kh ex cc --wait` poll the same ID, and stop on `unconfirmed` with exit code 0, printing the status and transaction hash.

### R2 — Receipts (client invariant)

Never infer success from `status=completed` alone.

1. `receiptStatus=success` is the only successful receipt.
2. `reverted` and `safe_inner_failure` are failure.
3. If a body claims `completed` plus any non-success receipt, treat it as failure.

KEEP-966 (`completeExecution`) re-verifies every claimed hash before writing `completed`. A reverted receipt is stored `verified: false` and the row settles as `failed`. `{status:"completed", verified:true, receiptStatus:"reverted"}` is **not** an observed production envelope. Fixtures that use that shape are labeled `kind: defensive` so the client stays fail-closed if the gate regresses.

The shipped CLI implements `--require-verified` on `kh ex status`; the write commands have no such flag. Without it, `completed` with an empty `receipts` array is still treated as success, matching a no-hash completion. With it, the CLI exits non-zero unless the execution completed carrying at least one receipt and every receipt is `verified: true` with `receiptStatus: success`; `unconfirmed` fails the gate as not proven landed.

**CLI conformance:** wait paths fail on `status=failed` and on non-success receipts as above.

### R3 — Write retry → stable idempotency key

If a write is retried after transport failure (timeout, 5xx) before an execution ID is known, the client MUST reuse the same `Idempotency-Key`. After an execution ID is known, prefer R1.

The server (`lib/idempotency.ts` `idempotencyEarlyResponse`) returns HTTP 409 for two codes:

| code | retryable | client |
| --- | --- | --- |
| `idempotency_in_progress` | true | Retry the **same** key. Do not mint a new key. |
| `idempotency_conflict` | false | Fail. The key is bound to a different payload. Do not rotate. |

**CLI conformance:** `kh ex transfer` and `kh ex cc` set `Idempotency-Key` once per invocation. HTTP-layer 5xx retries reuse it. 409 in_progress retries it until `--timeout`. 409 conflict is a hard error. Use `--idempotency-key` to pin a key across process restarts.

### R4 — Terminal failure / malformed

Statuses `failed` (direct) and unparseable/malformed bodies are terminal for that attempt. Missing `status` after a 200 decode is **malformed**, not success.

An unknown future `status` string is **unrecognized** (not malformed, not success) so a server addition does not look like a corrupt body.

### R5 — Rate limit

HTTP `429` responses require backoff. They are not success. Preserve the idempotency key for the next write attempt of the same intent. The HTTP client does not auto-retry 429.

### R6 — Cold start 404 (bounded)

A first status read that returns HTTP 404 immediately after submit may be transient. During `--wait`, poll until `--timeout` before treating 404 as terminal.

`--watch` does not apply this tolerance. `kh ex st <id> --watch` against a missing or foreign-org id must exit on 404, not loop.

## Fixture mapping + conformance

| Fixture | Kind | Rule | Expect | Consumed by |
| --- | --- | --- | --- | --- |
| `pending.json` | observed | R1 | pending | `TestFixtures_ClassifyTable` |
| `running.json` | observed | R1 | pending | `TestFixtures_ClassifyTable` |
| `unconfirmed.json` | observed | R1 | unconfirmed | `TestFixtures_ClassifyTable` |
| `completed_with_tx.json` | observed | R2 | success | `TestFixtures_ClassifyTable` |
| `completed_without_tx.json` | classifier | R2 | failure (strict option) | `TestFixtures_ClassifyTable` |
| `reverted.json` | defensive | R2 | failure | `TestFixtures_ClassifyTable` + `TestRevertedIsNeverSuccess` |
| `safe_inner_failure.json` | defensive | R2 | failure | `TestFixtures_ClassifyTable` |
| `failed.json` | observed | R4 | failure | `TestFixtures_ClassifyTable` |
| `malformed.json` | observed | R4 | malformed | `TestFixtures_ClassifyTable` |
| `not_found.json` | observed | R6 | pending (HTTP 404) | `TestFixtures_ClassifyTable` |
| `rate_limited.json` | observed | R5 | rate_limited | `TestFixtures_ClassifyTable` |
| `cold_start.sequence.json` | observed | R6 | 404→pending→success | `TestColdStartSequence_R6` |

Every fixture carries `"version": 1`. Loaders fail on a missing or unsupported version. `TestFixtures_ExpectedCountsAndRules` asserts the expected file count so a rename to `*.sequence.json` cannot drop coverage silently.

Fixtures use the **flat** direct-execution wire shape (`executionId`, not nested `execution.id`).
