# Changelog

## 0.3.1 - 2026-08-30

Security and deployment hardening release:

Published image: `ghcr.io/maroishiku/dyniku:0.3.1@sha256:42ec75dcb5d218ae8cc7f6ddbe0d43b8b777b4066e695d90da36679dde5e851f` (`linux/amd64`, `linux/arm64`).

- Hash new administrator passwords with Argon2id using the ishiku baseline parameters.
- Preserve compatibility with existing bcrypt accounts and upgrade their password hash automatically after a successful login.
- Rate-limit failed logins by normalized account and network signal, enforce idle and absolute session expiry, reject cross-origin browser mutations, add restrictive browser security headers, and emit structured authentication audit events without secret values.
- Apply the non-root, read-only, capability-free, bounded-resource runtime contract to every shipped Compose profile.
- Harden the Kubernetes example with persistent storage, non-root security contexts, probes, resource bounds, and a synthetic setup-secret replacement.
- Update the vulnerable `golang.org/x/mod` dependency to its fixed release.
- Update pinned GHCR publishing actions to their current major releases.

Upgrade:

1. Back up `/DATA/AppData/dyniku/data` and stop Dyniku.
2. Preserve all existing files, including `auth.json`; no reinstallation or account recreation is required.
3. Run `sudo chown -R 10001:10001 /DATA/AppData/dyniku/data` so the hardened runtime can read and update existing data.
4. Replace the Compose definition, keep the configured direct setup-secret placeholder replacement, and confirm host port `65000` is free.
5. Start Dyniku and verify `http://localhost:65000/readyz`, then sign in with the existing administrator password. A successful legacy login upgrades its bcrypt hash in place.

Rollback preserves `/DATA/AppData/dyniku/data` and pins the previous `0.3.0` image digest `sha256:4ed4741a1294281d4ede33a0345bd3abf624d0b2f5e65142b6a6c5a9b9575214`.

## 0.3.0 - 2026-08-29

Breaking deployment release:

- Adopt the centrally assigned host port `65000` in standalone and ZimaOS deployment metadata while preserving container port `8507`.
- Require every shipped Compose file to use a direct `ISHIKU_SETUP_SECRET` placeholder instead of an external setup-secret file mount.
- Pin all shipped deployment defaults to the scanned Dyniku `0.3.0` multi-architecture image digest.
- Remove interpolation from the primary ZimaOS Compose and add executable policy regression checks.

Migration:

1. Back up `/DATA/AppData/dyniku/data`.
2. Stop the existing stack.
3. Preserve all files and run `sudo chown -R 10001:10001 /DATA/AppData/dyniku/data` for the fixed non-root runtime identity.
4. Import the new `zimaos-compose.yaml` and replace its setup-secret placeholder with a unique value of at least 32 characters before first start.
5. Ensure host port `65000` is available; container port `8507` does not change.
6. Start the stack and verify `/readyz` through `http://localhost:65000/readyz`.

Initialized installations have already closed first-run registration, so changing the setup-secret source does not reset or recreate the administrator account.

Rollback:

1. Stop Dyniku and preserve `/DATA/AppData/dyniku/data`.
2. Restore the previous Compose definition.
3. Pin `ghcr.io/maroishiku/dyniku:0.2.6@sha256:3ed92940f1cf027e52ac9a0596c98f9afea756b7870e6b7c99740f4fcb8c164d`.
4. Start the stack and verify health before reopening access.
