#!/usr/bin/env bash
set -e
PROGRAM_NAME="micro-ci"
ARCHIVE_DIR="$(mktemp -d)"
OUTPUT_DIR="${ARCHIVE_DIR}/${PROGRAM_NAME}"
OUTPUT_FILE="${OUTPUT_DIR}/${PROGRAM_NAME}"

trap "rm -rf ${OUTPUT_DIR}" EXIT SIGINT SIGTERM
go build -o "${OUTPUT_FILE}" .
cp install.sh "${OUTPUT_DIR}/install.sh"
cp uninstall.sh "${OUTPUT_DIR}/uninstall.sh"
tar -czf "${PROGRAM_NAME}.tar.gz" -C "${ARCHIVE_DIR}" .
