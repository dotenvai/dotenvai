#!/bin/sh
set -eu

usage="usage: twisp-gql.sh OPERATION.graphql [VARIABLES.json]"
[ "$#" -ge 1 ] && [ "$#" -le 2 ] || { echo "$usage" >&2; exit 2; }
: "${TWISP_REGION:?set TWISP_REGION}"
: "${TWISP_ACCOUNT_ID:?set TWISP_ACCOUNT_ID}"
: "${TWISP_TOKEN:?set TWISP_TOKEN}"
command -v curl >/dev/null 2>&1 || { echo "twisp-gql: curl is required" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "twisp-gql: jq is required" >&2; exit 1; }

case "$TWISP_REGION" in
  us-east-1|us-east-2|us-west-1|us-west-2|ap-southeast-2) ;;
  *) echo "twisp-gql: unsupported public Cloud region: $TWISP_REGION" >&2; exit 2 ;;
esac

operation_file="$1"
variables_file="${2:-}"
[ -f "$operation_file" ] || { echo "twisp-gql: operation file not found" >&2; exit 2; }
query="$(cat "$operation_file")"
if [ -n "$variables_file" ]; then
  [ -f "$variables_file" ] || { echo "twisp-gql: variables file not found" >&2; exit 2; }
  payload="$(jq -n --arg query "$query" --slurpfile variables "$variables_file" '{query:$query,variables:$variables[0]}')"
else
  payload="$(jq -n --arg query "$query" '{query:$query,variables:{}}')"
fi

curl --fail-with-body --silent --show-error \
  "https://api.${TWISP_REGION}.cloud.twisp.com/financial/v1/graphql" \
  -H "authorization: Bearer ${TWISP_TOKEN}" \
  -H "x-twisp-account-id: ${TWISP_ACCOUNT_ID}" \
  -H "content-type: application/json" \
  --data-binary "$payload" |
  jq -e 'if (.errors? | length > 0) then error(.errors | tostring) else . end'
