#!/usr/bin/env bash

set -euo pipefail

repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

declare -a terraform_targets=()
for path in examples tests; do
  if [ -d "$path" ]; then
    terraform_targets+=("$path")
  fi
done

go_format_needed=0
terraform_output=""

set +e
go_check_output=$(make --no-print-directory fmt_go_check 2>&1)
go_check_exit_code=$?
set -e

if [ "$go_check_exit_code" -ne 0 ] && [ -n "$go_check_output" ]; then
  printf '%s\n' "$go_check_output"
  exit "$go_check_exit_code"
fi

if [ "$go_check_exit_code" -ne 0 ]; then
  go_format_needed=1
fi

for tf_target in "${terraform_targets[@]}"; do
  set +e
  tf_output=$(terraform fmt -check -recursive "$tf_target" 2>&1)
  tf_exit_code=$?
  set -e

  if [ "$tf_exit_code" -eq 0 ]; then
    continue
  fi

  if [ "$tf_exit_code" -ne 3 ]; then
    printf '%s\n' "$tf_output"
    exit "$tf_exit_code"
  fi

  if [ -n "$tf_output" ]; then
    terraform_output+="$tf_output"$'\n'
  fi
done

if [ "$go_format_needed" -eq 0 ] && [ -z "$terraform_output" ]; then
  exit 0
fi

if [ "$go_format_needed" -ne 0 ]; then
  make --no-print-directory fmt_go >/dev/null
fi

for tf_target in "${terraform_targets[@]}"; do
  terraform fmt -recursive "$tf_target" >/dev/null
done

echo "Formatting updates were applied (gci + gofmt + terraform fmt). Command failed so you can review and stage changes."
if [ "$go_format_needed" -ne 0 ]; then
  echo "Go formatting and/or import-order updates were applied."
fi
if [ -n "$terraform_output" ]; then
  echo "Files that needed terraform fmt:"
  printf '%s' "$terraform_output"
fi
echo "Run your commit flow again after reviewing/staging the updates."
exit 1
