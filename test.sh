#!/usr/bin/env bash

SCRIPT_PATH="$(realpath "$0")"
PROJECT_DIR="$(dirname "${SCRIPT_PATH:?}")"
TEST_DIR="${PROJECT_DIR}/freyja/tests"
FREYJA_WORKSPACE="${HOME}/freyja-workspace"
TEST_RESOURCES_DIR="${TEST_DIR}/resources"

cd "${PROJECT_DIR}/freyja" || exit

#
# RUNTIME
#

# freyja-testing
echo ". Testing full configuration"
poetry run python "${PROJECT_DIR}/freyja/__main__.py" machine create -c "${TEST_RESOURCES_DIR}/testing.yaml" --foreground
#rm -rf "${FREYJA_WORKSPACE}/build/freyja-testing"

echo ""
echo "OK"
