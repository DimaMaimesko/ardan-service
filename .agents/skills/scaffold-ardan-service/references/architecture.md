# Architecture Rules

A condensed version of the design guide for this repository. Use it to explain
the generated project or to customize it without breaking its shape.

## Philosophy

- Complexity is the cost being managed: one place for each kind of code, one
  direction for imports, one way to do each job.
- Start from a working system and add beside it (a new domain, store or
  extension) instead of generalizing what works.
- Data oriented: each layer owns its data model and converts at the boundary.
- Domain driven: inside each layer, code is grouped by business concept.
- Validate once at the edge; strong types carry the guarantee inward.
- Explicit over clever: no DI container, no `init()` wiring; `main.go` builds
  everything in order.
- Standard library first, small owned foundation packages, vendored modules.
- Deploy first: Kubernetes, tracing and logs exist from day one.

## Layers

```
api/         main packages, route sets (build/), API tests
app/         protocol edge: routes, decode, toBus validation, auth, error mapping
business/    domain rules, strong types, Storer interfaces and stores, migrations
foundation/  domain-free: web, logger, otel, keystore, docker
zarf/        Docker, Compose, Kind, Helm, keys (never imported)
```

- Imports point down only: `api → app → business → foundation`.
- `foundation/` never imports the project and never logs.
- App domain packages never import each other; an app package may call several
  business packages.
- Business domain imports go one way only; use the delegate for the reverse.

## Request lifecycle

`web.App.ServeHTTP` (CORS, preflight, HSTS, `path.Clean`) → `otelhttp` →
`http.ServeMux` → global middleware `Otel → Logger → Errors → Metrics → Panics`
→ route middleware (`Bearer`, `Authorize*`, optional `BeginCommitRollback`) →
handler → business (extensions → core → store).

- Handler signature: `func(ctx context.Context, r *http.Request) web.Encoder`.
- Success values and `*errs.Error` both implement `Encoder`.
- `web.Respond` writes the response: `HTTPStatus()` if present, 500 for other
  errors, 204 for `nil`.
- `Panics` turns a panic into an `InternalOnlyLog` error; `Errors` logs once and
  hides internal messages; `Logger` logs the final status.

## Data models

| Layer | Example | Types | Tags |
|---|---|---|---|
| app | `userapp.NewUser`, `userapp.User` | primitives | `json` |
| business | `userbus.User`, `NewUser`, `UpdateUser` | strong types | none |
| storage | `userdb.userDB` | SQL types, `sql.Null*`, `dbarray` | `db` |

- Converters: `toBus<Type>` (app → business, collect `errs.FieldErrors`),
  `toApp<Type>`, `toDB<Type>`, storage `toBus<Type>` (parses again).
- Strong types: unexported field, `Parse`, `MustParse`, `String`, `Equal`,
  `MarshalText`. Enums use a private registry.
- `Update<Type>` and `QueryFilter` use pointer fields for "not provided".
- Order fields go app name → opaque business constant → column name.

## Business package

Files: `<x>bus.go` (sentinel errors, `Storer`, `ExtBusiness`, `Extension`,
`Business`, `NewBusiness`, `NewWithTx`), `model.go`, `filter.go`, `order.go`,
`event.go`, `testutil.go`, `stores/<x>db/`, `extensions/<x>otel/`.

- `NewBusiness` applies extensions in reverse: the first listed is outermost.
- Every extension implements every `ExtBusiness` method.
- Stores hold `sqlx.ExtContext`, so the same code runs in or out of a
  transaction; `NewWithTx` swaps in the transaction.
- Stores translate `sqldb.ErrDBNotFound` into the domain's `ErrNotFound`.
- The business layer assigns IDs and timestamps.

## App package

Files: `route.go` (`Config` struct + `Routes`), `<x>app.go` (handlers on an
unexported `app`), `model.go`, `filter.go`, `order.go`.

- Handler recipe: `web.Decode` → `toBus*` → business call → map errors with
  `errors.Is` → `toApp*`.
- Entity middleware (`mid.AuthorizeUser`, `mid.AuthorizeProduct`) loads the
  entity, checks the rule, and stores it in the context.
- Context values use typed `setX`/`GetX` helpers with a private key type.

## Startup (`main.go`)

`main` builds the logger and calls `run`, which returns an error. `run`:
GOMAXPROCS → `conf.Parse(prefix, &cfg)` → log config (secrets masked) → DB →
business packages → keystore and `auth.New` → tracing → debug server (own mux)
→ API server with timeouts → wait for a server error or SIGINT/SIGTERM →
`Shutdown` with timeout, `Close` on failure.

## Cross-cutting

- Config: `ardanlabs/conf`, defaults in struct tags, env prefix, `mask` tag.
- Auth: RS256 JWT with `kid`; keystore from `zarf/keys` or
  `<PREFIX>_AUTH_KEYS_JSON`; OPA rules `rule_any`, `rule_admin_only`,
  `rule_user_only`, `rule_admin_or_subject`; user-enabled check on each
  authentication.
- Database: `sqlx` + `pgx`, named SQL only through `sqldb` helpers (span, query
  log on error, Postgres error translation).
- Migrations: one append-only `migrate.sql`, darwin checksums, run by
  `admin migrate` as an init container.
- Health: `/v1/liveness` and `/v1/readiness` with `HandlerFuncNoMid`.
- Logging: `slog` JSON with trace id; `logfmt` for humans.

## Testing

- Real Postgres in Docker via `dbtest.New`: one container, a fresh database per
  test, real migrations.
- Seeds go through the business API (`TestSeedUsers`, `TestGenerateSeedProducts`).
- Business tests use `unittest.Table`; API tests use `apitest.Table` through the
  real mux with real tokens.
- Name cases by outcome: `create200`, `create400`, `create401`.
