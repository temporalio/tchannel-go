#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
BUILD_DIR="${SCRIPT_DIR}/../build/thrift" && mkdir -p "${BUILD_DIR}"
OS_TYPE="$(uname -s | tr '[:upper:]' '[:lower:]')"

SHA256_CMD="sha256sum --check"
if [[ ${OS_TYPE} == 'darwin'* ]]; then
  SHA256_CMD="shasum -U -a 256 -c"
fi

# Thrift releases found at https://thrift.apache.org/download.
# To update this, you need to update the version, SHA256, and then run install-thrift.sh -c
THRIFT_VERSION="0.16.0"
THRIFT_SHA256="f460b5c1ca30d8918ff95ea3eb6291b3951cf518553566088f3f2be8981f6209"
THRIFT_TAR="thrift-${THRIFT_VERSION}.tar.gz"
THRIFT_URL="https://dlcdn.apache.org/thrift/${THRIFT_VERSION}/${THRIFT_TAR}"

THRIFT_SRC="${BUILD_DIR}/src" && mkdir -p "${THRIFT_SRC}"
THRIFT_BIN_DIR="${SCRIPT_DIR}/bin" && mkdir -p "${THRIFT_BIN_DIR}"
THRIFT_BUILD_LINUX="${BUILD_DIR}/build/linux" && mkdir -p "${THRIFT_BUILD_LINUX}"
THRIFT_BUILD_OSX="${BUILD_DIR}/build/osx" && mkdir -p "${THRIFT_BUILD_OSX}"

function usage() {
  cat <<EOF
Usage: $(basename "${BASH_SOURCE[0]}") [-c] [-h] [-o install_directory]

Script to either install Thrift, to prepare for building tchannel-go, or to re-build the Thrift compiler.
Set the GOOS env to Linux or Darwin to cross-compile, if needed.

Available options:
  -h                    Print this help and exit
  -c                    Recompile the Thrift compiler, recommended to run on OSX with Docker to rebuild all the binaries.
  -o install_directory  Copy the Thrift compiler binary to install_directory/thrift

EOF
  exit 1
}

function fetch_src() {
  if [ ! -f "${BUILD_DIR}/${THRIFT_TAR}" ]; then
    curl --silent "${THRIFT_URL}" --output "${BUILD_DIR}/${THRIFT_TAR}"
    echo "${THRIFT_SHA256}  ${BUILD_DIR}/${THRIFT_TAR}" | ${SHA256_CMD}
  fi
}

function prepare_src() {
  if [ -d "${THRIFT_SRC}" ]; then
     rm -r "${THRIFT_SRC}"
  fi
  mkdir -p "${THRIFT_SRC}"
  tar -xf "${BUILD_DIR}/${THRIFT_TAR}" --strip-components=1 -C "${THRIFT_SRC}"
}

function compile_linux() {
  rm -r "${THRIFT_BUILD_LINUX}"
  prepare_src
  pushd "${THRIFT_SRC}" &> /dev/null

  IMAGE_NAME="tchannel-go/thrift-build"
  DOCKERFILE="thrift.Dockerfile"

  docker build -f "${SCRIPT_DIR}/${DOCKERFILE}" -t "${IMAGE_NAME}" "${SCRIPT_DIR}"
  docker run -it -v "${THRIFT_SRC}":/thrift/src -v "${THRIFT_BUILD_LINUX}":/thrift/build "${IMAGE_NAME}" /thrift/build.sh

  cp -f "${THRIFT_BUILD_LINUX}"/bin/thrift "${THRIFT_BIN_DIR}"/thrift-linux

  popd &> /dev/null
}

function compile_osx() {
  rm -r "${THRIFT_BUILD_OSX}"
  prepare_src
  pushd "${THRIFT_SRC}" &> /dev/null

  OSX_DEPENDENCIES="flex bison cmake"
  for dep in ${OSX_DEPENDENCIES}; do
    which "${dep}" &> /dev/null || brew install "${dep}"
  done

  THRIFT_SRC="${THRIFT_SRC}" THRIFT_BUILD="${THRIFT_BUILD_OSX}" "${SCRIPT_DIR}"/build-thrift.sh
  cp -f "${THRIFT_BUILD_OSX}"/bin/Release/thrift "${THRIFT_BIN_DIR}"/thrift-osx

  popd &> /dev/null
}

function install_thrift() {
  # To support cross-compiling, we use the GOOS which the Makefile sets
  case "${GOOS:-$OS_TYPE}" in
    darwin*)
      THRIFT_BIN="${SCRIPT_DIR}/bin/thrift-osx"
      ;;
    linux*)
      THRIFT_BIN="${SCRIPT_DIR}/bin/thrift-linux"
      ;;
    *)
      echo "Unsupported OS type ${OS_TYPE}"
      exit 1
      ;;
  esac

  mkdir -p "${1}"
  cp "${THRIFT_BIN}" "${1}"/thrift
}

INSTALL_DIR=""
while getopts 'co:h' opt; do
  case "$opt" in
    c)
      fetch_src
      compile_linux
      if [ "${OS_TYPE}" == "darwin" ]; then
        compile_osx
      fi
      ;;
    o)
      if [ -z "${OPTARG}" ]; then
        echo "No install directory was provided"
        usage
      fi
      INSTALL_DIR=${OPTARG}
      ;;
    ?)
      usage
      ;;
  esac
done
shift "$((${OPTIND} -1))"

if [ -n "${INSTALL_DIR}" ]; then
  install_thrift "${INSTALL_DIR}"
fi
