#!/usr/bin/env bash

SCRIPT_PATH="$(realpath "$0")"
PROJECT_DIR="$(dirname "${SCRIPT_PATH:?}")"

#
# FUNCTIONS
#

function check_requirement(){
    if ! eval "$@" >> /dev/null 2>&1 ; then
        echo "! Fatal : missing requirement"
        if [ -n "${*: -1}" ]; then echo "${@: -1}"; fi
        exit 1
    fi
}

#
# MAIN
#

check_requirement poetry --version "Install poetry first"

cd "${PROJECT_DIR}" || exit
echo "Build environment :"
poetry env info

echo ""
echo "Update dependencies"
poetry check
poetry update --without dev
poetry install --without dev --sync

echo "Build Freyja"
rm -rf "${PROJECT_DIR}/dist"
poetry build --format wheel

echo "Built in dist/:"
ls "./dist/"