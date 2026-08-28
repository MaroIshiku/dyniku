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

```bash
mkdir -p /DATA/AppData/dyniku/data /DATA/AppData/dyniku/secrets
```

Lege anschliessend ein langes zufaelliges Setup-Secret an:

```bash
openssl rand -base64 48 > /DATA/AppData/dyniku/secrets/setup_secret.txt
chmod 600 /DATA/AppData/dyniku/secrets/setup_secret.txt
```

Starte die App:

```bash
docker compose up -d
```

Dyniku ist danach standardmaessig unter `http://localhost:65000` erreichbar. Compose veroeffentlicht den zentral zugewiesenen Host-Port `65000`; im Container bleibt Dyniku kompatibel auf Port `8507`.

### Erstes Starten

Beim ersten Oeffnen zeigt Dyniku automatisch das Registrierungsfenster fuer den ersten Adminaccount an. Die normale App ist vorher nicht erreichbar.

Die Registrierung ist nur moeglich, wenn das Setup-Secret korrekt eingegeben wird. Bevorzugt liest Dyniku das Secret aus:

```txt
/run/secrets/ishiku_setup_secret
```

### Adminaccount erstellen

Im Registrierungsfenster werden benoetigt:

- Setup-Secret aus `/DATA/AppData/dyniku/secrets/setup_secret.txt`
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

Bevorzugt wird ein Docker/Compose Secret als Datei. In `docker-compose.example.yml` wird dieses Secret nach `/run/secrets/ishiku_setup_secret` gemountet.

```yaml
secrets:
  ishiku_setup_secret:
    file: /DATA/AppData/dyniku/secrets/setup_secret.txt
```

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
