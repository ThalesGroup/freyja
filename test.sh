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

cd "${PROJECT_DIR}" || exit
pyversion=$(python --version)
echo "Install Freyja for: ${pyversion}"

echo ""
wheel="$(find "./dist/" -type f -name "*.whl")"
python -m pip install --upgrade pip
python -m pip install "${wheel}" --force-reinstall

echo ""
echo "OK"