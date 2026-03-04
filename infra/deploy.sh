#!/usr/bin/env bash
# Manual deploy script for VPS
#
# First-time setup:
#   1. Install Docker and Docker Compose on the VPS
#   2. Clone the repo: git clone <repo-url> /app/flight-scanner
#   3. Create .env: cp .env.example .env && edit .env
#   4. Run: ./infra/deploy.sh
#
# Usage:
#   chmod +x infra/deploy.sh
#   ./infra/deploy.sh

set -euo pipefail

APP_DIR="/app/flight-scanner"
cd "$APP_DIR"

echo "=== Flight Scanner Deploy ==="

echo "Pulling latest code..."
git pull origin main

echo "Building and starting containers..."
docker compose up --build -d

echo "Cleaning up old images..."
docker image prune -f

echo "Waiting for health check..."
sleep 5
if curl -sf http://localhost/health > /dev/null; then
  echo "=== Deploy successful! ==="
else
  echo "=== WARNING: Health check failed ==="
  docker compose logs --tail=20
  exit 1
fi
