<!--
Title format:  <type>: #<issue> <description>
    feat: #97 add --require-verified to execute status
    fix(execute): #98 bound a hung request under --watch --timeout

The issue must carry the `accepted` label before you open this. See ISSUES.md.
Exempt: docs / chore / style changes from the "no issue required" list.
The type prefix drives release-please, so keep it accurate.
-->

## Issue

Closes #

<!-- If this needs no issue, say which exemption applies and delete the line above. -->

## What this changes

<!--
What the diff does and why. Name anything a reader would not predict from the
title: a changed default, a new dependency, altered credential handling.
-->

## Scope

<!--
Confirm this is one change: could any part of it ship and be correct with the
rest reverted? If yes, split it.
-->

## How it was verified

<!--
Tests added and what they would catch. For a bug fix, the test that fails
without it. For a timeout or deadline, a test against a handler that blocks
past it - a promptly-responding server does not cover the case.
-->

---

- [ ] Targets `main`
- [ ] Title carries the issue number, or an exemption applies
- [ ] `make lint` and `make test` pass
- [ ] `go generate ./docs/` run and committed, if a command or flag changed
