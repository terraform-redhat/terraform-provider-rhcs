#!/usr/bin/env bash
# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=lib/git.sh
source "${script_dir}/lib/git.sh"

repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

provider_prefix="rhcs"
allowlist_file="hack/subsystem-registry-allowlist.yaml"
base_ref="${SUBSYSTEM_REGISTRY_BASE_REF:-}"

tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/rhcs-subsystem-registry-XXXXXX")
registered_types_file="$tmp_dir/registered.txt"
subsystem_types_file="$tmp_dir/subsystem.txt"
allowlist_types_file="$tmp_dir/allowlist.txt"
new_types_file="$tmp_dir/new.txt"
trap 'rm -rf "$tmp_dir"' EXIT

collect_registered_types() {
  : >"$registered_types_file"
  while IFS= read -r line; do
    suffix=$(printf '%s' "$line" | sed -nE 's/.*ProviderTypeName \+ "(_[a-z0-9_]+)".*/\1/p')
    if [ -n "$suffix" ]; then
      echo "${provider_prefix}${suffix}"
    fi
  done < <(git grep -h -E 'ProviderTypeName \+ "_[a-z0-9_]+"' -- 'provider/*.go' 'provider/**/*.go' ':!*_test.go' 2>/dev/null || true) \
    >>"$registered_types_file"

  while IFS= read -r line; do
    suffix=$(printf '%s' "$line" | sed -nE 's/.*resourceTypeName[[:space:]]*=[[:space:]]*"(_[a-z0-9_]+)".*/\1/p')
    if [ -n "$suffix" ]; then
      echo "${provider_prefix}${suffix}"
    fi
  done < <(git grep -h -E 'resourceTypeName[[:space:]]*=[[:space:]]*"_([a-z0-9_]+)"' -- 'provider/*.go' 'provider/**/*.go' ':!*_test.go' 2>/dev/null || true) \
    >>"$registered_types_file"

  sort -u -o "$registered_types_file" "$registered_types_file"
}

collect_subsystem_types() {
  : >"$subsystem_types_file"
  if [ ! -d subsystem ]; then
    return
  fi

  while IFS= read -r line; do
    type_name=$(printf '%s' "$line" | sed -nE 's/.*"(rhcs_[a-z0-9_]+)".*/\1/p')
    if [ -n "$type_name" ]; then
      echo "$type_name"
    fi
  done < <(git grep -h -E '(resource|data)[[:space:]]+"rhcs_[a-z0-9_]+"' -- 'subsystem/*.go' 'subsystem/**/*.go' 2>/dev/null || true) \
    | sort -u >"$subsystem_types_file"
}

collect_allowlist_types() {
  : >"$allowlist_types_file"
  if [ ! -f "$allowlist_file" ]; then
    return
  fi

  awk '
    $1 == "-" && $2 == "type:" {
      print $3
    }
  ' "$allowlist_file" | sort -u >"$allowlist_types_file"
}

resolve_merge_base_range() {
  resolve_changed_go_diff_range "$base_ref" || true
}

collect_new_types() {
  : >"$new_types_file"
  local diff_range
  diff_range=$(resolve_merge_base_range)
  if [ -z "$diff_range" ]; then
    return
  fi

  while IFS= read -r line; do
    suffix=$(printf '%s' "$line" | sed -nE 's/.*"(_[a-z0-9_]+)".*/\1/p')
    if [ -n "$suffix" ]; then
      echo "${provider_prefix}${suffix}"
    fi
  done < <(git diff "$diff_range" -U0 -- 'provider/*.go' 'provider/**/*.go' ':!*_test.go' \
    | grep -E '^\+.*TypeName.*\+ "_[a-z0-9_]+"' || true) >>"$new_types_file"

  while IFS= read -r line; do
    suffix=$(printf '%s' "$line" | sed -nE 's/.*"(_[a-z0-9_]+)".*/\1/p')
    if [ -n "$suffix" ]; then
      echo "${provider_prefix}${suffix}"
    fi
  done < <(git diff "$diff_range" -U0 -- 'provider/*.go' 'provider/**/*.go' ':!*_test.go' \
    | grep -E '^\+.*resourceTypeName[[:space:]]*=[[:space:]]*"_([a-z0-9_]+)"' || true) >>"$new_types_file"

  sort -u -o "$new_types_file" "$new_types_file"
}

is_listed_in_file() {
  local needle=$1
  local file=$2
  grep -Fxq "$needle" "$file"
}

collect_registered_types
collect_subsystem_types
collect_allowlist_types
collect_new_types

registered_count=$(wc -l <"$registered_types_file" | tr -d ' ')
subsystem_count=$(wc -l <"$subsystem_types_file" | tr -d ' ')
allowlist_count=$(wc -l <"$allowlist_types_file" | tr -d ' ')

echo "Subsystem registry check"
echo "  Registered types: ${registered_count}"
echo "  Referenced in subsystem tests: ${subsystem_count}"
echo "  Allowlisted: ${allowlist_count}"
echo

declare -a uncovered=()
declare -a allowlisted_missing=()
declare -a blocking_missing=()
declare -a new_uncovered=()

while IFS= read -r type_name; do
  [ -z "$type_name" ] && continue
  if is_listed_in_file "$type_name" "$subsystem_types_file"; then
    continue
  fi

  uncovered+=("$type_name")
  if is_listed_in_file "$type_name" "$allowlist_types_file"; then
    allowlisted_missing+=("$type_name")
  else
    blocking_missing+=("$type_name")
  fi

  if [ -s "$new_types_file" ] && is_listed_in_file "$type_name" "$new_types_file"; then
    new_uncovered+=("$type_name")
  fi
done <"$registered_types_file"

if [ "${#uncovered[@]}" -gt 0 ]; then
  echo "Types without subsystem test reference:"
  for type_name in "${uncovered[@]}"; do
    if is_listed_in_file "$type_name" "$allowlist_types_file"; then
      echo "  - ${type_name} (allowlisted)"
    else
      echo "  - ${type_name}"
    fi
  done
  echo
fi

if [ -s "$new_types_file" ]; then
  echo "New or changed types in branch (vs ${base_ref}):"
  while IFS= read -r type_name; do
    [ -z "$type_name" ] && continue
    echo "  - ${type_name}"
  done <"$new_types_file"
  echo
fi

exit_code=0

if [ "${#new_uncovered[@]}" -gt 0 ]; then
  echo "ERROR: New provider types must include a subsystem test (resource \"...\" or data \"...\" under subsystem/):"
  for type_name in "${new_uncovered[@]}"; do
    echo "  - ${type_name}"
  done
  echo
  exit_code=1
fi

if [ "${#blocking_missing[@]}" -gt 0 ]; then
  echo "ERROR: Registered types missing subsystem coverage (not allowlisted):"
  for type_name in "${blocking_missing[@]}"; do
    echo "  - ${type_name}"
  done
  echo "Add a subsystem test or an entry in ${allowlist_file} with ticket and reason."
  echo
  exit_code=1
fi

if [ "$exit_code" -eq 0 ]; then
  if [ "${#allowlisted_missing[@]}" -gt 0 ]; then
    echo "OK: All uncovered types are allowlisted (${#allowlisted_missing[@]} pending subsystem tests)."
  else
    echo "OK: All registered types are referenced in subsystem tests."
  fi
fi

exit "$exit_code"
