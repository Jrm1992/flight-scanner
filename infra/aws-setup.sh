#!/usr/bin/env bash
# AWS Infrastructure Setup for Flight Scanner
# Run this once to create all required AWS resources.
#
# Prerequisites:
#   - AWS CLI configured (`aws configure`)
#   - Permissions for EC2, RDS, ECR, VPC
#
# Usage:
#   chmod +x infra/aws-setup.sh
#   ./infra/aws-setup.sh

set -euo pipefail

REGION="${AWS_REGION:-us-east-1}"
APP_NAME="flight-scanner"
DB_USER="flightscanner"
DB_PASS="${DB_PASSWORD:?Set DB_PASSWORD env var before running}"
DB_NAME="flightscannerdb"
KEY_NAME="${APP_NAME}-key"

echo "=== Flight Scanner AWS Setup (region: $REGION) ==="

# --- 1. ECR Repositories ---
echo ""
echo "--- Creating ECR repositories ---"
for repo in "${APP_NAME}-api" "${APP_NAME}-web"; do
  if aws ecr describe-repositories --repository-names "$repo" --region "$REGION" &>/dev/null; then
    echo "  ECR repo '$repo' already exists"
  else
    aws ecr create-repository \
      --repository-name "$repo" \
      --region "$REGION" \
      --image-scanning-configuration scanOnPush=true \
      --query 'repository.repositoryUri' --output text
    echo "  Created ECR repo: $repo"
  fi
done

# --- 2. Security Group ---
echo ""
echo "--- Creating Security Group ---"
VPC_ID=$(aws ec2 describe-vpcs --filters "Name=isDefault,Values=true" \
  --query "Vpcs[0].VpcId" --output text --region "$REGION")

SG_ID=$(aws ec2 describe-security-groups \
  --filters "Name=group-name,Values=${APP_NAME}-sg" \
  --query "SecurityGroups[0].GroupId" --output text --region "$REGION" 2>/dev/null || true)

if [ "$SG_ID" = "None" ] || [ -z "$SG_ID" ]; then
  SG_ID=$(aws ec2 create-security-group \
    --group-name "${APP_NAME}-sg" \
    --description "Flight Scanner - HTTP, HTTPS, SSH" \
    --vpc-id "$VPC_ID" \
    --query "GroupId" --output text --region "$REGION")

  aws ec2 authorize-security-group-ingress --group-id "$SG_ID" --region "$REGION" \
    --ip-permissions \
    "IpProtocol=tcp,FromPort=22,ToPort=22,IpRanges=[{CidrIp=0.0.0.0/0,Description=SSH}]" \
    "IpProtocol=tcp,FromPort=80,ToPort=80,IpRanges=[{CidrIp=0.0.0.0/0,Description=HTTP}]" \
    "IpProtocol=tcp,FromPort=443,ToPort=443,IpRanges=[{CidrIp=0.0.0.0/0,Description=HTTPS}]"
  echo "  Created Security Group: $SG_ID"
else
  echo "  Security Group already exists: $SG_ID"
fi

# --- 3. Key Pair ---
echo ""
echo "--- Creating Key Pair ---"
if aws ec2 describe-key-pairs --key-names "$KEY_NAME" --region "$REGION" &>/dev/null; then
  echo "  Key pair '$KEY_NAME' already exists"
else
  aws ec2 create-key-pair \
    --key-name "$KEY_NAME" \
    --query "KeyMaterial" --output text --region "$REGION" \
    > "${KEY_NAME}.pem"
  chmod 400 "${KEY_NAME}.pem"
  echo "  Created key pair: ${KEY_NAME}.pem (SAVE THIS FILE!)"
fi

# --- 4. RDS PostgreSQL (Free Tier) ---
echo ""
echo "--- Creating RDS PostgreSQL instance ---"
if aws rds describe-db-instances --db-instance-identifier "$APP_NAME" --region "$REGION" &>/dev/null; then
  echo "  RDS instance '$APP_NAME' already exists"
else
  # Create a DB security group that allows EC2 SG to connect
  DB_SG_ID=$(aws ec2 create-security-group \
    --group-name "${APP_NAME}-db-sg" \
    --description "Flight Scanner DB - PostgreSQL from EC2" \
    --vpc-id "$VPC_ID" \
    --query "GroupId" --output text --region "$REGION" 2>/dev/null || \
    aws ec2 describe-security-groups \
      --filters "Name=group-name,Values=${APP_NAME}-db-sg" \
      --query "SecurityGroups[0].GroupId" --output text --region "$REGION")

  aws ec2 authorize-security-group-ingress --group-id "$DB_SG_ID" --region "$REGION" \
    --protocol tcp --port 5432 --source-group "$SG_ID" 2>/dev/null || true

  aws rds create-db-instance \
    --db-instance-identifier "$APP_NAME" \
    --db-instance-class db.t3.micro \
    --engine postgres \
    --engine-version "16" \
    --master-username "$DB_USER" \
    --master-user-password "$DB_PASS" \
    --allocated-storage 20 \
    --db-name "$DB_NAME" \
    --vpc-security-group-ids "$DB_SG_ID" \
    --no-multi-az \
    --backup-retention-period 1 \
    --no-auto-minor-version-upgrade \
    --publicly-accessible \
    --region "$REGION"
  echo "  RDS instance creating... (takes ~5 min)"
  echo "  Run: aws rds describe-db-instances --db-instance-identifier $APP_NAME --query 'DBInstances[0].Endpoint' --region $REGION"
fi

# --- 5. EC2 Instance ---
echo ""
echo "--- Creating EC2 instance ---"

# Get latest Amazon Linux 2023 AMI
AMI_ID=$(aws ec2 describe-images \
  --owners amazon \
  --filters "Name=name,Values=al2023-ami-2023*-x86_64" "Name=state,Values=available" \
  --query "sort_by(Images, &CreationDate)[-1].ImageId" --output text --region "$REGION")

INSTANCE_ID=$(aws ec2 describe-instances \
  --filters "Name=tag:Name,Values=${APP_NAME}" "Name=instance-state-name,Values=running,stopped" \
  --query "Reservations[0].Instances[0].InstanceId" --output text --region "$REGION" 2>/dev/null || true)

if [ "$INSTANCE_ID" != "None" ] && [ -n "$INSTANCE_ID" ]; then
  echo "  EC2 instance already exists: $INSTANCE_ID"
else
  # User data script to install Docker
  USER_DATA=$(cat <<'USERDATA'
#!/bin/bash
dnf update -y
dnf install -y docker
systemctl enable docker
systemctl start docker
usermod -aG docker ec2-user

# Install Docker Compose
DOCKER_COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep tag_name | cut -d '"' -f 4)
curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose

# Install AWS CLI (for ECR login)
# Already pre-installed on Amazon Linux 2023
USERDATA
)

  INSTANCE_ID=$(aws ec2 run-instances \
    --image-id "$AMI_ID" \
    --instance-type t2.micro \
    --key-name "$KEY_NAME" \
    --security-group-ids "$SG_ID" \
    --user-data "$USER_DATA" \
    --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=${APP_NAME}}]" \
    --query "Instances[0].InstanceId" --output text --region "$REGION")
  echo "  Created EC2 instance: $INSTANCE_ID"
fi

# --- Summary ---
echo ""
echo "=== Setup Complete ==="
echo ""
echo "Next steps:"
echo "  1. Wait for RDS to be available (~5 min):"
echo "     aws rds wait db-instance-available --db-instance-identifier $APP_NAME --region $REGION"
echo ""
echo "  2. Get RDS endpoint:"
echo "     aws rds describe-db-instances --db-instance-identifier $APP_NAME \\"
echo "       --query 'DBInstances[0].Endpoint.Address' --output text --region $REGION"
echo ""
echo "  3. Get EC2 public IP:"
echo "     aws ec2 describe-instances --instance-ids $INSTANCE_ID \\"
echo "       --query 'Reservations[0].Instances[0].PublicIpAddress' --output text --region $REGION"
echo ""
echo "  4. Add these GitHub Actions secrets:"
echo "     - AWS_ACCESS_KEY_ID"
echo "     - AWS_SECRET_ACCESS_KEY"
echo "     - EC2_HOST (public IP from step 3)"
echo "     - EC2_SSH_KEY (contents of ${KEY_NAME}.pem)"
echo "     - DATABASE_URL (postgres://${DB_USER}:${DB_PASS}@<rds-endpoint>:5432/${DB_NAME})"
echo "     - JWT_SECRET (openssl rand -base64 32)"
echo "     - SERPAPI_KEY"
