#!/usr/bin/env bash
set -euo pipefail

bin=${1:?path to the built evaluator binary is required}
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
work=${CONFORMANCE_WORK_ROOT:?CONFORMANCE_WORK_ROOT is required}
counts_out=${CONFORMANCE_COUNTS_OUT:?CONFORMANCE_COUNTS_OUT is required}
pair_out=${REUSE_PAIR_OUT:?REUSE_PAIR_OUT is required}
mkdir -p "$work" "$(dirname "$counts_out")" "$(dirname "$pair_out")"

meta="$root/.gooo/proof-aware-test-reuse.gooo"
contract="$root/contracts/denominator-v1.json"
cases="$root/fixtures/cases"
receipts="$root/fixtures/receipts"
suite_out="$work/suite"
"$bin" suite --meta "$meta" --contract "$contract" --cases "$cases" --receipts "$receipts" --out "$suite_out" --mode incremental >/dev/null

jq -e '
  .decision == "CLOSED" and .fixed_denominator == 9 and
  .actual.total == 9 and .actual.selected == 1 and .actual.executed == 1 and
  .actual.reused == 2 and .actual.unknown == 3 and .actual.failed == 3 and
  .actual_states.CLOSED == 3 and .actual_states.UNKNOWN == 3 and .actual_states.REFUTED == 3 and
  .expected_states.CLOSED == 3 and .expected_states.UNKNOWN == 3 and .expected_states.REFUTED == 3
' "$suite_out/suite-report.json" >/dev/null

jq -n \
	--arg schema "gooo-proof-aware-test-reuse/conformance-counts/v1" \
	--slurpfile suite "$suite_out/suite-report.json" \
	'{schema:$schema,total:$suite[0].actual.total,selected:$suite[0].actual.selected,executed:$suite[0].actual.executed,reused:$suite[0].actual.reused,unknown:$suite[0].actual.unknown,refuted:$suite[0].actual.failed,closed:$suite[0].actual_states.CLOSED,generated_artifacts:$suite[0].generated_artifacts,generated_bytes:$suite[0].generated_bytes}' > "$counts_out"

pair_root="$work/proof-pair"
mkdir -p "$pair_root/before" "$pair_root/after"
source="$root/examples/input/greeting.gooo"
measure="$root/scripts/measure-stage.sh"
before_out="$pair_root/before/run"
after_out="$pair_root/after/run"

"$measure" proof-plan-before "$pair_root/before/plan-stage.json" "$bin" run --meta "$meta" --contract "$contract" --source "$source" --receipts "$receipts" --out "$before_out" --mode full >/dev/null
"$measure" proof-build-before "$pair_root/before/build-stage.json" go build -trimpath -o "$pair_root/before/generated.bin" "$before_out/generated/program.go"
"$measure" proof-test-before "$pair_root/before/test-stage.json" bash -c 'test "$($1)" = "hello from Gooo"' _ "$pair_root/before/generated.bin"

"$measure" proof-plan-after "$pair_root/after/plan-stage.json" "$bin" run --meta "$meta" --contract "$contract" --source "$source" --receipts "$receipts" --out "$after_out" --mode incremental >/dev/null
"$measure" proof-build-after "$pair_root/after/build-stage.json" go build -trimpath -o "$pair_root/after/generated.bin" "$after_out/generated/program.go"
"$measure" proof-test-after "$pair_root/after/test-stage.json" jq -e '.plan.action == "REUSED" and .metrics.selected == 0 and .metrics.executed == 0 and .metrics.reused == 1 and .verification.reused_proof == true' "$after_out/report.json"

jq -e '
  .decision == "CLOSED" and .plan.action == "SELECTED" and
  .metrics.selected == 1 and .metrics.executed == 1 and .metrics.reused == 0 and
  .authority.repository_writes == 0 and .authority.local_validation_commands == 0
' "$before_out/report.json" >/dev/null
jq -e '
  .decision == "CLOSED" and .plan.action == "REUSED" and
  .metrics.selected == 0 and .metrics.executed == 0 and .metrics.reused == 1 and
  .authority.repository_writes == 0 and .authority.local_validation_commands == 0
' "$after_out/report.json" >/dev/null

jq -n \
	--arg schema "gooo-proof-aware-test-reuse/improvement-pair/v1" \
	--arg scenario "greeting" \
	--arg source_digest "$(jq -r '.source_digest' "$before_out/report.json")" \
	--arg contract_digest "$(jq -r '.contract_digest' "$before_out/report.json")" \
	--arg fixture_digest "$(jq -r '.fixture_digest' "$before_out/report.json")" \
	--arg toolchain "$(jq -r '.toolchain' "$before_out/report.json")" \
	--arg runner "$(jq -r '.runner' "$before_out/report.json")" \
	--arg job "${GITHUB_RUN_ID:-unknown}/${GITHUB_JOB:-unknown}" \
	--slurpfile before_report "$before_out/report.json" \
	--slurpfile after_report "$after_out/report.json" \
	--slurpfile before_build "$pair_root/before/build-stage.json" \
	--slurpfile after_build "$pair_root/after/build-stage.json" \
	--slurpfile before_test "$pair_root/before/test-stage.json" \
	--slurpfile after_test "$pair_root/after/test-stage.json" \
	'{schema:$schema,scenario:$scenario,job:$job,source_digest:$source_digest,contract_digest:$contract_digest,fixture_digest:$fixture_digest,toolchain:$toolchain,runner:$runner,before:{action:$before_report[0].plan.action,selected:$before_report[0].metrics.selected,executed:$before_report[0].metrics.executed,reused:$before_report[0].metrics.reused,build_wall_ms:$before_build[0].wall_ms,build_peak_rss_kib:$before_build[0].peak_rss_kib,test_wall_ms:$before_test[0].wall_ms,test_peak_rss_kib:$before_test[0].peak_rss_kib},after:{action:$after_report[0].plan.action,selected:$after_report[0].metrics.selected,executed:$after_report[0].metrics.executed,reused:$after_report[0].metrics.reused,build_wall_ms:$after_build[0].wall_ms,build_peak_rss_kib:$after_build[0].peak_rss_kib,test_wall_ms:$after_test[0].wall_ms,test_peak_rss_kib:$after_test[0].peak_rss_kib},exact_before_after_integer_pair:true,same_scenario_source_contract_fixture_toolchain_runner:true,improvement:{state:"CLOSED",reason:"EXACT_BEFORE_AFTER_INTEGER_PAIR_IN_ONE_CI_JOB",utility_state:"UNKNOWN",utility_reason:"NO_EXTERNAL_USER_EVIDENCE"}}' > "$pair_out"

jq -e '
  .exact_before_after_integer_pair == true and .same_scenario_source_contract_fixture_toolchain_runner == true and
  .before.action == "SELECTED" and .before.selected == 1 and .before.executed == 1 and .before.reused == 0 and
  .after.action == "REUSED" and .after.selected == 0 and .after.executed == 0 and .after.reused == 1 and
  (.before.build_wall_ms | type) == "number" and (.after.build_wall_ms | type) == "number" and
  (.before.build_peak_rss_kib | type) == "number" and (.after.build_peak_rss_kib | type) == "number" and
  (.before.test_wall_ms | type) == "number" and (.after.test_wall_ms | type) == "number" and
  (.before.test_peak_rss_kib | type) == "number" and (.after.test_peak_rss_kib | type) == "number" and
  .improvement.state == "CLOSED"
' "$pair_out" >/dev/null
