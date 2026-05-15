#!/usr/bin/env bash
# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0


set -euo pipefail

repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

staged_go_files=$(git diff --cached --name-only --diff-filter=ACMR -- '*.go')
staged_terraform_files=""
while IFS= read -r tf_file; do
  case "$tf_file" in
    examples/*|tests/*)
      staged_terraform_files+="$tf_file"$'\n'
      ;;
  esac
done < <(git diff --cached --name-only --diff-filter=ACMR -- '*.tf' '*.tfvars')

if [ -z "$staged_go_files" ] && [ -z "$staged_terraform_files" ]; then
  exit 0
fi

if [ -n "$staged_go_files" ]; then
  bin_ext=""
  if [ "$(go env GOOS)" = "windows" ]; then
    bin_ext=".exe"
  fi
  gci_bin="$repo_root/bin/gci$bin_ext"
  if [ ! -x "$gci_bin" ]; then
    set +e
    gci_install_output=$(make --no-print-directory gci 2>&1)
    gci_install_exit_code=$?
    set -e
    if [ "$gci_install_exit_code" -ne 0 ]; then
      printf '%s\n' "$gci_install_output"
      exit "$gci_install_exit_code"
    fi
  fi
  gci_flags=(
    -s standard
    -s default
    -s "prefix(k8s)"
    -s "prefix(sigs.k8s)"
    -s "prefix(github.com)"
    -s "prefix(gitlab)"
    -s "prefix(github.com/terraform-redhat/terraform-provider-rhcs)"
    --custom-order
    --skip-generated
    --skip-vendor
  )
fi

partially_staged_file_detected=0
for staged_files in "$staged_go_files" "$staged_terraform_files"; do
  while IFS= read -r file_path; do
    if [ -z "$file_path" ] || [ ! -f "$file_path" ]; then
      continue
    fi

    if ! git diff --quiet -- "$file_path"; then
      echo "Commit blocked: staged file has unstaged changes: $file_path"
      echo "Stage all changes for this file (or stash them) before committing."
      partially_staged_file_detected=1
    fi
  done <<< "$staged_files"
done

if [ "$partially_staged_file_detected" -ne 0 ]; then
  exit 1
fi

formatting_updated_files=""

while IFS= read -r go_file; do
  if [ -z "$go_file" ] || [ ! -f "$go_file" ]; then
    continue
  fi

  before_formatting=$(mktemp)
  cp "$go_file" "$before_formatting"

  "$gci_bin" write "${gci_flags[@]}" "$go_file" >/dev/null
  gofmt -s -w "$go_file"

  if ! cmp -s "$before_formatting" "$go_file"; then
    formatting_updated_files+="$go_file"$'\n'
  fi

  rm -f "$before_formatting"
done <<< "$staged_go_files"

while IFS= read -r tf_file; do
  if [ -z "$tf_file" ] || [ ! -f "$tf_file" ]; then
    continue
  fi

  before_formatting=$(mktemp)
  cp "$tf_file" "$before_formatting"

  terraform fmt "$tf_file" >/dev/null

  if ! cmp -s "$before_formatting" "$tf_file"; then
    formatting_updated_files+="$tf_file"$'\n'
  fi

  rm -f "$before_formatting"
done <<< "$staged_terraform_files"

if [ -n "$formatting_updated_files" ]; then
  echo "Commit blocked: formatting updates were applied to staged files:"
  printf '%s' "$formatting_updated_files"
  echo "Review the changes, stage them, and commit again."
  exit 1
fi
