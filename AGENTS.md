# Microcloud Cluster Manager

Management system for MicroCloud clusters: a Go backend with a REST API and a React UI (`ui/`). One binary (`cmd/app`) runs as the `management-api`, the `cluster-connector` or the `cli`, chosen with the `SERVICE` environment variable. Backend configuration comes from environment variables (`internal/pkg/config/config.go`).

More detail: [ARCHITECTURE.md](ARCHITECTURE.md), [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md).

## Rules

1. **Make the smallest change that does the job.** Follow the patterns that are already in the repository. Ask before introducing a new pattern, environment variable, script, Make target, CI job or config option. Do not refactor or optimize what the task does not touch.
2. **Go comments end with a period** (the `godot` linter) and exported identifiers have a comment. Wrap errors with `%w` and log with `logger.Log.Errorw` and structured fields.
3. **Keep the backend and the UI in sync.** Types in `ui/src/types/` mirror the API models in `internal/pkg/api/models/`.
4. **Never commit secrets, private keys or generated test certificates.** Keep machine-specific overrides in `.env.local` (gitignored).
5. **Commits** are signed off (`git commit -s`, CI checks the DCO) and use the style `type(scope): subject`, for example `fix(ui): sort cluster tables`.
6. Before you finish, run what applies to your change: `make test-unit`, `go fmt ./...`, `make lint-backend`, and in `ui/` `yarn lint-js` and `yarn test-unit`.

## Commands

| Task | Command |
| --- | --- |
| Start the backend (Postgres and Prometheus in Docker, both services) | `make dev` |
| Start the UI (once: `sudo make add-hosts`) | `make ui`, or `cd ui && PORT=8414 yarn start` |
| Build the binary (also builds and embeds the UI) | `make build` |
| Backend unit tests | `make test-unit` |
| Backend E2E tests (need `make dev`) | `make test-e2e` |
| OpenAPI tests (need `make dev` and network) | `make test-openapi` |
| Lint and format the backend | `make lint-backend`, `go fmt ./...` |
| Lint the UI (`tsc`, ESLint, Prettier) and format it | `cd ui && yarn lint-js`, `yarn format-js` |
| UI unit tests | `cd ui && yarn test-unit` |
| UI E2E tests (need `make dev`) | `make test-ui-e2e` (optional `PROJECT=chromium`) |
| Git hooks (golangci-lint, lint-staged) | `make add-hooks` |

Addresses: management API `https://ma.lxd-cm.local:30000`, cluster connector `https://cc.lxd-cm.local:9000`, UI dev server `https://ma.lxd-cm.local:8414`, database `127.0.0.1:5432`.

Do not use `go test ./...` for a quick check. It also runs `test/e2e` and `test/openapi`, which fail without the running backend.

## How a request reaches a cluster

The browser talks to the management API, which authenticates the user with OIDC. For calls about one MicroCloud, the management API forwards the request to the cluster connector, which sends it to the MicroCloud over the WebSocket reverse tunnel that the MicroCloud opened (`tunnel_registered`). There is no direct connection to the MicroCloud. Without a tunnel the API returns "remote cluster does not have an active tunnel". MicroClouds enroll with a join token and authenticate to the cluster connector with mTLS.

## Backend

- **Layout**: `internal/app/<service>/api/v1` (handlers), `internal/app/<service>/core` (logic), `internal/pkg` (shared: `api`, `config`, `database`, `logger`, `middleware`, `types`).
- **Endpoints** are `types.Endpoint` structs (handler, method, path, `AllowUnauthorized`, `AllowedEntitlements`), grouped in a `types.RouteGroup` per file in `internal/app/<service>/api/v1/`. The `RouteGroup` is listed in that service's `api/routes.go`. Handlers return `error` and respond with LXD's `github.com/canonical/lxd/lxd/response` package, for example `return response.SmartError(err).Render(w, r)`.
- **Authorization** runs in `registerEndpoint` (`internal/pkg/api/rest.go`) after authentication and before the handler, through the `Authorizor` interface in `internal/pkg/types/rest.go`. It returns 403 when it fails.
  - Management API: OIDC authentication. Every protected endpoint requires the `admins` group, read from the `mcm-idp-groups` claim into `UserInfo` (`ctx.Value(types.UserInfoKey)`). Set `AllowUnauthorized` only for public endpoints.
  - Cluster connector: mTLS authentication with certificate pinning against the database. Every authenticated client is authorized.
- **Database**: PostgreSQL through sqlx. Queries live in `internal/pkg/database/store`. Schema changes are numbered Goose files in `internal/pkg/database/schema/migrations`. `VersionMap` in `schema.go` maps each release version to a migration number, so a release that adds a migration needs a new entry. The `development` version applies all migrations.
- **API contract**: MicroCloud calls these APIs. `make test-openapi` checks them against the spec in `canonical/microcloud` (`doc/_extra/cluster-manager-api.yaml`), so API changes must stay compatible with it.
- **Embedded UI**: `make build` builds the UI and copies it to `internal/app/management-api/api/v1/static`, which the management API embeds and serves under `/ui`. That folder is gitignored; never edit or commit it.

## Frontend

- **`ui/src/api/`**: one file per resource with the fetch functions. **`ui/src/context/use*.tsx`**: React Query hooks (`useQuery`, `useMutation`) that use the keys in `ui/src/util/queryKeys.tsx`. Invalidate the related keys when a mutation succeeds.
- **`ui/src/pages/`**: one folder per feature. **`ui/src/components/`**: shared components. Prefer `@canonical/react-components`, with SCSS and BEM class names for custom styles.
- Use the path aliases from `ui/tsconfig.json` (`import { useAuth } from 'context/auth'`).
- **Adding a feature**: types, API functions, hook, components, page, then tests.

## Testing

Follow LXD for the backend tests and lxd-ui for the UI tests, and keep the two separate.

**Backend** (`test/`): a test is a function of type `types.Test` (`types.UnitTest` for unit tests, see `test/types/types.go`) in `test/e2e/` or `test/unit/`. Add it to the `tests` list in `e2e_test.go` or `unit_test.go`, which is the only runner, like LXD's `test_*` functions. Shared helpers are in `test/helpers/`. Most E2E tests register synthetic clusters through the API and delete what they create.

**UI unit tests**: `*.spec.ts(x)` next to the source, with Vitest.

**UI E2E tests** (`ui/tests/*.spec.ts`, Playwright): one spec per feature, helpers in `ui/tests/helpers/`, in the style of lxd-ui.
- Import `test` and `expect` from `./fixtures/test-extension`. It starts the coverage collection.
- Write top-level tests. Keep waiting and retry logic in helpers and use plain `await expect(locator).toBeVisible()` in the spec.
- Run shell commands with `runCommand` from `ui/tests/helpers/shell.ts`.
- First run: `cd ui && npx playwright install`, and create `ui/.env.local` with `OIDC_USER="cluster-manager-e2e-tests@example.org"` and `OIDC_PASSWORD="cluster-manager-e2e-password"`.

**Keep tests simple**: no extra environment variables, scripts, Make targets, CI jobs, Playwright projects or config options. CI runners are ephemeral, so do not add cleanup, secret redaction or isolation for them. Do not make a test depend on whether a cluster is synthetic or real.

### E2E fixtures

- The development build seeds the synthetic clusters `cluster-01` to `cluster-20`. The seed runs unless the `VERSION` environment variable is set to something other than `development`.
- The tunnel tests use real MicroClouds that `scripts/mcm-cluster-onboard.sh` creates in LXD containers. Their names are constants in `test/e2e/remote_cluster_tunnel.go` and `ui/tests/helpers/remote-clusters.ts`.
  - `mcm-e2e-01` and `mcm-e2e-02`, enrolled with the reverse tunnel enabled: `./scripts/mcm-cluster-onboard.sh --count 2 --instance-prefix mcm-e2e`
  - `mcm-e2e-join-01`, not enrolled: `./scripts/mcm-cluster-onboard.sh --count 1 --instance-prefix mcm-e2e-join --provision-only`. The UI join test enrolls it, so the test runs once for each enrollment. To run it again, remove the cluster in the UI and run `lxc --project microcloud-mcm exec mcm-e2e-join-01 -- microcloud cluster-manager delete --force`.
- `CLI_BIN` sets the binary the script uses (default `cmd/app`). `scripts/run-coverage.sh` sets it to `cmd/app-coverage`, which `make build-coverage` writes.

## CI

`.github/workflows/pr.yaml` runs on GitHub-hosted runners and pins actions by commit SHA (see ARCHITECTURE.md for the job graph). The E2E jobs set up LXD with `canonical/setup-lxd`, start the backend, onboard the real MicroClouds with `scripts/mcm-cluster-onboard.sh` and then run the tests. `.github/workflows/tics.yaml` runs `scripts/run-coverage.sh` on pushes to `main`, with the same LXD setup and onboarding.

## Local machine caveats

- The backend uses the fixed ports 30000, 9000, 5432 and 9090. `scripts/setup-env.sh` force-removes the `my-postgres` and `prometheus` containers.
- The machine may already have clusters in the `microcloud-mcm` LXD project. Do not run `--nuke` or `--re-register` without asking.
- Each git worktree has its own `test/keys` and needs its own `yarn install` in `ui/`. Enrolling clusters with a backend from another worktree fails on a certificate mismatch (see `CONTRIBUTING.md`).
