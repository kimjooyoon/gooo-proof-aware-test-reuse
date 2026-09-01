#!/usr/bin/env bash
set -euo pipefail

bin=${1:?path to the built evaluator binary is required}
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
result=${INTEGRATION_RESULT_OUT:?INTEGRATION_RESULT_OUT is required}
work=$(mktemp -d "${RUNNER_TEMP:-/tmp}/gooo-proof-aware-integration.XXXXXX")
trap 'rm -rf "$work"' EXIT
mkdir -p "$(dirname "$result")"

meta="$root/.gooo/proof-aware-test-reuse.gooo"
contract="$root/contracts/denominator-v1.json"
source="$root/examples/input/greeting.gooo"
receipts="$root/fixtures/receipts"
out="$work/run"
before=$(git -C "$root" status --porcelain=v1 -z --untracked-files=all | sha256sum | awk '{print $1}')
"$bin" run --meta "$meta" --contract "$contract" --source "$source" --receipts "$receipts" --out "$out" --mode full >/dev/null

for artifact in semantic-ir.json dependency-graph.json test-plan.json reuse-receipt.json execution.json human-report.md generated/program.go report.json; do
	test -f "$out/$artifact"
done
jq -e '
  .decision == "CLOSED" and .case_id == "greeting" and .mode == "full" and
  .plan.action == "SELECTED" and .metrics.total == 1 and .metrics.selected == 1 and
  .metrics.executed == 1 and .metrics.reused == 0 and .metrics.unknown == 0 and .metrics.failed == 0 and
  .authority.repository_writes == 0 and .authority.commit_authority == 0 and .authority.merge_authority == 0 and
  .authority.release_mutation == 0 and .authority.local_validation_commands == 0 and
  (.artifact_digests["semantic-ir.json"] | startswith("sha256:"))
' "$out/report.json" >/dev/null
jq -e '.schema == "gooo/proof-aware-test-reuse/program/v1" and .case_id == "greeting" and (.obligations | length) == 1' "$out/semantic-ir.json" >/dev/null
jq -e '.schema == "gooo/proof-aware-test-reuse/dependency-graph/v1" and .path_to_target == false' "$out/dependency-graph.json" >/dev/null
go build -trimpath -o "$work/generated.bin" "$out/generated/program.go"
test "$("$work/generated.bin")" = "hello from Gooo"
jq -e '.action == "SELECTED" and .verified == true and (.result_digest | startswith("sha256:"))' "$out/execution.json" >/dev/null
jq -e '.state == "PASS" and .obligation == "smoke"' "$out/reuse-receipt.json" >/dev/null

after=$(git -C "$root" status --porcelain=v1 -z --untracked-files=all | sha256sum | awk '{print $1}')
test "$before" = "$after"

jq -n \
	--arg schema "gooo-proof-aware-test-reuse/integration/v1" \
	--arg source_digest "$(jq -r '.source_digest' "$out/report.json")" \
	--arg contract_digest "$(jq -r '.contract_digest' "$out/report.json")" \
	--arg fixture_digest "$(jq -r '.fixture_digest' "$out/report.json")" \
	'{schema:$schema,source_digest:$source_digest,contract_digest:$contract_digest,fixture_digest:$fixture_digest,caller_owned_output:true,repository_writes:0,semantic_ir:true,dependency_graph:true,generated_artifact:true,generated_artifact_executed:true,execution_verified:true,human_report:true,deterministic_digest_fields:true}' > "$result"
