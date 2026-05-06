#!/bin/bash
set -euo pipefail
source .env
migrate -path migrations -database "$BPCL_DB_URL" "$@"
