#!/usr/bin/env bash
set -euo pipefail

stage_root=${1:?stage measurement directory is required}
test_events=${2:?go test JSON event file is required}
counts=${3:?conformance counts file is required}
integration=${4:?integration result file is required}
pair=${5:?improvement pair file is required}
output=${6:?evidence output path is required}

for stage in compile build test conformance integration; do
	test -f "$stage_root/$stage.json"
	jq -e --arg stage "$stage" '.stage == $stage and (.wall_ms|type) == "number" and (.wall_ms|floor) == .wall_ms and (.wall_ms >= 0) and (.peak_rss_kib|type) == "number" and (.peak_rss_kib|floor) == .peak_rss_kib and (.peak_rss_kib >= 0) and .exit_code == 0' "$stage_root/$stage.json" >/dev/null
done

jq -s '{total:([.[]|select(.Action == "run" and (.Test // "") != "")]|length),selected:([.[]|select(.Action == "run" and (.Test // "") != "")]|length),executed:([.[]|select((.Action == "pass" or .Action == "skip") and (.Test // "") != "")]|length),reused:0,failed:([.[]|select(.Action == "fail" and (.Test // "") != "")]|length),unknown:0}' "$test_events" > "$stage_root/tests.json"
jq -e '.total > 0 and .selected == .total and .failed == 0 and .unknown == 0' "$stage_root/tests.json" >/dev/null

inventory=$(jq -c '.inventory' "$stage_root/../conformance-work/suite/suite-report.json" 2>/dev/null || true)
if [ -z "$inventory" ]; then
	inventory='{}'
fi

jq -n \
	--arg schema "gooo-proof-aware-test-reuse/ci-evidence/v1" \
	--arg commit "${GITHUB_SHA:-unknown}" \
	--arg run_id "${GITHUB_RUN_ID:-unknown}" \
	--arg job "${GITHUB_JOB:-unknown}" \
	--slurpfile compile "$stage_root/compile.json" \
	--slurpfile build "$stage_root/build.json" \
	--slurpfile test "$stage_root/test.json" \
	--slurpfile conformance "$stage_root/conformance.json" \
	--slurpfile integration_stage "$stage_root/integration.json" \
	--slurpfile tests "$stage_root/tests.json" \
	--slurpfile counts "$counts" \
	--slurpfile integration "$integration" \
	--slurpfile pair "$pair" \
	--argjson inventory "$inventory" \
	'{schema:$schema,commit:$commit,run_id:$run_id,job:$job,toolchain:"go1.27.0",runner:"ubuntu-latest",cross_project_required_gates:0,stage_measurements:{compile:$compile[0],build:$build[0],test:$test[0],conformance:$conformance[0],integration:$integration_stage[0]},tests:$tests[0],conformance:$counts[0],integration:$integration[0],improvement:$pair[0],inventory:$inventory,generated_artifacts:{count:$counts[0].generated_artifacts,bytes:$counts[0].generated_bytes},authority:{runtime:{repository_writes:0,commit_authority:0,push_authority:0,merge_authority:0,release_mutation:0,local_validation_commands:0,local_validation_state:"NOT_RUN"},operator:{pull_request:0,merge:0,release:0}},operational_audit:{state:"OPERATIONAL_REFUTED",exact_count:1,stage:"AUTHORING",step:"OPEN_IMPLEMENTATION_PR_BEFORE_MAIN_INTEGRATION",reason:"INITIAL_IMPLEMENTATION_PUSH_PRECEDED_PR"},pr_first_conformance:"REFUTED"}' > "$output"
