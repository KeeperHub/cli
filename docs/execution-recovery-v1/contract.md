# Execution recovery contract (normative)

Version: **1.0.0**  
Audience: KeeperHub CLI / MCP / HTTP adapter authors

## Definitions

- **Write**: any API call that may create or re-drive an onchain side effect (execute workflow, transfer, contract call).
- **Status read**: polling `get_execution` / `GET /api/execute/{id}/status` (or equivalent).
- **Chain evidence**: at least one transaction hash with a verified successful receipt for the expected predicate (token, amount, recipient, chain).
- **Idempotency key**: client-supplied key that must be byte-identical across retries of the same logical write.

## Rules

### R1 — Unconfirmed → poll, do not resubmit

If status is `queued`, `running`, `pending`, or `unconfirmed`, the client MUST continue status reads against the **same** execution ID. The client MUST NOT issue a new write for the same logical intent while that execution ID remains durable.

### R2 — No chain evidence → fail closed

If the caller requires payment/landing proof and status is `completed` (or `success`) but no transaction hash / verified receipt is present, the client MUST treat the outcome as **not successful** for payment purposes.

### R3 — Write retry → stable idempotency key

If a write is retried after transport failure (timeout, 429, 5xx) before an execution ID is known, the client MUST reuse the same idempotency key. After an execution ID is known, prefer R1.

### R4 — Terminal failure

Statuses `failed`, `reverted`, `not_found`, and unparseable/malformed bodies are terminal for that attempt. Clients MUST NOT invent a success path from partial fields.

### R5 — Rate limit

HTTP `429` responses require backoff. They are not success. Preserve the idempotency key for the next write attempt of the same intent.

### R6 — Cold start

A first status read that returns `not_found` immediately after submit may be transient. Clients SHOULD poll briefly before concluding not_found as terminal, unless the product documents otherwise.

## Fixture mapping

| Fixture file | Exercises |
| --- | --- |
| `queued.json` | R1 |
| `completed_with_tx.json` | happy path with chain evidence |
| `completed_without_tx.json` | R2 |
| `failed.json` | R4 |
| `unconfirmed.json` | R1 |
| `reverted.json` | R4 |
| `not_found.json` | R4 / R6 |
| `rate_limited.json` | R5 |
| `cold_start.json` | R6 |
| `malformed.json` | R4 |
