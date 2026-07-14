#!/bin/sh
# renew.sh - Certbot post-renew hook to reload Nginx

echo "Certificate successfully renewed! Reloading Nginx container..."
if [ -S /var/run/docker.sock ]; then
  # Send HUP signal to reload nginx configuration.
  # Using container_name "nginx" as set in docker-compose.yml.
  curl --unix-socket /var/run/docker.sock -X POST http://localhost/containers/nginx/kill?signal=HUP
  echo "HUP signal sent to Nginx container."
else
  echo "Docker socket not found. Make sure Nginx is reloaded manually."
fi
