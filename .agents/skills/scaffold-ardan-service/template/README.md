# tmplsvc

A Go web service generated from the Ardan Labs service architecture
(github.com/ardanlabs/service): a Domain Driven, Data Oriented layout with
in-process JWT/OPA auth, Postgres, OpenTelemetry, Kind and Helm.

## Requirements

- Go 1.27
- Docker (tests start a Postgres container)
- For Kubernetes: kind, kubectl, helm (`make dev-brew`)

## Quick start

```
make dev-docker
make test

# Kind with Postgres, Tempo and Grafana
make dev-run
make token
export TOKEN=<token from the previous call>
make users
make grafana          # http://localhost:3100

# Or Docker Compose (no tracing)
make compose-build-up
```

Seed users: `admin@example.com` and `user@example.com`, both with password
`gophers`.

## Layout

```
api/services/tmplsvc   service main.go, route set (build/), API tests
api/tooling/admin      migrate, seed, useradd, users, genkey, gentoken
api/tooling/logfmt     makes JSON logs readable: ... | go run api/tooling/logfmt/main.go
app/domain             authapp, checkapp, userapp, productapp
app/sdk                mid, errs, mux, auth (+rego), query, debug, metrics, apitest
business/domain        userbus, productbus (stores/, extensions/)
business/sdk           sqldb, migrate, delegate, page, order, dbtest, unittest
business/types         name, role, password, money, quantity
foundation             web, logger, otel, keystore, docker
zarf                   docker, compose, k8s/dev, helm/charts/tmplsvc, keys
```

`AGENTS.md` explains the architecture rules. The `add-domain` skill in
`.agents/skills` walks through adding a new domain.

## Endpoints

| Method | Path | Auth |
|---|---|---|
| GET | /v1/liveness, /v1/readiness | none |
| GET | /v1/auth/token/{kid} | Basic |
| GET, POST | /v1/users | admin |
| GET, PUT, DELETE | /v1/users/{user_id} | admin or the user |
| PUT | /v1/users/role/{user_id} | admin |
| GET | /v1/products | any role |
| POST | /v1/products | user |
| GET, PUT, DELETE | /v1/products/{product_id} | admin or the owner |

The debug server on :3010 serves pprof, expvar and statsviz.

## Configuration

Run `make run-help` for every setting. Environment variables use the
`TMPLSVC_` prefix, for example `TMPLSVC_DB_HOST`, `TMPLSVC_TEMPO_HOST` (empty
disables tracing) and `TMPLSVC_AUTH_KEYS_JSON`.

## Signing keys

`zarf/keys/<kid>.pem` was generated for this project and is for development
only. In Kubernetes the chart mounts keys from the `tmplsvc-keys` secret; in
production create that secret from your secret store, or set
`TMPLSVC_AUTH_KEYS_JSON`, and never deploy the dev key. Generate a new pair with
`go run api/tooling/admin/main.go genkey`.
