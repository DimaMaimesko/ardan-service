---
name: scaffold-ardan-service
description: Create a new Go web service project in the Ardan Labs service architecture (Bill Kennedy's Domain Driven, Data Oriented layout) from a tested template. The project has api/app/business/foundation layers, routing, env-var config, CORS, JSON encoding, Postgres with darwin migrations, in-process JWT + OPA auth, user and product example domains, OpenTelemetry, Docker Compose, Kind and Helm. Use when the user asks to scaffold, bootstrap, generate or start a new service, project or template based on this repository or its architecture.
---

# Scaffold an Ardan Labs Style Service

This skill creates a new project by copying `template/`, a trimmed copy of this
repository that builds and passes its tests, and renaming its placeholders. Do
not write the project by hand: run the scaffold program, then verify.

## What the user gets

- `foundation/`: web framework, logger, otel, keystore, docker test helper.
- `business/`: `userbus` and `productbus` (Postgres store + OTel extension each,
  delegate wired for "user deleted"), sqldb, darwin migrations and seed, page,
  order, strong types (name, role, password, money, quantity), dbtest, unittest.
- `app/`: authapp (token by Basic auth), checkapp, userapp, productapp;
  middleware, errs, mux, query, debug, metrics, auth (JWT RS256 + embedded OPA
  rego), apitest.
- `api/`: the service `main.go` (config, CORS, graceful shutdown, debug port),
  admin tool (migrate, seed, useradd, users, genkey, gentoken), logfmt, API tests.
- `zarf/`: Dockerfile, Compose, Kind config, dev manifests for Postgres, Tempo and
  Grafana, a Helm chart, and a fresh dev signing key.
- `AGENTS.md`, `README.md`, makefile, and project skills in `.agents/skills`
  (`add-domain`, `layered-architecture-types`, `business-layer-extensions`,
  `branching-logic-flow`, `use-modern-go`).

Read `references/decisions.md` before answering questions about why the project
differs from this repository, and `references/architecture.md` for the rules.

## Workflow

### 1. Collect inputs

Ask for anything the user has not given. Do not guess.

| Input | Example | Rule |
|---|---|---|
| Module path | `github.com/acme/orders` | valid Go module path |
| Service name | `orders` | 2-40 chars, lowercase letters, digits, hyphens; starts with a letter |
| Env prefix | `ORDERS` | optional; defaults to the name upper-cased with `-` → `_` |
| Output directory | `~/code/orders` | must not exist or must be empty; keep it outside this repository |

None of them may contain the placeholders `tmplsvc` or `TMPLKID`. Confirm the
values with the user before running.

### 2. Preflight

- `go version` must be 1.27 or newer (the template's `go.mod` says `go 1.27.0`).
- Docker is needed only for the tests and for running the system.

### 3. Generate

Run from any directory, using the absolute path to this skill:

```
go run <skill-dir>/scripts/scaffold/main.go \
	-module <module> \
	-name <name> \
	-out <output-dir>
```

Add `-prefix <PREFIX>` to override the env prefix. The program refuses a
non-empty output directory, replaces placeholders in paths and file contents,
strips `.tmpl` suffixes, runs `go/format` on Go files, and writes
`zarf/keys/<kid>.pem`. It prints the kid.

### 4. Resolve dependencies and compile

In the output directory:

```
go mod tidy
go mod vendor
go build ./...
go vet ./...
```

Stop and report the error if any step fails. Do not patch generated code to
silence it; a failure means the template or the inputs are wrong.

### 5. Test

Ask the user before running the tests, because they start a Postgres container:

```
make test-only
```

If `helm` is installed, also run `make chart-lint`. If Docker or Helm is not
available, say which checks were skipped.

### 6. Report

Tell the user:

- where the project is, the module path, the service name, the env prefix and
  the kid;
- the seed logins `admin@example.com` and `user@example.com`, password `gophers`;
- how to run it: `make dev-run` (Kind with Tempo and Grafana) or
  `make compose-build-up`, then `make token` and `make users`;
- that `zarf/keys/<kid>.pem` is a development key and production keys must come
  from a secret (`<PREFIX>_AUTH_KEYS_JSON` or the mounted `<name>-keys` secret);
- which checks ran and which were skipped.

Do not run git commands in the new project. Give the user the commands instead:

```
cd <output-dir>
git init
git add .
git commit -m "Initial project from scaffold-ardan-service"
```

## After generation

- New domain: the generated project's `add-domain` skill.
- Remove the product example: delete `app/domain/productapp`,
  `business/domain/productbus`, `api/services/<name>/tests/productapi`, the
  product parts of `mux.BusConfig`, `main.go`, `dbtest/business.go`,
  `apitest/start.go`, `build/all.go`, `mid.AuthorizeProduct` and the product
  context helpers in `mid/mid.go`, `unittest`/`apitest` `Products` fields, the
  products table in `migrate.sql` and its seed rows. Then `go build ./...`.
- Tracing off: set `<PREFIX>_TEMPO_HOST` to an empty string.

## Maintaining the template

`template/` is a real Go module (`github.com/tmplorg/tmplsvc`) that the parent
repository's `go` commands ignore because it sits under a dot directory.

- Keep the placeholders: module `github.com/tmplorg/tmplsvc`, name `tmplsvc`,
  prefix `TMPLSVC`, kid `TMPLKID`. Use them only inside strings, paths and
  config, never as part of a Go identifier.
- Files that tools would otherwise pick up from this repository keep a `.tmpl`
  suffix: `AGENTS.md.tmpl`, `.gitignore.tmpl`, `.dockerignore.tmpl` and every
  `SKILL.md.tmpl`. The scaffold strips the suffix.
- After changing the template, verify it in place (`go mod download`,
  `go build ./...`, `go vet ./...`, `go test ./...` with Docker), then generate
  a throwaway project and repeat steps 4 and 5 there.
