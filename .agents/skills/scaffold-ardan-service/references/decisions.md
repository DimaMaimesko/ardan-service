# Decisions and Differences from ardanlabs/service

The template follows this repository's code. These are the places where it
deliberately differs, and why. Agreed with the project owner on 2026-10-02.

## Scope

| Kept | Left out |
|---|---|
| user and product domains | home, audit, vproduct (view) domains |
| OTel extensions | audit extension, cache store |
| delegate (user deleted → product) | build-tag route sets (only `build/all.go`) |
| transaction middleware and `NewWithTx` (no example route) | SQLite dialect, `commondb` split |
| Compose, Kind, Helm, Tempo, Grafana | separate auth service, gRPC, OAuth |
| expvar metrics on the debug port | metrics sidecar, Prometheus, Loki, Alloy, Tilt, Vue frontend |

## Architecture choices

- **Auth runs in the service.** `mid.Bearer` verifies the JWT with
  `auth.Authenticate` (keystore + OPA) and `mid.Authorize*` evaluates rules with
  `auth.Authorize`. There is no auth client. `authapp` only issues tokens.
  Authorization failures return the same client message the separate auth
  service returned, so OPA details stay out of responses.
- **Business domain imports go one way.** `productbus` imports `userbus`, as in
  the code. The repository's `layered-architecture-types` skill forbids it; the
  generated project's copy of that skill states the one-way rule instead.
- **Converter names are `toApp<Type>`**, matching the code, not
  `fromBus<Type>Response` from this repository's AGENTS.md.

## Fixes over upstream

- `admin gentoken` passed the literal string `"SALAES_PEM"` to
  `keystore.LoadByJSON`, which always failed. It now loads keys from the folder
  only, and takes the issuer from config so tokens match the service.
- The sales Helm chart set `SALES_DB_HOST_PORT`, which no config field reads.
  The template chart sets `<PREFIX>_DB_HOST`.
- The sales Dockerfile stamped `main.tag` into the admin binary, whose variable
  is `main.build`. The template uses `main.build`.
- `otel.InitTracing` created the OTLP exporter before checking for an empty
  host. It now builds the exporter only when a host is set, so an empty
  `<PREFIX>_TEMPO_HOST` cleanly disables tracing.
- `foundation/web/web.go` was not gofmt clean.

## Operational choices

- **Signing keys are never baked into the image.** Compose bind-mounts
  `zarf/keys`. The Helm chart mounts the `<name>-keys` secret; in dev,
  `make dev-apply` passes the key with `--set-file` and the chart creates the
  secret. Production creates the secret or sets `<PREFIX>_AUTH_KEYS_JSON`.
- **Each project gets its own dev key** (random kid, RSA 2048, PKCS8). The file
  is mode 0644 so the container user can read it through the bind mount.
- **Production init container only migrates.** `values.yaml` runs
  `./admin migrate`; `values-dev.yaml` runs `./admin migrate-seed`.
- **Compose runs without tracing** (`<PREFIX>_TEMPO_HOST=` empty). Kind is the
  full environment with Tempo and Grafana.
- **Port names are generic** (`http`, `debug`) because Kubernetes limits port
  names to 15 characters and service names can be longer.
- **Test container is per project** (`<name>test`) instead of `servicetest`.
- **The generated AGENTS.md has no persona and no git rule.** Those are this
  repository's preferences; add them to the generated file if wanted.

## Known upstream quirks kept as is

- `name.Null.String()` returns `"NULL"` for an empty value, so a user without a
  department is returned with `"department":"NULL"`.
- Authorization failures use `errs.Unauthenticated` (401), not
  `errs.PermissionDenied` (403). The API tests depend on 401.
