#!/usr/bin/env bash
# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

# verify_commit ref
# Returns 0 when ref resolves to a commit.
verify_commit() {
  git rev-parse --verify "${1}^{commit}" >/dev/null 2>&1
}

# ci_pr_context
# True when running in automated CI for a pull request.
ci_pr_context() {
  if [ "${GITHUB_EVENT_NAME:-}" = "pull_request" ]; then
    return 0
  fi
  if [ -n "${OPENSHIFT_CI:-}" ] && [ "${JOB_TYPE:-}" = "presubmit" ]; then
    return 0
  fi
  if [ -n "${PULL_BASE_SHA:-${BASE_SHA:-}}" ] && [ -n "${PULL_PULL_SHA:-${HEAD_SHA:-}}" ]; then
    return 0
  fi
  return 1
}

# newest_merge_base HEAD ref...
# Prints the newest merge-base among merge-base(HEAD, ref) for each reachable ref.
newest_merge_base() {
  local head=$1
  shift

  local best_mb=""
  local ref mb

  for ref in "$@"; do
    [ -z "$ref" ] && continue
    if ! verify_commit "$ref"; then
      continue
    fi
    if ! mb=$(git merge-base "$head" "${ref}" 2>/dev/null); then
      continue
    fi
    if [ -z "$best_mb" ]; then
      best_mb=$mb
      continue
    fi
    if git merge-base --is-ancestor "$best_mb" "$mb" 2>/dev/null; then
      best_mb=$mb
    fi
  done

  printf '%s' "$best_mb"
}

# merge_base_range [HEAD] ref...
# Prints merge_base...HEAD when a merge base exists.
merge_base_range() {
  local head=${1:-HEAD}
  shift

  local mb
  mb=$(newest_merge_base "$head" "$@")
  if [ -n "$mb" ]; then
    printf '%s...%s' "$mb" "$head"
  fi
}

# resolve_changed_go_diff_range [explicit_base_ref]
# Prints a git diff range (base...head) for PR changed-file checks.
# Priority: PULL_BASE_SHA...PULL_PULL_SHA, explicit base ref merge range,
# then merge range against origin/main, upstream/main, or main.
resolve_changed_go_diff_range() {
  local explicit_base_ref=${1:-}
  local pull_base pull_head diff_range

  pull_base="${PULL_BASE_SHA:-${BASE_SHA:-}}"
  pull_head="${PULL_PULL_SHA:-${HEAD_SHA:-}}"

  if [ -n "$pull_base" ] && [ -n "$pull_head" ] \
     && verify_commit "$pull_base" && verify_commit "$pull_head"; then
    printf '%s...%s' "$pull_base" "$pull_head"
    return 0
  fi

  if [ -n "$explicit_base_ref" ] && verify_commit "$explicit_base_ref"; then
    diff_range=$(merge_base_range HEAD "$explicit_base_ref")
    if [ -n "$diff_range" ]; then
      printf '%s' "$diff_range"
      return 0
    fi
  fi

  diff_range=$(merge_base_range HEAD origin/main upstream/main main)
  if [ -n "$diff_range" ]; then
    printf '%s' "$diff_range"
    return 0
  fi

  return 1
}

# count_provider_internal_go_changes diff_range
# Prints the number of changed non-test Go files under provider/ or internal/.
count_provider_internal_go_changes() {
  local diff_range=$1
  local file_path

  if [ -z "$diff_range" ]; then
    printf '0'
    return 0
  fi

  local count=0
  while IFS= read -r file_path; do
    [ -z "$file_path" ] && continue
    case "$file_path" in
      provider/*|internal/*)
        ;;
      *)
        continue
        ;;
    esac
    case "$file_path" in
      *_test.go)
        continue
        ;;
    esac
    count=$((count + 1))
  done < <(git diff "$diff_range" --name-only --diff-filter=ACMR -- '*.go')

  printf '%s' "$count"
}
