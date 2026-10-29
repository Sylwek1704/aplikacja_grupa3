#!/bin/bash
set -e

cd "$(dirname "$0")"

bash build.sh
docker compose up -d --build

echo "czat:  http://localhost:8081"
echo "maile: http://localhost:1081"
