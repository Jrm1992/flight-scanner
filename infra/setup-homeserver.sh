#!/usr/bin/env bash
# Home Server Setup for Flight Scanner
#
# This script sets up:
#   1. Cloudflare Tunnel (expose to internet with HTTPS)
#   2. GitHub Actions self-hosted runner (CI/CD)
#
# Prerequisites:
#   - Docker Desktop installed and running
#   - Cloudflare account with a domain
#   - GitHub repo access (admin)
#
# Usage:
#   chmod +x infra/setup-homeserver.sh
#   ./infra/setup-homeserver.sh

set -euo pipefail

APP_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "=== Flight Scanner Home Server Setup ==="
echo ""

# --- 1. Environment ---
echo "--- Step 1: Environment file ---"
if [ ! -f "$APP_DIR/.env" ]; then
  cp "$APP_DIR/.env.example" "$APP_DIR/.env"
  echo "  Created .env from .env.example"
  echo "  EDIT .env now with your values, then re-run this script."
  echo "  Required: POSTGRES_PASSWORD, JWT_SECRET"
  echo ""
  echo "  Generate JWT_SECRET with: openssl rand -base64 32"
  exit 0
else
  echo "  .env already exists"
fi

# --- 2. Start the app ---
echo ""
echo "--- Step 2: Starting app ---"
cd "$APP_DIR"
docker compose up --build -d
echo "  App is running on http://localhost"

# --- 3. Cloudflare Tunnel ---
echo ""
echo "--- Step 3: Cloudflare Tunnel ---"
echo ""
echo "  To expose your server to the internet:"
echo ""
echo "  1. Install cloudflared:"
echo "     brew install cloudflared"
echo ""
echo "  2. Login to Cloudflare:"
echo "     cloudflared tunnel login"
echo ""
echo "  3. Create a tunnel:"
echo "     cloudflared tunnel create flight-scanner"
echo ""
echo "  4. Route your domain to the tunnel:"
echo "     cloudflared tunnel route dns flight-scanner your-domain.com"
echo ""
echo "  5. Run the tunnel (points to nginx on port 80):"
echo "     cloudflared tunnel --url http://localhost:80 run flight-scanner"
echo ""
echo "  Or use the docker container (add to docker-compose.yml):"
echo "     cloudflared tunnel --no-autoupdate run --token <YOUR_TUNNEL_TOKEN>"

# --- 4. GitHub Actions Self-Hosted Runner ---
echo ""
echo "--- Step 4: GitHub Actions Self-Hosted Runner ---"
echo ""
echo "  1. Go to: https://github.com/Jrm1992/flight-scanner/settings/actions/runners/new"
echo "  2. Select macOS"
echo "  3. Follow the instructions to download and configure the runner"
echo "  4. Start the runner as a service:"
echo "     ./svc.sh install"
echo "     ./svc.sh start"
echo ""
echo "  The runner will automatically pick up deploy jobs from GitHub Actions."

echo ""
echo "=== Setup complete ==="
echo ""
echo "Summary:"
echo "  - App running at http://localhost"
echo "  - Configure Cloudflare Tunnel for HTTPS + public access"
echo "  - Install GitHub Actions runner for auto-deploy"
