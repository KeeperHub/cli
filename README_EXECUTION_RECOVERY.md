# Execution Recovery Contract Pack v1

Fixture suite for [KeeperHub/cli#53](https://github.com/KeeperHub/cli/issues/53).

Does **not** implement CLI `--require-verified` (see #95). Complements Option B of #53 with golden status envelopes + recovery rules.

## Layout
- `testdata/execution_recovery_v1/*.json` — synthetic DEMO FIXTURE envelopes
- `docs/execution-recovery-v1/contract.md` — normative rules R1–R6

## Consumer note
A TypeScript reference consumer lives in the EMBER repo. Prefer Go tests here that load these fixtures and assert the decision table in `contract.md`.
