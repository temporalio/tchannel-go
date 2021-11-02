#!/bin/bash

set -euo pipefail

if [ -z "${1}" ]; then
  echo "usage: ${0} installDirPath" >&2
  exit 1
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
OS_TYPE="$(uname -s)"

case "${OS_TYPE}" in
  Darwin*) 
    THRIFT_BIN="${SCRIPT_DIR}/bin/thrift-osx"
    ;;
  Linux*) 
    THRIFT_BIN="${SCRIPT_DIR}/bin/thrift-linux"
    ;;
  *)
    echo "Unsupported OS type ${OS_TYPE}"
    exit 1
    ;;
esac

mkdir -p "${1}"
cp ${THRIFT_BIN} ${1}/thrift

