#!/usr/bin/env bash
# Manual deploy script for EC2
# Run this on the EC2 instance or via SSH.
#
# First-time setup on EC2:
#   mkdir -p /home/ec2-user/flight-scanner
#   # Copy docker-compose.yml, infra/nginx.conf, and .env to the server
#
# Usage:
#   chmod +x deploy.sh
#   ./deploy.sh

set -euo pipefail

REGION="${AWS_REGION:-us-east-1}"
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
ECR_REGISTRY="${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com"
APP_DIR="/home/ec2-user/flight-scanner"

echo "=== Flight Scanner Deploy ==="

# Login to ECR
echo "Logging in to ECR..."
aws ecr get-login-password --region "$REGION" | docker login --username AWS --password-stdin "$ECR_REGISTRY"

# Pull latest images
echo "Pulling latest images..."
docker pull "${ECR_REGISTRY}/flight-scanner-api:latest"
docker pull "${ECR_REGISTRY}/flight-scanner-web:latest"

# Deploy
cd "$APP_DIR"
echo "Restarting containers..."
export ECR_REGISTRY="$ECR_REGISTRY"
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml up -d

# Cleanup
docker image prune -f

# Health check
echo "Waiting for health check..."
sleep 5
if curl -sf http://localhost/health > /dev/null; then
  echo "=== Deploy successful! ==="
else
  echo "=== WARNING: Health check failed ==="
  docker compose logs --tail=20
  exit 1
fi
