#!/bin/bash
# Makes scalar-or-array query filters usable in the generated client.
# oapi-codegen emits anonymous structs with an inaccessible union for these
# filters. Use the array branch: one element produces the scalar wire form,
# and several elements produce repeated query keys. No operations are removed.
# Normalize provider descriptions to plain punctuation and name Polar as the
# speaker in generated comments. This changes documentation only.
# Empty JSON response schemas become json.RawMessage. Their shape is unknown;
# retaining JSON bytes avoids eagerly decoding arbitrary trees into interface{}.
# The generated client already imports encoding/json; an import extension would
# duplicate that import in the pinned generator's output.
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/patch-spec.log}"
exec > >(tee -a "$LOG_FILE") 2>&1

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

# shellcheck disable=SC2016 # jq variables are not shell expansions
readonly PATCH_FILTER='
.paths |= with_entries(.value |= with_entries(
  if (.value | type) == "object" and (.value.parameters? | type) == "array" then
    .value.parameters |= map(
      if .in == "query" and (.schema.anyOf? | type) == "array"
        and any(.schema.anyOf[]; .type? == "array")
        and any(.schema.anyOf[]; .type? != "array" and .type? != "null")
      then .schema = ((.schema | del(.anyOf)) +
        ([.schema.anyOf[] | select(.type? == "array")][0]))
      else . end)
  else . end))
| walk(if type == "object" and has("responses") then
    .responses |= with_entries(.value |=
      if (.content? | type) == "object" then
        .content |= with_entries(
          if (.key | endswith("json")) and .value.schema == {} then
            .value.schema = {"x-go-type": "json.RawMessage"}
          else . end)
      else . end)
  else . end)
| walk(if type == "object" then
    with_entries(if (.key == "description" or .key == "summary")
      and (.value | type) == "string"
      then .value |= (gsub("[—–]"; ",") | gsub("\\bwe\\b"; "Polar"))
      else . end)
  else . end)
'

main() {
	local tmp
	tmp="$(mktemp "${SPEC_DIR}/.openapi.patched.XXXXXX")"
	# shellcheck disable=SC2064 # local is gone when EXIT fires
	trap "rm -f -- '${tmp}'" EXIT

	jq "$PATCH_FILTER" "${SPEC_DIR}/openapi.json" >"$tmp"
	chmod 0644 "$tmp"
	mv -- "$tmp" "${SPEC_DIR}/openapi.patched.json"
	log INFO "normalized scalar-or-array query filters"
}

main "$@"
