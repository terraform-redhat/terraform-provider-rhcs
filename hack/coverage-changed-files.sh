#!/usr/bin/env bash
# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0
#
# Optional local diagnostic: changed-line unit coverage for provider/ and internal/
# (gocovdiff vs merge base with main). Not run by make pre-push-checks or CI.
# Subsystem tests do not count toward the score. Invoke via: make coverage-changed-files

set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=lib/git.sh
source "${script_dir}/lib/git.sh"

repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

required_coverage_percent="80"
gocovdiff_module="github.com/vearutop/gocovdiff"
gocovdiff_version="v1.4.2"
coverage_base_ref="${COVERAGE_BASE_REF:-}"

tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/rhcs-changed-coverage-XXXXXX")
coverage_profile="$tmp_dir/cover.out"
diff_file="$tmp_dir/changes.diff"
delta_file="$tmp_dir/delta-cov.txt"
trap 'rm -rf "$tmp_dir"' EXIT

resolve_diff_args() {
  local diff_range

  diff_range=$(resolve_changed_go_diff_range "$coverage_base_ref" || true)
  if [ -n "$diff_range" ]; then
    mapfile -t merge_range_go_files < <(git diff "$diff_range" --name-only --diff-filter=ACMR -- '*.go')
    if [ "${#merge_range_go_files[@]}" -gt 0 ]; then
      diff_base_args=("$diff_range")
      return
    fi
  fi

  diff_base_args=(--cached)
  mapfile -t candidate_files < <(git diff "${diff_base_args[@]}" --name-only --diff-filter=ACMR -- '*.go')
  if [ "${#candidate_files[@]}" -eq 0 ]; then
    diff_base_args=()
  fi
}

fail_if_ci_skipped_required_coverage() {
  local diff_range provider_changes

  if ! ci_pr_context; then
    return 0
  fi

  diff_range=$(resolve_changed_go_diff_range "$coverage_base_ref" || true)
  if [ -z "$diff_range" ]; then
    echo "ERROR: changed-files coverage could not resolve a diff range in CI" >&2
    echo "Set PULL_BASE_SHA/PULL_PULL_SHA or fetch origin/main before running checks." >&2
    exit 1
  fi

  provider_changes=$(count_provider_internal_go_changes "$diff_range")
  if [ "$provider_changes" -gt 0 ]; then
    echo "ERROR: changed-files coverage did not run for ${provider_changes} provider/internal Go file(s) in CI" >&2
    echo "Diff range: ${diff_range}" >&2
    exit 1
  fi
}

declare -a diff_base_args=()
resolve_diff_args

mapfile -t candidate_files < <(git diff "${diff_base_args[@]}" --name-only --diff-filter=ACMR -- '*.go')

declare -A package_seen=()
declare -a changed_packages=()
declare -a coverage_candidate_files=()
for file_path in "${candidate_files[@]}"; do
  [ -z "$file_path" ] && continue
  case "$file_path" in
    provider/*|internal/*)
      ;;
    *)
      continue
      ;;
  esac

  case "$file_path" in
    vendor/*|.tmp/*|*_test.go)
      continue
      ;;
  esac

  [ -f "$file_path" ] || continue
  coverage_candidate_files+=("$file_path")
  package_name=$(go list "./$(dirname "$file_path")" 2>/dev/null || true)
  if [ -n "$package_name" ] && [ -z "${package_seen[$package_name]+x}" ]; then
    package_seen["$package_name"]=1
    changed_packages+=("$package_name")
  fi
done

if [ "${#coverage_candidate_files[@]}" -eq 0 ]; then
  fail_if_ci_skipped_required_coverage
  exit 0
fi

git diff "${diff_base_args[@]}" -U0 -- "${coverage_candidate_files[@]}" > "$diff_file"
if [ ! -s "$diff_file" ]; then
  fail_if_ci_skipped_required_coverage
  exit 0
fi

if [ "${#changed_packages[@]}" -eq 0 ]; then
  fail_if_ci_skipped_required_coverage
  exit 0
fi

go test -vet=off -count=1 -covermode=atomic -coverprofile="$coverage_profile" "${changed_packages[@]}"

GOFLAGS='-mod=mod' go run "${gocovdiff_module}@${gocovdiff_version}" \
  -diff "$diff_file" \
  -cov "$coverage_profile" \
  -exclude "vendor/,.tmp/" \
  -target-delta-cov "$required_coverage_percent" \
  -delta-cov-file "$delta_file"

if [ -s "$delta_file" ]; then
  cat "$delta_file"
  echo
fi

if grep -q "coverage is less than" "$delta_file"; then
  exit 1
fi
