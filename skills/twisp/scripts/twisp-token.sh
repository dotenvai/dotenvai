#!/bin/sh
set -eu

command -v aws >/dev/null 2>&1 || { echo "twisp-token: aws is required" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "twisp-token: jq is required" >&2; exit 1; }

response="$(aws sts get-web-identity-token --audience ledger-service --signing-algorithm RS256)"
token="$(printf '%s' "$response" | jq -r '.WebIdentityToken // empty')"
[ -n "$token" ] || { echo "twisp-token: AWS returned no WebIdentityToken" >&2; exit 1; }
printf '%s\n' "$token"
