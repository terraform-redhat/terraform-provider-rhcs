#!/usr/bin/env bash
# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

# This script adds Apache 2.0 license headers to source files
# Usage:
#   ./add-license-header.sh          # Add headers to files
#   ./add-license-header.sh -check   # Check for missing headers (returns non-zero if headers are missing)

set -euo pipefail

ADDLICENSE="${ADDLICENSE:-addlicense}"

# Parse arguments
CHECK_MODE=""
if [[ $# -gt 0 ]]; then
    if [[ "$1" == "-check" ]]; then
        CHECK_MODE="-check"
    else
        echo "Usage: $0 [-check]" >&2
        exit 1
    fi
fi

# Run addlicense
"${ADDLICENSE}" ${CHECK_MODE} \
    -c "Red Hat" \
    -l apache \
    -y "" \
    -s=only \
    -ignore "**/*.md" \
    -ignore "**/*.yaml" \
    -ignore "**/*.yml" \
    -ignore "**/*.toml" \
    -ignore "**/Dockerfile" \
    -ignore "**/*.Dockerfile" \
    -ignore "**/vendor/**" \
    -ignore "**/mock_*.go" \
    -ignore "examples/**" \
    .
