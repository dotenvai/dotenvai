#!/bin/sh
set -eu

cd "${GITHUB_WORKSPACE:-/github/workspace}"
git config --global --add safe.directory "$PWD"

set -- scan --format github --fail-on "${INPUT_FAIL_ON:-high}"
if [ -n "${INPUT_BASE:-}" ]; then
  set -- "$@" --base "$INPUT_BASE" --head "${INPUT_HEAD:-HEAD}"
elif [ "${INPUT_STAGED:-false}" = "true" ]; then
  set -- "$@" --staged
fi

exec dotenvai "$@"
