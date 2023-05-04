#!/usr/bin/env bash

SCRIPT_PATH="$(realpath "$0")"
PROJECT_DIR="$(dirname "${SCRIPT_PATH:?}")"

#
# FUNCTIONS
#

function check_requirement(){
    if ! [ -x "$(command -v "$1")" ]; then
        echo "! Fatal : missing requirement"
        if [ -n "${*: -1}" ]; then echo "${@: -1}"; fi
        exit 1
    fi
}

#
# MAIN
#

echo ". Check requirements"
cd "${PROJECT_DIR}" || exit
check_requirement npm version "Install npm first"

echo ". Install dependencies"
cd pages || exit
npm install --force

echo ". Build website"
npm run build

echo ". Delete old website"
cd "${PROJECT_DIR}" || exit
rm -rf ./docs

echo ""
echo ". Move website sources to Github pages' publish dir"
mv pages/build ./docs

echo ". Done"
exit 0
