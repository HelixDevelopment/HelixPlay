#!/usr/bin/env bash
# vasic-digital/Containers/scripts/bootstrap.sh
set -e
echo "Bootstrapping HelixPlay containers..."
docker-compose up -d
echo "Containers started. Verifying..."
docker-compose ps
