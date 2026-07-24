#!/bin/bash
# ==============================================================================
# Hasir VPS Setup Script
# Configures swap, Docker, firewall, fail2ban, directories, and bootstraps SSL.
# ==============================================================================

set -euo pipefail

# Ensure the script is run as root
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root (sudo)."
  exit 1
fi

echo "=== 1. Creating Swap Space (2GB) ==="
if [ ! -f /swapfile ]; then
  fallocate -l 2G /swapfile
  chmod 600 /swapfile
  mkswap /swapfile
  swapon /swapfile
  echo '/swapfile none swap sw 0 0' >> /etc/fstab
  echo "Swap file created and enabled."
else
  echo "Swap file already exists."
fi

echo "=== 2. Installing Prerequisites & Docker ==="
apt-get update
apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release openssl ufw fail2ban

if ! command -v docker &> /dev/null; then
  curl -fsSL https://get.docker.com -o get-docker.sh
  sh get-docker.sh
  rm get-docker.sh
  echo "Docker installed."
else
  echo "Docker is already installed."
fi

echo "=== 3. Configuring Firewall (UFW) ==="
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp comment 'Default SSH'
ufw allow 2222/tcp comment 'Hasir Git SSH'
ufw allow 80/tcp comment 'HTTP'
ufw allow 443/tcp comment 'HTTPS'
ufw --force enable
echo "Firewall configured and enabled."

echo "=== 4. Configuring Fail2Ban ==="
systemctl enable fail2ban
systemctl start fail2ban
echo "Fail2Ban service started."

echo "=== 5. Creating Local Directories & Initializing .env ==="
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

mkdir -p "$REPO_DIR/certbot/www"
mkdir -p "$REPO_DIR/certbot/conf"

if [ ! -f "$REPO_DIR/.env" ]; then
  cp "$REPO_DIR/.env.example" "$REPO_DIR/.env"
  # Generate secrets automatically
  JWT_SEC=$(openssl rand -hex 32)
  PG_PASS=$(openssl rand -hex 24)
  sed -i "s/secure_random_jwt_secret_here/$JWT_SEC/g" "$REPO_DIR/.env"
  sed -i "s/secure_random_postgres_password_here/$PG_PASS/g" "$REPO_DIR/.env"
  echo "Created .env from .env.example and generated random secrets."
else
  echo ".env file already exists."
fi

# Load variables from .env
# shellcheck disable=SC1090
source "$REPO_DIR/.env"

echo "=== 6. Bootstrapping Dummy SSL Certificate for Nginx ==="
CERT_PATH="$REPO_DIR/certbot/conf/live/$DOMAIN"
if [ ! -f "$CERT_PATH/fullchain.pem" ]; then
  echo "No existing SSL certificate found. Bootstrapping dummy self-signed certificate for Nginx start..."
  mkdir -p "$CERT_PATH"
  openssl req -x509 -nodes -newkey rsa:2048 -days 30 \
    -keyout "$CERT_PATH/privkey.pem" \
    -out "$CERT_PATH/fullchain.pem" \
    -subj "/CN=$DOMAIN"
  echo "Dummy SSL certificate bootstrapped."
else
  echo "SSL certificate already exists."
fi

echo "=== Setup Completed Successfully ==="
echo "Please review your .env file and then run: ./scripts/deploy.sh"
