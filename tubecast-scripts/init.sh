#!/usr/bin/env bash
#
# init.sh  —  One-time setup helper for TubeCast
#
# 1) docker compose pull         (fetch image only)
# 2) curl -LOs ia binary         (from Internet Archive)
# 3) chmod +x ia                 (make it runnable)
# 4) ./ia configure              (prompt for S3 keys)
#
# Usage:  ./init.sh
#
set -e  # abort on any error
set -u  # error on unset vars

echo "🔧 0. Setting +x on all .sh scripts…"
chmod +x ./*.sh

echo "🍇 1. Pulling Docker images (this may take a minute)…"
docker compose pull

echo "🎧 2. Downloading Internet-Archive CLI (ia)…"
curl -LOs https://archive.org/download/ia-pex/ia

echo "🔑 3. Making 'ia' executable…"
chmod +x ia

echo "✅  Done!."