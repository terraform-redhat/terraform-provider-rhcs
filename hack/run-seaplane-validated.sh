#!/usr/bin/env bash
# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0
#
# Run seaplane-validated e2e tests across all seaplane cluster profiles
# with deduplication: a test executed on one profile is skipped on later
# profiles. This drives a single `make e2e-seaplane-validated` command to
# cover all profile-specific behavior without redundant runs.
#
# Deduplication key: Ginkgo spec FullText
#   FullText = strings.Join(append(ContainerHierarchyTexts, LeafNodeText), " ")
# Specs that Skip (profile gate not satisfied) are NOT recorded for dedup, so a
# profile-specific test will run on the first profile where it is applicable.
# Skipped seaplane-validated specs are recorded separately in
# tests/rhcs_output/seaplane-validated-skipped.txt for visibility without
# affecting dedup (dedup still uses only passed/failed).
#
# Requirements: jq (or python3 fallback), yq (optional, falls back to grep/sed),
#               GINKGO binary at ./bin/ginkgo or $GINKGO_BIN

set -euo pipefail

REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
cd "$REPO_ROOT"

PROFILES_FILE="${PROFILES_FILE:-tests/ci/profiles/tf_seaplane_cluster_profiles.yml}"
LABEL_FILTER="${LABEL_FILTER:-seaplane-validated}"
GINKGO_BIN="${GINKGO_BIN:-./bin/ginkgo}"
REPORT_DIR="${REPORT_DIR:-tests/rhcs_output}"
DEDUP_FILE="${DEDUP_FILE:-$REPORT_DIR/seaplane-validated-dedup.txt}"
SKIPPED_FILE="${SKIPPED_FILE:-$REPORT_DIR/seaplane-validated-skipped.txt}"
SEAPLANE_PROFILES_ENV="${SEAPLANE_PROFILES:-}"

mkdir -p "$REPORT_DIR"

GOCACHE_DIR="${GOCACHE:-/private/tmp/ginkgo-cache}"
export GOCACHE="$GOCACHE_DIR"
export GOTMPDIR="${GOTMPDIR:-/private/tmp}"

if [ ! -x "$GINKGO_BIN" ]; then
  echo "Ginkgo binary not found at $GINKGO_BIN, building..." >&2
  make --no-print-directory "$(basename "$GINKGO_BIN")" 2>&1 | tail -5
fi

if [ "${DRY_RUN:-}" != "true" ] && [ "${DRY_RUN:-}" != "1" ]; then
  echo "Installing provider for e2e tests..." >&2
  make --no-print-directory install 2>&1 | tail -20
fi

if [ ! -f "$PROFILES_FILE" ]; then
  echo "Profiles file not found: $PROFILES_FILE" >&2
  exit 1
fi

# Validate required Seaplane environment unless in dry-run.
if [ "${DRY_RUN:-}" != "true" ] && [ "${DRY_RUN:-}" != "1" ]; then
  if [ -z "${SEAPLANE_URL:-}" ]; then
    echo "ERROR: SEAPLANE_URL must be set to run seaplane-validated tests" >&2
    echo "  Example: SEAPLANE_URL=https://<api-id>.execute-api.<region>.amazonaws.com/prod" >&2
    echo "  The region is derived from the URL hostname; REGION override is optional." >&2
    exit 1
  fi
  if [ -z "${SEAPLANE_RELEASE_IMAGE:-}" ]; then
    echo "ERROR: SEAPLANE_RELEASE_IMAGE must be set to run seaplane-validated tests" >&2
    echo "  This is the HostedCluster release image pullspec, not an OCP version label." >&2
    echo "  Examples:" >&2
    echo "    quay.io/openshift-release-dev/ocp-release@sha256:<digest>" >&2
    echo "    registry.ci.openshift.org/ocp/release:4.19.0-0.nightly-... " >&2
    echo "  Retrieve from:" >&2
    echo "    - Rosa CLI: rosa list versions --hosted-cp" >&2
    echo "    - Release controller: https://amd64.ocp.releases.ci.openshift.org/" >&2
    echo "    - Previous CI run artifacts or ask the Seaplane platform team." >&2
    echo "  Set with: export SEAPLANE_RELEASE_IMAGE=<pullspec>" >&2
    exit 1
  fi
fi

escape_regex() {
  printf '%s' "$1" | sed 's/[][\\.*^$(){}|+?]/\\&/g'
}

get_profiles() {
  if [ -n "$SEAPLANE_PROFILES_ENV" ]; then
    for p in $SEAPLANE_PROFILES_ENV; do
      printf '%s\n' "$p"
    done
    return
  fi
  if command -v yq >/dev/null 2>&1; then
    if yq eval '.profiles[].as' "$PROFILES_FILE" 2>/dev/null | grep -v '^null$' | grep -q .; then
      yq eval '.profiles[].as' "$PROFILES_FILE" 2>/dev/null | grep -v '^null$'
      return
    fi
  fi
  grep -E '^\s*-\s+as:' "$PROFILES_FILE" | sed -E 's/.*as:\s*//; s/\s*#.*//; s/^[[:space:]]*//; s/[[:space:]]*$//' | grep -v '^$'
}

extract_executed_tests() {
  local report_file="$1"
  local out_file="$2"
  if [ ! -f "$report_file" ]; then
    return 0
  fi
  if command -v jq >/dev/null 2>&1; then
    jq -r '
      .[].SpecReports[]
      | select(.LeafNodeType == "It")
      | select(.State == "passed" or .State == "failed")
      | ((.ContainerHierarchyTexts // []) | join(" ")) + " " + .LeafNodeText
      | gsub("^ +"; "") | gsub(" +$"; "")
    ' "$report_file" 2>/dev/null >> "$out_file" || true
  elif command -v python3 >/dev/null 2>&1; then
    python3 -c "
import json, sys
try:
    data=json.load(open(sys.argv[1]))
    for suite in data:
        for spec in suite.get('SpecReports') or []:
            if spec.get('LeafNodeType')!='It':
                continue
            if spec.get('State') not in ('passed','failed'):
                continue
            texts=spec.get('ContainerHierarchyTexts') or []
            leaf=spec.get('LeafNodeText') or ''
            full=' '.join(texts + [leaf]).strip()
            if full:
                print(full)
except Exception as e:
    pass
" "$report_file" >> "$out_file" 2>/dev/null || true
  else
    echo "WARNING: jq and python3 not found, cannot parse deduplication report; dedup disabled for this profile" >&2
  fi
}

extract_skipped_tests() {
  local report_file="$1"
  local out_file="$2"
  if [ ! -f "$report_file" ]; then
    return 0
  fi
  if command -v python3 >/dev/null 2>&1; then
    python3 -c "
import json, sys
try:
    data=json.load(open(sys.argv[1]))
    for suite in data:
        for spec in suite.get('SpecReports') or []:
            if spec.get('LeafNodeType')!='It':
                continue
            if spec.get('State') != 'skipped':
                continue
            labs=[]
            for lst in (spec.get('ContainerHierarchyLabels') or []):
                if lst:
                    labs.extend(lst)
            labs.extend(spec.get('LeafNodeLabels') or [])
            if not any('seaplane-validated' in str(x) for x in labs):
                continue
            texts=spec.get('ContainerHierarchyTexts') or []
            leaf=spec.get('LeafNodeText') or ''
            full=' '.join(texts + [leaf]).strip()
            if full:
                print(full)
except Exception as e:
    pass
" "$report_file" >> "$out_file" 2>/dev/null || true
  elif command -v jq >/dev/null 2>&1; then
    jq -r '
      .[].SpecReports[]
      | select(.LeafNodeType == "It")
      | select(.State == "skipped")
      | select(
          ([.ContainerHierarchyLabels[][] // empty] + (.LeafNodeLabels // []))
          | map(tostring) | any(test("seaplane-validated"))
        )
      | ((.ContainerHierarchyTexts // []) | join(" ")) + " " + .LeafNodeText
      | gsub("^ +"; "") | gsub(" +$"; "")
    ' "$report_file" 2>/dev/null >> "$out_file" || true
  fi
}

cluster_exists_for_profile() {
  local profile="$1"
  local manifests_dir="tests/tf-manifests/rhcs/clusters/seaplane"
  local output_json
  output_json=$(TF_WORKSPACE="$profile" terraform -chdir="$manifests_dir" output -json 2>/dev/null || echo "{}")
  local cluster_id=""
  local cluster_phase=""
  if command -v jq >/dev/null 2>&1; then
    cluster_id=$(printf '%s' "$output_json" | jq -r '.cluster_id.value // empty' 2>/dev/null || echo "")
    cluster_phase=$(printf '%s' "$output_json" | jq -r '.cluster_phase.value // empty' 2>/dev/null || echo "")
  else
    cluster_id=$(printf '%s' "$output_json" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('cluster_id',{}).get('value','') or '')" 2>/dev/null || echo "")
    cluster_phase=$(printf '%s' "$output_json" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('cluster_phase',{}).get('value','') or '')" 2>/dev/null || echo "")
  fi
  if [ -n "$cluster_id" ] && [ "$cluster_id" != "null" ] && [ "$cluster_id" != "" ]; then
    if [ -n "$cluster_phase" ] && [ "$cluster_phase" != "null" ] && [ "$cluster_phase" != "" ] && [ "$cluster_phase" != "Ready" ]; then
      echo "  [cluster-check] WARNING: cluster $profile exists but phase is '$cluster_phase' (not Ready), treating as not ready" >&2
      return 1
    fi
    return 0
  fi
  return 1
}

pending_specs_count() {
  local label_filter="$1"
  local skip_regex="$2"
  local tmp_report
  tmp_report=$(mktemp)
  trap 'rm -f "$tmp_report"' RETURN
  local args=(
    run
    --label-filter="$label_filter"
    --timeout 5h
    -r
    --json-report="$tmp_report"
    --dry-run
  )
  if [ -n "$skip_regex" ]; then
    args+=(--skip="$skip_regex")
  fi
  args+=(tests/e2e)
  set +e
  "$GINKGO_BIN" "${args[@]}" >/dev/null 2>&1
  local exit_code=$?
  set -e
  if [ ! -f "$tmp_report" ]; then
    echo 0
    return
  fi
  local count=0
  if command -v jq >/dev/null 2>&1; then
    count=$(jq '[.[].PreRunStats.SpecsThatWillRun] | add // 0' "$tmp_report" 2>/dev/null || echo 0)
  elif command -v python3 >/dev/null 2>&1; then
    count=$(python3 -c "
import json
data=json.load(open('$tmp_report'))
print(sum(s.get('PreRunStats',{}).get('SpecsThatWillRun',0) for s in data))
" 2>/dev/null || echo 0)
  fi
  echo "${count:-0}"
}

ensure_cluster_for_profile() {
  local profile="$1"
  if [ "${DRY_RUN:-}" = "true" ] || [ "${DRY_RUN:-}" = "1" ]; then
    echo "  [ensure-cluster] DRY_RUN: would ensure cluster exists for profile $profile" >&2
    return 0
  fi
  if cluster_exists_for_profile "$profile"; then
    echo "  [ensure-cluster] cluster already exists for profile $profile" >&2
    return 0
  fi
  echo "  [ensure-cluster] no cluster found for profile $profile, creating one..." >&2
  echo "    CLUSTER_PROFILE=$profile SEAPLANE_URL and SEAPLANE_RELEASE_IMAGE must be set" >&2
  if [ -z "${SEAPLANE_URL:-}" ] || [ -z "${SEAPLANE_RELEASE_IMAGE:-}" ]; then
    echo "    ERROR: cannot create cluster without SEAPLANE_URL and SEAPLANE_RELEASE_IMAGE" >&2
    return 1
  fi
  set +e
  CLUSTER_PROFILE="$profile" go run ./hack/ensure-seaplane-cluster.go
  local rc=$?
  set -e
  if [ $rc -ne 0 ]; then
    echo "  [ensure-cluster] failed to create cluster for $profile (exit $rc)" >&2
    return $rc
  fi
  echo "  [ensure-cluster] cluster ensured for $profile" >&2
  return 0
}

: > "$DEDUP_FILE"
: > "$SKIPPED_FILE"
DEDUP_TMP=$(mktemp)
SKIPPED_TMP=$(mktemp)
trap 'rm -f "$DEDUP_TMP" "$SKIPPED_TMP"' EXIT

PROFILE_LIST=$(get_profiles)
if [ -z "$PROFILE_LIST" ]; then
  echo "No seaplane profiles found in $PROFILES_FILE" >&2
  exit 1
fi

PROFILE_COUNT=$(printf '%s\n' "$PROFILE_LIST" | wc -l | tr -d ' ')
echo "Found $PROFILE_COUNT seaplane profiles to cover with label filter '$LABEL_FILTER':" >&2
printf '%s\n' "$PROFILE_LIST" | sed 's/^/  - /' >&2
echo "" >&2

TOTAL_PROFILES=0
SKIPPED_PROFILES=0
FAILED_PROFILES=0

# Phases ensure intra-profile dependencies: day1 creates, day1-post verifies.
# Each phase is a ginkgo invocation with a phase-specific label filter; dedup is
# maintained across all phases and profiles so a spec executed once is never
# duplicated in a later phase/profile.
PHASES=(
  "day1-prepare"
  "day1"
  "day1-post"
  "day2"
  "day3"
  "destroy"
  "other"
)

for PROFILE in $PROFILE_LIST; do
  TOTAL_PROFILES=$((TOTAL_PROFILES+1))
  echo "======================================================================" >&2
  echo "[$TOTAL_PROFILES/$PROFILE_COUNT] Running '$LABEL_FILTER' tests for profile: $PROFILE" >&2
  echo "======================================================================" >&2

  export CLUSTER_PROFILE="$PROFILE"

  PROFILE_HAS_FAILURE=0
  for PHASE in "${PHASES[@]}"; do
    PHASE_LABEL_FILTER=""
    case "$PHASE" in
      day1-prepare) PHASE_LABEL_FILTER="$LABEL_FILTER && day1-prepare" ;;
      day1) PHASE_LABEL_FILTER="$LABEL_FILTER && day1 && !day1-prepare && !day1-post" ;;
      day1-post) PHASE_LABEL_FILTER="$LABEL_FILTER && day1-post" ;;
      day2) PHASE_LABEL_FILTER="$LABEL_FILTER && day2" ;;
      day3) PHASE_LABEL_FILTER="$LABEL_FILTER && day3" ;;
      destroy) PHASE_LABEL_FILTER="$LABEL_FILTER && destroy" ;;
      other) PHASE_LABEL_FILTER="$LABEL_FILTER && !day1 && !day1-prepare && !day1-post && !day2 && !day3 && !destroy" ;;
    esac

    SKIP_REGEX=""
    if [ -s "$DEDUP_FILE" ]; then
      ESCAPED=""
      while IFS= read -r line || [ -n "$line" ]; do
        [ -z "$line" ] && continue
        esc=$(escape_regex "$line")
        if [ -z "$ESCAPED" ]; then
          ESCAPED="$esc"
        else
          ESCAPED="$ESCAPED|$esc"
        fi
      done < "$DEDUP_FILE"
      if [ -n "$ESCAPED" ]; then
        SKIP_REGEX="($ESCAPED)"
      fi
    fi

    if [ -n "$SKIP_REGEX" ]; then
      DEDUP_COUNT=$(wc -l < "$DEDUP_FILE" | tr -d ' ')
      echo "  [$PHASE] Dedup: skipping $DEDUP_COUNT already-executed specs" >&2
    else
      echo "  [$PHASE] Dedup: no previously executed specs" >&2
    fi

    PENDING_COUNT=$(pending_specs_count "$PHASE_LABEL_FILTER" "$SKIP_REGEX")
    if [ "$PENDING_COUNT" -eq 0 ] 2>/dev/null; then
      echo "  [$PHASE] No pending specs after dedup (0), skipping ginkgo run" >&2
      continue
    fi
    echo "  [$PHASE] $PENDING_COUNT specs pending (after dedup)" >&2

    # Ensure cluster exists before phases that require an existing cluster.
    case "$PHASE" in
      day1-post|day2|day3|other|destroy)
        if [ "${DRY_RUN:-}" != "true" ] && [ "${DRY_RUN:-}" != "1" ]; then
          if ! cluster_exists_for_profile "$PROFILE"; then
            echo "  [$PHASE] cluster required for $PENDING_COUNT pending specs but not found in workspace $PROFILE" >&2
            if ! ensure_cluster_for_profile "$PROFILE"; then
              echo "  [$PHASE] failed to ensure cluster for $PROFILE, marking profile as failed" >&2
              PROFILE_HAS_FAILURE=1
              continue
            fi
          else
            echo "  [$PHASE] cluster exists for $PROFILE, proceeding" >&2
          fi
        fi
        ;;
    esac

    REPORT_FILE="$REPORT_DIR/seaplane-validated-$PROFILE-$PHASE.json"
    rm -f "$REPORT_FILE"

    EXTRA_ARGS=()
    if [ -n "${GINKGO_EXTRA_ARGS:-}" ]; then
      # shellcheck disable=SC2206
      EXTRA_ARGS=(${GINKGO_EXTRA_ARGS})
    fi
    GINKGO_ARGS=(
      run
      --label-filter="$PHASE_LABEL_FILTER"
      --timeout 5h
      -r
      --json-report="$REPORT_FILE"
    )
    if [ ${#EXTRA_ARGS[@]} -gt 0 ]; then
      GINKGO_ARGS+=("${EXTRA_ARGS[@]}")
    fi
    if [ "${DRY_RUN:-}" = "true" ] || [ "${DRY_RUN:-}" = "1" ]; then
      GINKGO_ARGS+=(--dry-run)
    fi
    if [ -n "$SKIP_REGEX" ]; then
      GINKGO_ARGS+=(--skip="$SKIP_REGEX")
    fi
    GINKGO_ARGS+=(tests/e2e)

    echo "  [$PHASE] Running: $GINKGO_BIN ${GINKGO_ARGS[*]}  (filter: $PHASE_LABEL_FILTER)" >&2
    set +e
    "$GINKGO_BIN" "${GINKGO_ARGS[@]}"
    GINKGO_EXIT=$?
    set -e

    if [ ! -f "$REPORT_FILE" ]; then
      if [ $GINKGO_EXIT -ne 0 ]; then
        echo "  [$PHASE] WARNING: no JSON report at $REPORT_FILE, ginkgo exit $GINKGO_EXIT" >&2
        PROFILE_HAS_FAILURE=1
      else
        echo "  [$PHASE] No specs matched filter '$PHASE_LABEL_FILTER' (skipped)" >&2
      fi
      continue
    fi

    : > "$DEDUP_TMP"
    extract_executed_tests "$REPORT_FILE" "$DEDUP_TMP"

    if [ -s "$DEDUP_TMP" ]; then
      sort -u "$DEDUP_TMP" -o "$DEDUP_TMP"
      NEW_COUNT=$(wc -l < "$DEDUP_TMP" | tr -d ' ')
      cat "$DEDUP_TMP" >> "$DEDUP_FILE"
      sort -u "$DEDUP_FILE" -o "$DEDUP_FILE"
      echo "  [$PHASE] Recorded $NEW_COUNT new specs, total dedup now $(wc -l < "$DEDUP_FILE" | tr -d ' ') specs" >&2
    else
      if command -v jq >/dev/null 2>&1; then
        PASSED_IN_REPORT=$(jq '[.[].SpecReports[] | select(.LeafNodeType=="It" and (.State=="passed" or .State=="failed"))] | length' "$REPORT_FILE" 2>/dev/null || echo 0)
        if [ "$PASSED_IN_REPORT" = "0" ]; then
          echo "  [$PHASE] No new specs executed (all skipped via dedup or profile gates)" >&2
        fi
      fi
    fi

    : > "$SKIPPED_TMP"
    extract_skipped_tests "$REPORT_FILE" "$SKIPPED_TMP"
    if [ -s "$SKIPPED_TMP" ]; then
      sort -u "$SKIPPED_TMP" -o "$SKIPPED_TMP"
      cat "$SKIPPED_TMP" >> "$SKIPPED_FILE"
      sort -u "$SKIPPED_FILE" -o "$SKIPPED_FILE"
    fi

    if [ $GINKGO_EXIT -ne 0 ]; then
      echo "  [$PHASE] Ginkgo exit $GINKGO_EXIT (recorded dedup, continuing to next phase)" >&2
      PROFILE_HAS_FAILURE=1
    fi
  done

  # Merge per-phase reports into one per-profile report for backward compatibility and summary
  MERGED_REPORT="$REPORT_DIR/seaplane-validated-$PROFILE.json"
  if command -v jq >/dev/null 2>&1; then
    jq -s 'add' "$REPORT_DIR/seaplane-validated-$PROFILE-"*.json 2>/dev/null > "$MERGED_REPORT" 2>/dev/null || true
  fi

  if [ $PROFILE_HAS_FAILURE -ne 0 ]; then
    FAILED_PROFILES=$((FAILED_PROFILES+1))
  else
    # If no new specs across all phases, count as skipped profile (all deduped)
    if [ ! -s "$DEDUP_FILE" ]; then
      SKIPPED_PROFILES=$((SKIPPED_PROFILES+1))
    fi
  fi

  echo "" >&2
done

echo "======================================================================" >&2
echo "Seaplane-validated run complete across $TOTAL_PROFILES profiles" >&2
echo "  Total unique specs executed: $(wc -l < "$DEDUP_FILE" 2>/dev/null | tr -d ' ' || echo 0)" >&2
echo "  Total unique specs skipped: $(wc -l < "$SKIPPED_FILE" 2>/dev/null | tr -d ' ' || echo 0)" >&2
echo "  Profiles with no new specs: $SKIPPED_PROFILES" >&2
echo "  Profiles with failures: $FAILED_PROFILES" >&2
echo "  Dedup record: $DEDUP_FILE" >&2
echo "  Skipped record: $SKIPPED_FILE" >&2
echo "  Per-profile reports: $REPORT_DIR/seaplane-validated-*.json" >&2
if [ -s "$DEDUP_FILE" ]; then
  echo "" >&2
  echo "Executed specs (FullText):" >&2
  sed 's/^/  - /' "$DEDUP_FILE" >&2
fi
if [ -s "$SKIPPED_FILE" ]; then
  echo "" >&2
  echo "Skipped specs (FullText):" >&2
  sed 's/^/  - /' "$SKIPPED_FILE" >&2
fi
echo "======================================================================" >&2

if [ $FAILED_PROFILES -gt 0 ]; then
  exit 1
fi
