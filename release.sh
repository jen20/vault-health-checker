#!/usr/bin/env bash

set -o errexit
set -o pipefail

ROOT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

goreleaser --rm-dist \
	--config "${ROOT_DIR}/goreleaser/vault-health-checker.yml"
