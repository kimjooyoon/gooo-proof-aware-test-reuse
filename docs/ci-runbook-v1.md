# CI runbook v1

The conformance job is the validation authority. It records compile, build, test, conformance, and integration measurements as integer `wall_ms` and `peak_rss_kib` fields, then joins them with Go test totals and the nine-case semantic report.

The evidence keeps product decisions (`CLOSED`, `UNKNOWN`, `REFUTED`) separate from the authoring procedure audit. The audit is intentionally retained as `OPERATIONAL_REFUTED` with `exact_count: 1`, `stage: AUTHORING`, `step: OPEN_IMPLEMENTATION_PR_BEFORE_MAIN_INTEGRATION`, and `reason: INITIAL_IMPLEMENTATION_PUSH_PRECEDED_PR`; `pr_first_conformance` therefore remains `REFUTED` even after the remediation/conformance PR and main validation succeed.
