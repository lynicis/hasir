# Hasir Docker Compose Stack

This repository provides a production-ready, highly secure Docker Compose stack designed to deploy the full Hasir platform on a single VPS or VM.

## Architecture Diagram

```
                 +--------------------------------------------------------+
                 |                      Client (Web / Git)                |
                 +--------------------+-------------------+---------------+
                                      |                   |
                                HTTP  |                   | Git (SSH)
                              Port 80 |                   | Port 2222
                                      v                   v
                 +--------------------+-------------------+---------------+
                 |                                                        |
                 |                      Docker Host                       |
                 |                                                        |
                 |   +------------------+          +------------------+   |
                 |   |                  |          |                  |   |
                 |   |  nginx (80/443)  |          |  certbot (ACME)  |   |
                 |   |  (TLS, Headers,  |          |  (SSL Renew)     |   |
                 |   |   Rate Limit)    |          |                  |   |
                 |   +--------+---------+          +--------+---------+   |
                 |            |                             |             |
                 |            | Proxy                       | Shared Vol  |
                 |            v                             v             |
                 |   +--------+---------+          +--------+---------+   |
                 |   |                  |          |                  |   |
                 |   | hasir-dashboard  |          | /etc/letsencrypt |   |
                 |   |    (Port 3000)   |          |                  |   |
                 |   +------------------+          +------------------+   |
                 |            |                             |             |
                 |            +-----------------------+     |             |
                 |                                    |     |             |
                 |                                    v     v             |
                 |                           +--------+---------+         |
                 |                           |                  |         |
                 |                           |    hasir-api     |         |
                 |                           |   (Port 8080/    |         |
                 |                           |    SSH 2222)     |         |
                 |                           +--------+---------+         |
                 |                                    |                   |
                 |                                    | DB Query          |
                 |                                    v                   |
                 |                           +--------+---------+         |
                 |                           |                  |         |
                 |                           |     postgres     |         |
                 |                           |   (Port 5432)    |         |
                 |                           +------------------+         |
                 |                                                        |
                 +--------------------------------------------------------+
```

## Features

- **Automated SSL:** Certbot generates and automatically renews Let's Encrypt certificates every 12 hours.
- **Nginx Security:** Robust configuration with TLS 1.3/1.2, HSTS, secure headers (CSP, Frame options), WebSocket/gRPC support, and rate-limiting zones.
- **PostgreSQL Database:** Handled inside the internal network with container-based health checks.
- **Persistent Storage:** Volumes for DB data (`pgdata`), SSH Keys (`ssh_keys`), repositories (`repos`), and SDK assets (`sdk`).
- **Idempotent Automation:** Complete VPS setup, deployment, and backup scripts.
- **Non-root & Low resource:** Optimized container logging and resource boundaries.

---

## Prerequisites

Before starting, ensure you have:
1. A Linux VPS (Ubuntu 20.04+ recommended) with a public IP.
2. A registered domain name pointing to your VPS IP address (DNS A/AAAA records).
3. Open ports: `80` (HTTP), `443` (HTTPS), and `2222` (Hasir Git SSH).

---

## Getting Started

### 1. Clone the repository
Clone this repository to your target VPS:
```bash
git clone git@github.com:protohasir/docker-compose.git
cd docker-compose
```

### 2. Run the Setup Script
The setup script configures 2GB of swap space, installs Docker and Docker Compose, configures the UFW firewall, enables Fail2Ban, initializes the `.env` file, and bootstraps a dummy self-signed SSL certificate so Nginx can start successfully:
```bash
sudo ./scripts/setup.sh
```

### 3. Edit the Environment variables
Open the newly created `.env` file and verify or update your settings:
```bash
nano .env
```
Ensure `DOMAIN` and `LETSENCRYPT_EMAIL` are correct. Database and JWT secrets are generated automatically during the setup script.

### 4. Deploy the Stack
Run the deploy script to pull the latest images, start services, and automatically request a real Let's Encrypt production certificate (overwriting the bootstrapped dummy certificate):
```bash
sudo ./scripts/deploy.sh
```

---

## Backup and Restore

### Creating a Backup
Backups are performed using the `scripts/backup.sh` script. It creates a gzipped PostgreSQL dump, applies a retention policy (keeping daily backups for 7 days, and weekly Sunday backups for 4 weeks), and optionally uploads to an AWS S3 bucket if configured.

Run backups manually or schedule them via cron:
```bash
# Run backup manually
sudo ./scripts/backup.sh
```

Add a daily cron job to automate backups:
```bash
# Edit crontab
sudo crontab -e

# Add the following line to run backups daily at 2:00 AM
0 2 * * * /path/to/docker-compose/scripts/backup.sh >> /var/log/hasir-backup.log 2>&1
```

### Restoring a Backup
To restore a database dump:
1. Locate the backup file in `backups/`.
2. Extract and restore it into the running database container:
```bash
# Gunzip and restore directly to the container
gunzip -c backups/hasir_db_YYYYMMDD_HHMMSS.sql.gz | docker compose exec -T postgres psql -U hasir -d hasir
```

---

## Troubleshooting FAQ

### Nginx fails to start due to missing certificates
If Nginx fails to start because Let's Encrypt certificates do not exist, run `./scripts/setup.sh` again to ensure the dummy self-signed certificates are bootstrapped. The deploy script handles requesting the actual certificate and reloading Nginx afterwards.

### Checking container logs
You can view the logs of any service by running:
```bash
docker compose logs -f <service-name>
# Examples:
docker compose logs -f nginx
docker compose logs -f hasir-api
```

### Checking health status of services
All critical services define health checks. You can check the current health status of services with:
```bash
docker compose ps
```

### Connecting to Hasir Git via SSH
Once deployed, SSH is exposed on port `2222`. Ensure your users add their SSH public keys via the dashboard and then connect via:
```bash
ssh -p 2222 git@your-domain.com
```
