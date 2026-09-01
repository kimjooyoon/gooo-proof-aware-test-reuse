# Gooo proof-aware test reuse

This repository advances Gooo as a self-improving programming language. A `.gooo` program owns semantic claims, test obligations, dependency edges, immutable receipt references, and change impact. Go is limited to parsing, semantic IR and graph generation, planning, generated-code evaluation, and reporting.

The proof-aware planner reduces redundant test work without treating missing evidence as proof. A test is `REUSED` only when all five proof-key digests match—source, contract, fixture, toolchain, and runner—the invalidation frontier has no path to the obligation, and the receipt is an explicit immutable `PASS`. A stale, missing, or ambiguous receipt is `UNKNOWN`; it is never reused. `REFUTED > UNKNOWN > CLOSED` is the only decision precedence, and unknown top-level decisions fail closed.

## End-to-end path

```text
input .gooo
  -> semantic-ir.json
  -> dependency-graph.json + invalidation frontier
  -> deterministic test-plan.json
  -> generated/program.go
  -> execution.json + reuse-receipt.json
  -> human-report.md
```

Run the evaluator from a Go 1.27 environment:

```text
go run ./cmd/gooo-proof-aware-test-reuse run \
  --meta .gooo/proof-aware-test-reuse.gooo \
  --contract contracts/denominator-v1.json \
  --source examples/input/greeting.gooo \
  --receipts fixtures/receipts \
  --out /caller-owned/output \
  --mode full
```

`/caller-owned/output` is created by the caller. The evaluator never writes to the repository, commits, pushes, merges, tags, or releases. The integration workflow compiles and executes the generated Go program from that output directory and verifies the human-readable report.

## Fixed semantic denominator

`contracts/denominator-v1.json` fixes nine canonical cases: three `CLOSED`, three `UNKNOWN`, and three `REFUTED`. CI runs all nine in incremental mode and records separate `selected`, `executed`, `reused`, `unknown`, and `failed` counts. The suite includes exact PASS reuse, dependency-impact selection, missing/stale/ambiguous receipts, explicit failure, forbidden effect, and fail-closed handling for an unknown top-level decision. Broken dependency edges are also refuted by the engine.

The same CI job also runs the identical `greeting` scenario once as a full baseline and once as a proof-aware incremental run. It records a seven-item indicator vector with exact integer `before`, `after`, and signed `after-before` values for build/test wall time, build/test peak RSS, selected, executed, and reused. Each item carries its own observation and claim state; there is no aggregate improvement score or state. A decrease is `IMPROVED` only when `after < before`, an equal value is `UNCHANGED` with `NOT_CLAIMED`, and an increase against a decrease goal is `REGRESSED` with `REFUTED`. The selected/executed decrease is `CLOSED` only when the same CI job also proves exact PASS receipt reuse. Reused uses the opposite increase goal. Utility remains `UNKNOWN` until external user workload evidence exists.

## CI evidence

The GitHub Actions workflow records `wall_ms` and `peak_rss_kib` for compile, build, test, conformance, and integration. It also records Go/Gooo file counts, physical lines, descendant directories, regular files, generated artifact count and bytes, fixed-denominator decisions, and runtime/operator authority counts. `cross_project_required_gates` is fixed at `0`, and workflows use the standard `github.token`.

The release workflow runs only from `main`, creates one annotated tag, publishes one immutable evidence tarball, and verifies the release asset digest and tag target through the GitHub API. Failed runs, tags, releases, and assets are not deleted or recreated.

See [docs/proof-aware-test-reuse-v1.md](docs/proof-aware-test-reuse-v1.md) for the normative semantic contract.
