#!/usr/bin/env bash
set -euo pipefail

stage=${1:?stage name is required}
output=${2:?stage output path is required}
shift 2
if [ "$#" -eq 0 ]; then
	printf '%s\n' "a command is required" >&2
	exit 2
fi

mkdir -p "$(dirname "$output")"
time_file=$(mktemp "${RUNNER_TEMP:-/tmp}/gooo-proof-aware-time.XXXXXX")
trap 'rm -f "$time_file"' EXIT
start_ns=$(date +%s%N)
set +e
/usr/bin/time -f '%M' -o "$time_file" "$@"
status=$?
set -e
end_ns=$(date +%s%N)
wall_ms=$(( (end_ns - start_ns) / 1000000 ))
peak_rss_kib=$(awk 'NF {value=$1} END {if (value == "") value=0; print value+0}' "$time_file")

jq -n \
	--arg stage "$stage" \
	--arg command "$*" \
	--argjson wall_ms "$wall_ms" \
	--argjson peak_rss_kib "$peak_rss_kib" \
	--argjson exit_code "$status" \
	'{stage:$stage,wall_ms:$wall_ms,peak_rss_kib:$peak_rss_kib,exit_code:$exit_code,command:$command}' > "$output"
exit "$status"
