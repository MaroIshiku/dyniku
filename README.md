# Dyniku

DDNS-Updater Web GUI

> Dyniku ist eine self-hosted Weboberflaeche zum Verwalten, Speichern und Ueberwachen dynamischer DNS-Updates.

[![Build status](https://github.com/MaroIshiku/dyniku/actions/workflows/build.yml/badge.svg)](https://github.com/MaroIshiku/dyniku/actions/workflows/build.yml)
[![MIT](https://img.shields.io/github/license/MaroIshiku/dyniku)](LICENSE)
![Go version](https://img.shields.io/github/go-mod/go-version/MaroIshiku/dyniku)

<img height="160" alt="Dyniku logo" src="assets/icon-256.png">

## Kurzbeschreibung

Dyniku ist eine self-hosted Web-App aus der ishiku-Familie. Sie verbindet den bewaehrten DDNS-Updater-Kern mit einer Pixel Soft Utility Web GUI fuer Status, Konfiguration und IP-Verlauf.

Die App ist fuer private oder kleine eigene Deployments gedacht. Der erste Start ist durch ein Setup-Secret geschuetzt und legt genau einen initialen Adminaccount an.

## Teil der ishiku-Familie

Dyniku verwendet die gemeinsame ishiku Oberflaeche:

- ruhige, abgerundete Pixel-Soft-Utility-Komponenten
- sechs gemeinsame Themes: Lavender, Mint, Sky, Amber, Rose und Graphite
- Light, Dark und System Mode
- einheitlicher AppHeader, Profil-, Settings-, About- und Admin-Bereiche
- einheitliches First-Run-Setup fuer den ersten Adminaccount

Die App soll sich bewusst wie Teil einer gemeinsamen Suite anfuehlen, nicht wie eine separate Marke mit eigener Designsprache.

## Funktionen

- Web GUI fuer DDNS-Status, Provider-Konfiguration und Public-IP-Verlauf
- First-Run-Setup mit Setup-Secret und Adminaccount
- Login mit HttpOnly Session-Cookie
- Passwort-Hashing mit bcrypt, keine Klartext-Passwoerter
- Unterstuetzung vieler DNS-Provider, unter anderem Cloudflare, DuckDNS, Dynu, Hetzner, IONOS, Netcup, OVH, Porkbun, Route53, Strato und weitere
- Manuelles "Update now" und periodische Updates
- JSON-Konfiguration in einem persistenten Datenordner
- Docker-Image fuer `amd64`, `arm64` und weitere Plattformen
- Healthcheck fuer Containerbetrieb

## Tech Stack

- Frontend: Vanilla HTML, CSS und JavaScript mit Pixel Soft Utility Designsystem
- Backend: Go
- Datenhaltung: JSON-Dateien im persistenten Datenordner
- Deployment: Docker / Docker Compose

## Installation

### Docker Compose

All shipped Compose profiles contain a clearly synthetic setup-secret replacement value. Replace it with a unique value of at least 32 characters before the first start:

```yaml
ISHIKU_SETUP_SECRET: "REPLACE-WITH-A-UNIQUE-SECRET-OF-AT-LEAST-32-CHARACTERS"
```

Create the persistent data directory:

```bash
mkdir -p /DATA/AppData/dyniku/data
```

Starte die App:

```bash
docker compose up -d
```

Dyniku ist danach standardmaessig unter `http://localhost:65000` erreichbar. Compose veroeffentlicht den zentral zugewiesenen Host-Port `65000`; im Container bleibt Dyniku kompatibel auf Port `8507`.

### ZimaOS (breaking since 0.3.0)

`zimaos-compose.yaml` is the primary ZimaOS import and intentionally uses direct scalar values. Before importing it, replace:

```yaml
ISHIKU_SETUP_SECRET: "REPLACE-WITH-A-UNIQUE-SECRET-OF-AT-LEAST-32-CHARACTERS"
```

with a unique value of at least 32 characters. Do not commit or reuse that value. ZimaOS publishes host port `65000`; Dyniku continues to listen on container port `8507`.

Upgrading an existing ZimaOS installation to `0.3.0` is a breaking deployment change. Back up `/DATA/AppData/dyniku/data`, stop the existing stack, import the new Compose file, verify that port `65000` is free, and then start Dyniku. An installation that already has an administrator remains initialized; the setup-secret source change does not recreate the account.

### Erstes Starten

Beim ersten Oeffnen zeigt Dyniku automatisch das Registrierungsfenster fuer den ersten Adminaccount an. Die normale App ist vorher nicht erreichbar.

Die Registrierung ist nur moeglich, wenn das Setup-Secret korrekt eingegeben wird. Bevorzugt liest Dyniku das Secret aus:

```txt
/run/secrets/ishiku_setup_secret
```

### Adminaccount erstellen

Im Registrierungsfenster werden benoetigt:

- the setup-secret value configured in the Compose file
- Admin-Benutzername
- Anzeigename
- optional E-Mail
- Admin-Passwort
- Passwort-Wiederholung

Das Admin-Passwort muss mindestens 12 Zeichen lang sein, darf nicht mit dem Setup-Secret uebereinstimmen und darf kein Platzhalter wie `admin`, `password`, `passwort`, `changeme`, `123456` oder `ishiku` sein.

Nach erfolgreicher Erstellung des ersten Adminaccounts wird die oeffentliche Registrierung geschlossen. Weitere App-Funktionen sind danach nur nach Login erreichbar.

## Konfiguration

### Umgebungsvariablen

| Variable | Beschreibung | Standard |
| --- | --- | --- |
| `TZ` | Zeitzone fuer Logs und Anzeige | leer |
| `ISHIKU_BASE_PATH` | Basis-Pfad hinter Reverse Proxy | `/` |
| `ISHIKU_DATA_DIR` | Persistenter Datenpfad im Container | `/data` |
| `ISHIKU_CONFIG_FILE` | Pfad zur DDNS JSON-Konfiguration | `/data/config.json` |
| `ISHIKU_LOG_LEVEL` | Log-Level | `info` |
| `ISHIKU_SETUP_SECRET_FILE` | Pfad zum Docker Secret | `/run/secrets/ishiku_setup_secret` |
| `ISHIKU_SETUP_SECRET` | Fallback-Secret als ENV, nur wenn kein Secret-File genutzt wird | leer |
| `LISTENING_ADDRESS` | Interne HTTP-Adresse | `:8507` |
| `PERIOD` | Update-Intervall | `5m` |
| `UPDATE_COOLDOWN_PERIOD` | Cooldown zwischen Updates | `5m` |
| `HTTP_TIMEOUT` | HTTP-Timeout fuer Provider und Public-IP-Abfragen | `10s` |
| `BACKUP_PERIOD` | Backup-Intervall, `0` deaktiviert Backups | `0` |
| `ISHIKU_BACKUP_DIRECTORY` | Backup-Zielordner | `/data` |
| `SHOUTRRR_ADDRESSES` | Optionale [Shoutrrr](https://containrrr.dev/shoutrrr/v0.8/services/overview/) Notification-URLs | leer |

Legacy-Variablen wie `DATADIR`, `CONFIG_FILEPATH`, `ROOT_URL`, `LOG_LEVEL` und `BACKUP_DIRECTORY` werden weiterhin akzeptiert. Die `ISHIKU_*` Namen sind fuer neue Deployments bevorzugt.

### Docker Secrets

The shipped Compose files use a direct synthetic `ISHIKU_SETUP_SECRET` replacement value for consistent ZimaOS, Docker Compose, and Portainer imports. Replace it locally before deployment and never commit the real value.

Dyniku still supports `ISHIKU_SETUP_SECRET_FILE` as an optional compatibility input for custom deployments, but it is not the default in any shipped Compose file.

### Persistente Daten

Persistente Daten liegen standardmaessig in:

```txt
/DATA/AppData/dyniku/data
```

In diesem Ordner liegen unter anderem:

- `config.json` fuer DDNS-Provider
- `auth.json` fuer den gehashten Adminaccount
- Update- und Verlaufdaten

Sichere diesen Ordner regelmaessig, wenn Dyniku produktiv genutzt wird.

## Sicherheit

- Das Setup-Secret dient nur zur ersten Admin-Registrierung.
- Setup-Versuche mit falschem Secret werden rate-limited.
- Das Admin-Passwort darf nicht dem Setup-Secret entsprechen.
- Passwoerter werden mit bcrypt gehasht gespeichert.
- Sessions werden ueber HttpOnly Cookies mit SameSite=Lax verwaltet.
- Die oeffentliche Registrierung wird nach dem ersten Adminaccount geschlossen.
- `/healthz` und `/readyz` sind oeffentlich, App- und Config-APIs brauchen Login.
- Secrets, `.env`, Datenbanken und Logs gehoeren nicht ins Repository.

## Updates und Backup

```bash
docker compose pull
docker compose up -d
```

Vor Updates sollte der persistente Datenordner gesichert werden:

```bash
tar -czf backup-dyniku-$(date +%Y%m%d).tar.gz data
```

Release `0.3.0` is pinned in the shipped Compose files by both semantic version and image digest. Detailed migration and rollback instructions are recorded in [CHANGELOG.md](CHANGELOG.md).

## Entwicklung

```bash
go test ./...
go run ./cmd/dyniku
```

Wichtige lokale URLs:

- Web GUI: `http://localhost:8507`
- Health: `http://localhost:8507/healthz`
- Ready: `http://localhost:8507/readyz`

Codex soll bei Aenderungen das gemeinsame Pixel Soft Utility Designsystem beibehalten und keine app-spezifischen UI-Abweichungen einfuehren.

## Erstellt mit ChatGPT Codex

Dieses Projekt wurde mit Unterstuetzung von ChatGPT Codex erstellt bzw. ueberarbeitet. Codex wurde verwendet, um Code, Struktur, UI-Komponenten und Dokumentation nach den Vorgaben der ishiku / Pixel Soft Utility Standards zu generieren.

Die Verantwortung fuer Betrieb, Pruefung, Sicherheit und Veroeffentlichung liegt beim Repository-Betreiber.

## Status und Lizenz

Status: aktiv gepflegtes Self-hosted Utility.

Lizenz: [MIT](LICENSE)
