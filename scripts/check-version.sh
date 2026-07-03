#!/usr/bin/env bash
set -euo pipefail

BASE_REF="${1:-origin/main}"
VERSION_FILE="version.go"

if [[ ! -f "${VERSION_FILE}" ]]; then
  echo "::error::${VERSION_FILE} not found"
  exit 1
fi

version="$(sed -nE 's/^const[[:space:]]+Version[[:space:]]*=[[:space:]]*"([0-9]+\.[0-9]+\.[0-9]+)"$/\1/p' "${VERSION_FILE}")"

if [[ -z "${version}" ]]; then
  echo "::error::Version must be declared as: const Version = \"X.Y.Z\""
  exit 1
fi

if [[ ! "${version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "::error::Version ${version} must follow SemVer format X.Y.Z"
  exit 1
fi

if git rev-parse --verify --quiet "refs/tags/v${version}" >/dev/null; then
  echo "::error::Git tag v${version} already exists"
  exit 1
fi

if git show "${BASE_REF}:${VERSION_FILE}" >/dev/null 2>&1; then
  base_version="$(git show "${BASE_REF}:${VERSION_FILE}" | sed -nE 's/^const[[:space:]]+Version[[:space:]]*=[[:space:]]*"([0-9]+\.[0-9]+\.[0-9]+)"$/\1/p')"

  if [[ "${version}" == "${base_version}" ]]; then
    echo "::error::Version must be changed in ${VERSION_FILE}. Current version ${version} is the same as ${BASE_REF}."
    exit 1
  fi
fi

echo "Version ${version} is valid and unreleased."
