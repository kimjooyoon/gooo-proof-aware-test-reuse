# Proof-aware test reuse v1

## Scope

The `.gooo` layer is authoritative for meaning. It declares claims, obligations, dependency edges, receipt policy and fields, change-impact policy, activities, and forbidden effects. The Go implementation is a generator/evaluator/runtime for those declarations; it is not an alternate semantic authority.

## Proof key

Each obligation is identified by this exact tuple:

```text
(source_digest, contract_digest, fixture_digest, toolchain_digest, runner_digest)
```

The receipt must also name the obligation, use the v1 receipt schema, state `PASS`, and carry a non-empty result digest. A result digest mismatch is stale evidence. No timestamp, score, percentage, or weighted sum can substitute for the proof key.

## Planning

The dependency graph is built from semantic nodes and directed edges. Every declared impact node is placed at the invalidation frontier. Reuse is allowed only when the frontier has no path to the obligation. A path, an explicit receipt failure, a forbidden effect, or a broken edge prevents reuse. The plan keeps four disjoint action classes:

* `SELECTED`: the obligation must run;
* `REUSED`: an exact PASS receipt proves the obligation;
* `UNKNOWN`: information is insufficient, with `stage`, `step`, `reason`, `unknown_class`, `next_operation`, and `blocked_by` always present;
* `REFUTED`: an explicit contradiction or boundary violation exists.

The report reduces claims using `REFUTED > UNKNOWN > CLOSED`. It never infers `CLOSED` or `FIXED_POINT` from an unknown top-level decision.

## Indicator vector and authority

The `.gooo` contract declares seven indicators: `build_wall_ms`, `build_peak_rss_kib`, `test_wall_ms`, `test_peak_rss_kib`, `selected`, `executed`, and `reused`. Each matched CI pair records exact integer `before`, `after`, and `signed_delta = after - before` values, plus its direction, observation, and independent claim state. The vector is the only improvement evidence; there is no aggregate improvement score or state.

For `DECREASE` indicators, `after < before` is `IMPROVED`/`CLOSED`, equality is `UNCHANGED`/`NOT_CLAIMED`, and `after > before` is `REGRESSED`/`REFUTED`. `reused` has an `INCREASE` goal, so the signs reverse. A decrease in `selected` or `executed` is `CLOSED` only when the same CI job proves exact PASS receipt reuse; otherwise the reduction is not a closed proof claim. The scenario, source, contract, fixture, toolchain, runner, and CI job identity must match. Utility remains `UNKNOWN` without external user evidence.

Runtime authority is separate from operator authority. Runtime values are zero for repository writes, commit, push, merge, release mutation, and local validation commands. CI is the only validation authority for this repository; local validation is intentionally not run by the implementation workflow. If a local validation command is ever executed, its exact count must be preserved and the state becomes `OPERATIONAL_REFUTED`.
