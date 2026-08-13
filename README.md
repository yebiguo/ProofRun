<div align="center">

# ProofRun

**A local verification receipt for AI coding agents.**

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go 1.22+](https://img.shields.io/badge/go-1.22%2B-00ADD8?logo=go&logoColor=white)](go.mod)
[![CI](https://github.com/yebiguo/proofrun/actions/workflows/test.yml/badge.svg)](https://github.com/yebiguo/proofrun/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/yebiguo/proofrun/graph/badge.svg)](https://codecov.io/gh/yebiguo/proofrun)
[![Release](https://img.shields.io/github/v/release/yebiguo/proofrun?include_prereleases&label=release)](https://github.com/yebiguo/proofrun/releases)

[English](README.md) · [简体中文](README.zh-CN.md)

</div>

---

![ProofRun: run a check, it PASSes, edit the code, it goes STALE automatically](docs/demo.gif)

ProofRun doesn't judge whether your code is correct. It proves — cryptographically, not by asking nicely — which checks actually ran against the exact code you have right now.

## The problem

An AI coding agent says "all tests pass." Is that true?

Maybe. It was true the last time the agent actually ran the tests. But that might have been three edits ago. The agent might not even remember running them — it might just be inferring "the change looks right, tests probably still pass." From the words alone, you have no way to tell "I ran it and it passed" apart from "I'm pretty sure it would pass."

ProofRun closes that gap. Not by making the agent more honest — by making the claim itself checkable.

## How it works

```bash
$ proofrun run test -- pytest
...
test: pass (exit 0, 1841ms)

$ proofrun status
test                 PASS    (exit 0, 1841ms)

# code changes after this point — agent or human, doesn't matter

$ proofrun status
test                 STALE   (last run: pass, exit 0 — code changed since)
```

Every check result is bound to a fingerprint of your exact code state: the git commit, plus a hash of everything uncommitted — staged or not, tracked or not. Change a single byte, and the result flips to `STALE` automatically. Nobody has to remember to ask "does this PASS still count?"

## Install

```bash
curl -L https://github.com/yebiguo/proofrun/releases/download/v0.1.0/proofrun_linux_amd64.tar.gz | tar xz
# other platforms: https://github.com/yebiguo/proofrun/releases
```

Or build from source:

```bash
go install github.com/yebiguo/proofrun/cmd/proofrun@latest
```

## Quick start

```bash
proofrun init                      # writes .proofrun.yml
proofrun run test -- pytest        # runs pytest for real, binds the result
proofrun status --strict           # non-zero exit if anything isn't PASS
```

## Why this, not just trusting the agent

- **No LLM calls, anywhere.** ProofRun doesn't use AI to verify AI. It starts a real subprocess and reads its real exit code — that's the entire mechanism.
- **Four statuses, never a guess.** `PASS`, `FAIL`, `STALE`, `NOT RUN` — each one comes from an observed execution, or the documented absence of one. There's no fifth "probably fine."
- **Fully offline.** Zero network calls, zero telemetry, zero accounts.
- **Argv-exact, not string-matched.** A check declared as `pytest -k "foo bar"` can't be satisfied by a command that merely looks similar once flattened to text — ProofRun compares real argument arrays, not strings.

## What ProofRun deliberately does not do

It does not parse test output, does not judge code quality, and does not auto-fix anything. See [AGENTS.md](AGENTS.md) for the complete boundary.

## Built by an AI agent, held accountable by one

ProofRun was written by an AI coding agent (Claude Code) under human direction, then went through several rounds of independent, read-only adversarial review before the first release. That review found that ProofRun's own command comparison could be tricked: a misquoted shell argument made a check silently run zero tests and still report `PASS`. Full repro, the exact fix, and why a simple patch wasn't enough → [docs/case-study.md](docs/case-study.md).

Every fix was verified against a real reproduction before being accepted — not just reviewed for plausibility. A tool built to hold AI agents accountable has no business existing if it can't survive that same scrutiny applied to itself.

## Commands

```bash
proofrun init                      # generate .proofrun.yml
proofrun run <check-name> -- <cmd> # run <cmd> for real, bind exit code + duration to current git state
proofrun run-all [--only <name>]   # run every declared check, saving a result after each one
proofrun status [--strict]         # PASS / FAIL / STALE / NOT RUN per check; --strict exits non-zero if a required check isn't PASS
proofrun report [--json]           # full report, human- or machine-readable
```

## Config: `.proofrun.yml`

```yaml
checks:
  test:
    command: [pytest]
    required: true
  build:
    command: [npm, run, build]
    required: true
  lint:
    command: [ruff, check, .]
    required: false
```

`command` is an argv list, not a shell string — ProofRun never goes through a shell, and comparing what actually ran against what's declared has to be exact, element for element. `required: true` is what makes a check block `status --strict`, which is what you'd wire into a pre-commit hook or CI gate.

## How the fingerprint works

Every result is bound to your current git `HEAD` plus a SHA-256 hash of `git diff HEAD` combined with the contents of any untracked, non-ignored files. `proofrun status` recomputes that fingerprint every time and compares it against what's stored locally — any mismatch, down to a single changed space or one new file, reports `STALE`.

## GitHub Action

```yaml
on: pull_request
jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: yebiguo/proofrun@v1
```

This does its own checkout of the exact PR head commit — it never trusts whatever the calling workflow already checked out, so a `pull_request` trigger can't silently hand it GitHub's synthetic merge-preview commit instead. It then clears out any `receipt.json` that came in on the PR branch, downloads a checksum-verified `proofrun` binary, and runs `proofrun run-all` for real before gating on `proofrun status --strict`. Nothing about a receipt checked into the PR branch is ever trusted — every result the gate sees was produced by this run.

**Known limitation:** this does not protect `.proofrun.yml` itself from being weakened by the same PR that changes the code — a PR could loosen or remove a check's command and the Action would faithfully re-run the weaker version. It warns (via a build annotation) when `.proofrun.yml` differs from the PR's base branch, but it does not block on that; review that diff the same way you'd review any other part of the change.

## Roadmap

- **v0.3** — structured output support for common test runners (pytest, Jest, JUnit)
- Signed, tamper-evident receipts are on the radar, not yet designed
- Protecting `.proofrun.yml` itself from being weakened within the same PR that changes the code (currently only warned about, not blocked — see "Known limitation" above)

## Contributing

Issues and PRs welcome. This is a young project (v0.1) with a narrow, deliberate scope — see [AGENTS.md](AGENTS.md) before proposing anything that touches STALE detection or the receipt schema; those are the parts this project can least afford to get wrong.

## License

MIT
# trivial change to give this branch a distinct head commit
