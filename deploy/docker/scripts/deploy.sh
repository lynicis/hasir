#!/bin/bash
# ==============================================================================
# Hasir Deploy Script
# Pulls latest images, starts services, requests Let's Encrypt SSL if dummy cert
# is present, and cleans up old images.
# ==============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

cd "$REPO_DIR"

# Load environment variables
if [ -f .env ]; then
  # shellcheck disable=SC1091
  source .env
else
  echo "No .env file found. Run ./scripts/setup.sh first."
  exit 1
fi

echo "=== 1. Pulling Latest Docker Images ==="
docker compose pull

echo "=== 2. Starting Services ==="
docker compose up -d

# Check if the certificate is a bootstrapped dummy certificate
CERT_FILE="certbot/conf/live/$DOMAIN/fullchain.pem"
if [ -f "$CERT_FILE" ]; then
  if openssl x509 -in "$CERT_FILE" -text -noout | grep -qiE "localhost|$DOMAIN"; then
    # In some setups, check if issuer is Let's Encrypt or check if the cert subject matches the placeholder CN
    # To be safe, if the certificate's issuer details indicate a self-signed or placeholder, we replace it.
    # We can inspect the issuer's Common Name.
    ISSUER=$(openssl x509 -in "$CERT_FILE" -issuer -noout)
    if [[ "$ISSUER" == *"localhost"* || "$ISSUER" == *"$DOMAIN"* ]]; then
      echo "=== 3. Requesting Production Let's Encrypt SSL Certificate ==="
      
      # Stop Nginx first if it fails to serve the challenge because of self-signed configs,
      # but with our configuration, Nginx serves port 80 challenge fine regardless of port 443 SSL.
      # So we can run the certbot certonly command while Nginx is running!
      docker compose run --rm --entrypoint \
        "certbot certonly --webroot -w /var/www/certbot \
        --email $LETSENCRYPT_EMAIL --agree-tos --no-eff-email \
        -d $DOMAIN" certbot
      
      echo "=== 4. Reloading Nginx with Production Certificate ==="
      docker compose exec nginx nginx -s reload
    fi
  fi
fi

echo "=== 5. Pruning Old Docker Images ==="
docker image prune -f

echo "=== Deployment Completed Successfully ==="
