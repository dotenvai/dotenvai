#!/bin/sh
set -eu

[ "$#" -eq 1 ] || { echo "usage: twisp-schema.sh OUTPUT.graphql" >&2; exit 2; }
: "${TWISP_REGION:?set TWISP_REGION}"
: "${TWISP_ACCOUNT_ID:?set TWISP_ACCOUNT_ID}"
: "${TWISP_TOKEN:?set TWISP_TOKEN}"

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT HUP INT TERM
printf 'query sdl { _service { sdl } }\n' > "$tmp"
"$script_dir/twisp-gql.sh" "$tmp" | jq -er '.data._service.sdl' > "$1"
printf 'wrote %s\n' "$1" >&2
