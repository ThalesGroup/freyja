#!/usr/bin/env bash

SCRIPT_PATH="$(realpath "$0")"
PROJECT_DIR="$(dirname "${SCRIPT_PATH:?}")"
DIST_DIR="${PROJECT_DIR:?}/dist"
BIN_NAME="freyja"
GO_MAIN="${PROJECT_DIR:?}/cmd/freyja/main.go"

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

#check_requirement go version "Install Go first"

echo "Build freyja"

cd "${PROJECT_DIR}" || exit
mkdir -p "${DIST_DIR:?}"
go build -o "${DIST_DIR:?}/${BIN_NAME}" "${GO_MAIN:?}"

echo "Built in ${DIST_DIR:?}/${BIN_NAME}"

exit 0
