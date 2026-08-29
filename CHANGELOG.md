# Changelog

## 0.3.0 - 2026-08-29

Breaking deployment release:

- Adopt the centrally assigned host port `65000` in standalone and ZimaOS deployment metadata while preserving container port `8507`.
- Require the primary ZimaOS Compose to use a direct `ISHIKU_SETUP_SECRET` placeholder instead of an external setup-secret file mount.
- Pin all shipped deployment defaults to the scanned Dyniku `0.3.0` multi-architecture image digest.
- Remove interpolation from the primary ZimaOS Compose and add executable policy regression checks.

Migration:

1. Back up `/DATA/AppData/dyniku/data`.
2. Stop the existing stack.
3. Import the new `zimaos-compose.yaml` and replace its setup-secret placeholder with a unique value of at least 32 characters before first start.
4. Ensure host port `65000` is available; container port `8507` does not change.
5. Start the stack and verify `/readyz` through `http://localhost:65000/readyz`.

Initialized installations have already closed first-run registration, so changing the setup-secret source does not reset or recreate the administrator account.

Rollback:

1. Stop Dyniku and preserve `/DATA/AppData/dyniku/data`.
2. Restore the previous Compose definition.
3. Pin `ghcr.io/maroishiku/dyniku:0.2.6@sha256:3ed92940f1cf027e52ac9a0596c98f9afea756b7870e6b7c99740f4fcb8c164d`.
4. Start the stack and verify health before reopening access.
