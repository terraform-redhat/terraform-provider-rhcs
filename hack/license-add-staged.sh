#!/usr/bin/env bash
# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

ADDLICENSE="${ADDLICENSE:-addlicense}"

repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

# Get all staged files
all_staged_files=$(git diff --cached --name-only --diff-filter=ACMR)

if [ -z "$all_staged_files" ]; then
  exit 0
fi

# Filter out ignored files (matching add-license-header.sh patterns)
filtered_staged_files=""
while IFS= read -r file_path; do
  if [ -z "$file_path" ] || [ ! -f "$file_path" ]; then
    continue
  fi

  # Skip directories we never modify
  case "$file_path" in
    vendor/*|examples/*) continue ;;
  esac

  filtered_staged_files+="$file_path"$'\n'
done <<< "$all_staged_files"

if [ -z "$filtered_staged_files" ]; then
  exit 0
fi

# Check for partially staged files
partially_staged_file_detected=0
while IFS= read -r file_path; do
  if [ -z "$file_path" ] || [ ! -f "$file_path" ]; then
    continue
  fi

  if ! git diff --quiet -- "$file_path"; then
    echo "Commit blocked: staged file has unstaged changes: $file_path"
    echo "Stage all changes for this file (or stash them) before committing."
    partially_staged_file_detected=1
  fi
done <<< "$filtered_staged_files"

if [ "$partially_staged_file_detected" -ne 0 ]; then
  exit 1
fi

# Check license headers on staged files
missing_headers=0
files_with_missing_headers=""

while IFS= read -r file_path; do
  if [ -z "$file_path" ] || [ ! -f "$file_path" ]; then
    continue
  fi

  if ! "${ADDLICENSE}" -check -c "Red Hat" -l apache -y "" -s=only "$file_path" 2>/dev/null; then
    missing_headers=1
    files_with_missing_headers+="$file_path"$'\n'
  fi
done <<< "$filtered_staged_files"

# If headers are missing, add them and block the commit
if [ "$missing_headers" -ne 0 ]; then
  while IFS= read -r file_path; do
    if [ -z "$file_path" ] || [ ! -f "$file_path" ]; then
      continue
    fi
    "${ADDLICENSE}" -c "Red Hat" -l apache -y "" -s=only "$file_path" >/dev/null 2>&1
  done <<< "$files_with_missing_headers"

  echo "Commit blocked: license headers were added to staged files:"
  printf '%s' "$files_with_missing_headers"
  echo "Review the changes, stage them, and commit again."
  exit 1
fi
