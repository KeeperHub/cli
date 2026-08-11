# Execution recovery

Agents and adapters that submit KeeperHub writes must recover safely when the
network flakes, a status read races ahead of persistence, or an onchain
receipt reverts.

This guide is the published summary of the normative contract in
[`execution-recovery-v1/contract.md`](./execution-recovery-v1/contract.md).

## Safe first-write sequence

1. Simulate when available (`"simulate": true`) and continue only if the call would not revert.
2. Broadcast once with a stable `Idempotency-Key` that names the **work**, not the attempt.
3. Save `executionId`.
4. Poll `GET /api/execute/{executionId}/status`.
5. Treat `receipts[]` as authoritative: `verified` + `receiptStatus=success` prove landing. `receiptStatus=reverted` is failure even if `status=completed`.

See also the Direct Execution API docs.

## CLI behaviour

- `kh ex transfer` / `kh ex cc` attach `Idempotency-Key` automatically; override with `--idempotency-key`.
- `--wait` polls the same execution ID, tolerates a bounded initial `404`/`not_found`, and fails closed on reverted receipts.
- Workflow run status uses a different vocabulary (`success` / `error` / `cancelled`) — do not mix it with direct-execution statuses.

## Fixtures

Golden responses live under `testdata/execution_recovery_v1/` and are loaded by
`go test ./internal/execrecovery/...`.
