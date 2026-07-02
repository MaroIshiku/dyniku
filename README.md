# Dyniku

DDNS updater web GUI

> Dyniku is a self-hosted web interface for managing, storing, and monitoring dynamic DNS updates.

[![Build status](https://github.com/MaroIshiku/dyniku/actions/workflows/build.yml/badge.svg)](https://github.com/MaroIshiku/dyniku/actions/workflows/build.yml)
[![MIT](https://img.shields.io/github/license/MaroIshiku/dyniku)](LICENSE)
![Go version](https://img.shields.io/github/go-mod/go-version/MaroIshiku/dyniku)

<img height="160" alt="Dyniku logo" src="internal/server/ui/static/dyniku-logo.png">

## Summary

Dyniku is a self-hosted web app from the ishiku family. It combines the proven DDNS updater core with a Pixel Soft Utility web GUI for status, configuration, and IP history.

The app is designed for private or small personal deployments. The first start is protected by a setup secret and creates exactly one initial admin account.

## Part of the ishiku Family

Dyniku uses the shared ishiku interface:

- calm, rounded Pixel Soft Utility components
- six shared themes: Lavender, Mint, Sky, Amber, Rose, and Graphite
- Light, Dark, and System modes
- consistent AppHeader, profile, settings, About, and Admin areas
- consistent first-run setup for the first admin account

The app is intentionally meant to feel like part of a shared suite, not like a separate brand with its own design language.

## Features

- Web GUI for DDNS status, provider configuration, and public IP history
- First-run setup with setup secret and admin account
- Login with HttpOnly session cookie
- Password hashing with bcrypt, no plaintext passwords
- Support for many DNS providers, including Cloudflare, DuckDNS, Dynu, Hetzner, IONOS, Netcup, OVH, Porkbun, Route53, Strato, and more
- Manual "Update now" and periodic updates
- JSON configuration in a persistent data folder
- Docker image for `amd64`, `arm64`, and other platforms
- Health check for container operation

## Tech Stack

- Frontend: vanilla HTML, CSS, and JavaScript with the Pixel Soft Utility design system
- Backend: Go
- Storage: JSON files in the persistent data folder
- Deployment: Docker / Docker Compose

## Installation

### Docker Compose

Create the persistent host folders on ZimaOS or your Docker host:

```bash
mkdir -p /DATA/AppData/dyniku/data /DATA/AppData/dyniku/secrets
```

Create a long random setup secret:

```bash
openssl rand -base64 48 > /DATA/AppData/dyniku/secrets/setup_secret.txt
chmod 600 /DATA/AppData/dyniku/secrets/setup_secret.txt
```

Start the app:

```bash
docker compose up -d
```

Dyniku is then available by default at `http://localhost:8507`.

### First Start

On first open, Dyniku automatically shows the registration window for the first admin account. The normal app is not available before that account has been created.

Registration is only possible when the setup secret is entered correctly. Dyniku prefers to read the secret from:

```text
/run/secrets/ishiku_setup_secret
```

### Create the Admin Account

The registration window requires:

- setup secret from `/DATA/AppData/dyniku/secrets/setup_secret.txt`
- admin username
- display name
- optional email
- admin password
- password confirmation

The admin password must be at least 12 characters long, must not match the setup secret, and must not be a placeholder such as `admin`, `password`, `passwort`, `changeme`, `123456`, or `ishiku`.

After the first admin account is created successfully, public registration is closed. Further app functions require login.

## Configuration

### Environment Variables

| Variable | Description | Default |
| --- | --- | --- |
| `TZ` | Time zone for logs and display | empty |
| `ISHIKU_BASE_PATH` | Base path behind a reverse proxy | `/` |
| `ISHIKU_DATA_DIR` | Persistent data path in the container | `/data` |
| `ISHIKU_CONFIG_FILE` | DDNS JSON configuration path | `/data/config.json` |
| `ISHIKU_LOG_LEVEL` | Log level | `info` |
| `ISHIKU_SETUP_SECRET_FILE` | Path to the Docker secret | `/run/secrets/ishiku_setup_secret` |
| `ISHIKU_SETUP_SECRET` | Fallback secret as an environment variable, used only when no secret file is configured | empty |
| `LISTENING_ADDRESS` | Internal HTTP address | `:8507` |
| `PERIOD` | Update interval | `5m` |
| `UPDATE_COOLDOWN_PERIOD` | Cooldown between updates | `5m` |
| `HTTP_TIMEOUT` | HTTP timeout for provider and public IP requests | `10s` |
| `BACKUP_PERIOD` | Backup interval, `0` disables backups | `0` |
| `ISHIKU_BACKUP_DIRECTORY` | Backup target folder | `/data` |
| `SHOUTRRR_ADDRESSES` | Optional [Shoutrrr](https://containrrr.dev/shoutrrr/v0.8/services/overview/) notification URLs | empty |

Legacy variables such as `DATADIR`, `CONFIG_FILEPATH`, `ROOT_URL`, `LOG_LEVEL`, and `BACKUP_DIRECTORY` are still accepted. The `ISHIKU_*` names are preferred for new deployments.

### Docker Secrets

A Docker/Compose secret file is preferred. In `docker-compose.example.yml`, this secret is mounted to `/run/secrets/ishiku_setup_secret`.

```yaml
secrets:
  ishiku_setup_secret:
    file: /DATA/AppData/dyniku/secrets/setup_secret.txt
```

### Persistent Data

Persistent data is stored by default in:

```text
/DATA/AppData/dyniku/data
```

This folder contains, among other files:

- `config.json` for DDNS providers
- `auth.json` for the hashed admin account
- update and history data

Back up this folder regularly when Dyniku is used in production.

## Security

- The setup secret is only used for the first admin registration.
- Setup attempts with a wrong secret are rate-limited.
- The admin password must not match the setup secret.
- Passwords are stored with bcrypt hashes.
- Sessions are handled through HttpOnly cookies with SameSite=Lax.
- Public registration is closed after the first admin account.
- `/healthz` and `/readyz` are public; app and config APIs require login.
- Secrets, `.env`, databases, and logs do not belong in the repository.

## Updates and Backup

```bash
docker compose pull
docker compose up -d
```

Back up the persistent data folder before updates:

```bash
tar -czf backup-dyniku-$(date +%Y%m%d).tar.gz /DATA/AppData/dyniku
```

## Development

```bash
go test ./...
go run ./cmd/dyniku
```

Important local URLs:

- Web GUI: `http://localhost:8507`
- Health: `http://localhost:8507/healthz`
- Ready: `http://localhost:8507/readyz`

When making changes, keep the shared Pixel Soft Utility design system intact and avoid app-specific UI deviations.

## Created with ChatGPT Codex

This project was created and revised with support from ChatGPT Codex. Codex was used to generate and refine code, structure, UI components, and documentation according to the ishiku / Pixel Soft Utility standards.

Responsibility for operation, review, security, and publication remains with the repository owner.

## Status and License

Status: actively maintained self-hosted utility.

License: [MIT](LICENSE)
