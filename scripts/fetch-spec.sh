#!/bin/bash
# Downloads the Polar OpenAPI spec into api/. The spec is not committed;
# `go generate` fetches it as its first step.
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

readonly INDEX_URL="https://api.github.com/repos/polarsource/polar/contents/docs/openapi?ref=main"
readonly RAW_URL="https://raw.githubusercontent.com/polarsource/polar/main/docs/openapi"
readonly DOWNLOAD_TIMEOUT_SECONDS=60
readonly SPEC_NAME="openapi.json"
SPEC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/api"
readonly SPEC_DIR

on_error() {
	local exit_code="$1"
	local line="$2"

	printf '{"level":"ERROR","file":"%s","line":%d,"msg":"command failed","exit_code":%d}\n' \
		"${BASH_SOURCE[1]}" "$line" "$exit_code" >&2
}

log() {
	local level="$1"
	shift

	jq -nc \
		--arg time "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
		--arg level "$level" \
		--arg file "${BASH_SOURCE[1]##*/}" \
		--arg line "${BASH_LINENO[0]}" \
		--arg func "${FUNCNAME[1]:-main}" \
		--arg msg "$*" \
		'{time:$time,level:$level,file:$file,line:($line|tonumber),func:$func,msg:$msg}' >&2
}

main() {
	mkdir -p "$SPEC_DIR"
	local tmp index current_month filename
	tmp="$(mktemp "${SPEC_DIR}/.${SPEC_NAME}.XXXXXX")"
	index="${tmp}.index"
	# shellcheck disable=SC2064 # locals are gone when EXIT fires
	trap "rm -f -- '${tmp}' '${index}'" EXIT

	# A YYYY-MM version takes effect on the first day of that month in UTC.
	# Compare fixed-width month strings and reject future versions.
	current_month="$(date -u '+%Y-%m')"
	wget -q -T "$DOWNLOAD_TIMEOUT_SECONDS" -O "$index" "$INDEX_URL"
	filename="$(jq -er --arg month "$current_month" '
		[.[] | select(.type == "file") | .name
		| select(test("^[0-9]{4}-(0[1-9]|1[0-2])\\.openapi\\.json$"))
		| select(.[0:7] <= $month)] | sort | last // error("no active spec")
	' "$index")"
	log INFO "downloading spec=${filename}"
	wget -q -T "$DOWNLOAD_TIMEOUT_SECONDS" -O "$tmp" "${RAW_URL}/${filename}"
	jq -e --arg version "${filename%.openapi.json}" '
		(.openapi | startswith("3.")) and .info.version == $version
		and (.paths | type == "object" and length > 0)
	' "$tmp" >/dev/null
	chmod 0644 "$tmp"
	mv -- "$tmp" "${SPEC_DIR}/${SPEC_NAME}"
	log INFO "saved spec=${filename}"
}

main "$@"
