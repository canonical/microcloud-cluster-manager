# Management API development CLI

`mcm-dev` is a development and testing client for management-api. It provides
only `query` and `auth`: new endpoints and changing payloads are supplied at the
command line, without adding resource commands, client models, or renderers.
It talks to management-api over HTTPS, including for forwarded cluster requests.
The existing database-backed enrollment CLI in `cmd/cli` is separate and unchanged.

## Build and run

From the repository root, use the Go version in `go.mod` to build `mcm-dev`
into your user bin directory and add it to your `PATH`.

```sh
mkdir -p "$HOME/.local/bin"
go build -o "$HOME/.local/bin/mcm-dev" ./dev-cli
export PATH="$HOME/.local/bin:$PATH"
mcm-dev --help
```

Add the `export` line to `~/.bashrc` to use `mcm-dev` in new Bash terminals.
You can also run it from the repo root with `go run ./dev-cli --help`.

In GitHub Actions, add the binary directory to `GITHUB_PATH` so later steps in
the job can call `mcm-dev`. Place these steps after checkout and Go setup.

```yaml
- name: Build mcm-dev
  run: |
    mkdir -p "$RUNNER_TEMP/bin"
    go build -o "$RUNNER_TEMP/bin/mcm-dev" ./dev-cli
    printf '%s\n' "$RUNNER_TEMP/bin" >> "$GITHUB_PATH"

- name: Check mcm-dev
  run: mcm-dev --help
```

Start the usual development services with `make dev`. The default server is
`https://ma.lxd-cm.local:30000`. Global connection flags are:

| Flag | Behavior |
| --- | --- |
| `--server https://HOST:PORT` | Management API origin, without an API path |
| `--ca-cert FILE` | Add PEM certificates to the CLI's TLS trust pool. Defaults to `MCM_DEV_CA_CERT` |
| `--insecure` | Disable TLS verification explicitly for development |
| `--session-file FILE` | Override the saved browser session location |
| `--timeout DURATION` | Deadline for each request, default `30s` |

Set `MCM_DEV_CA_CERT` once to use the same PEM file for every CLI command in
your shell or CI job. For local `make dev`, run this from the repository root
after the backend generates its certificate.

```sh
export MCM_DEV_CA_CERT="$PWD/test/keys/ca.crt"
mcm-dev auth login
mcm-dev auth status
```

In GitHub Actions, set the variable in the job's `env` mapping. This example
assumes the backend starts from the same checkout and generates the file.

```yaml
env:
  MCM_DEV_CA_CERT: ${{ github.workspace }}/test/keys/ca.crt
```

An explicit `--ca-cert FILE` takes precedence over the environment variable.
Use `--ca-cert=` to ignore the variable for one command and use system trust.
TLS verification stays enabled with a custom CA file. `--insecure` conflicts
with a CA file supplied through either setting. No server private key,
database access, or backend configuration is needed.

## Authentication

1. Log in at `https://ma.lxd-cm.local:30000/oidc/login` in a browser.
2. In the browser's DevTools cookie storage, copy the **values** of
   `oidc_session` and `user_secret`. These cookies are HttpOnly, so use the
   storage inspector rather than `document.cookie`.
3. Import them at the two hidden prompts:

```sh
mcm-dev --ca-cert /path/to/dev-ca.crt auth login
mcm-dev --ca-cert /path/to/dev-ca.crt auth status
```

`auth login --stdin` instead reads exactly two newline-delimited values in that
order. Supply a private input stream; do not put cookie values in shell command
arguments. Credentials are verified with `GET /1.0` before saving. There is no
IdP-specific password flow, bearer-token login, or backend authentication change.

The session defaults to `$XDG_CONFIG_HOME/mcm-dev/session.json`, falling back to
`~/.config/mcm-dev/session.json`. The CLI creates the session directory with
`0700` permissions. For an existing directory, including when using
`--session-file`, saving requires a directory with `0700` permissions and rejects
a symbolic link as the immediate parent. The CLI saves the file with `0600`
permissions using atomic replacement. Credentials are bound to the exact HTTPS
origin. Never commit session files.

The server refreshes OIDC credentials and may rotate both cookies during a
request; the client saves replacements and deletions even on error responses.
Embedded session JWT expiry alone does not cause local rejection because the
server may renew it. On an unusable session, log in again and reimport both
cookies. The imported session shares its browser's lifecycle: browser logout or
renewal elsewhere may invalidate the CLI copy. Use one CLI process per file and
separate acquired sessions/files for parallel jobs.

`auth status` verifies the session and prints identity/groups as JSON.
`auth logout` attempts server logout without following redirects and removes
local credentials even when the remote call fails. Protected endpoints require
the existing `admins` group; forwarded operations also enforce LXD permissions
and capabilities.

## Query any endpoint

```text
mcm-dev query ENDPOINT [-X METHOD] [-d JSON|-] [--raw]
                       [--wait] [--wait-timeout DURATION]
mcm-dev auth login [--stdin]
mcm-dev auth status
mcm-dev auth logout
```

Supply an exact path including `/1.0` and any query string. Absolute URLs and
fragments are rejected; credentials always stay on the configured server.
GET is the default even when data is supplied. `-X`/`--request` accepts GET,
POST, PUT, PATCH, or DELETE. `-d`/`--data` accepts JSON or `-` for JSON from stdin.

```sh
# Readiness and resource inspection use the same command:
mcm-dev query /1.0/status
mcm-dev query /1.0/remote-cluster
mcm-dev query '/1.0/remote-cluster/lab-a/replicators?project=demo'

# Change an existing resource or test a new endpoint's payload directly:
mcm-dev query /1.0/remote-cluster/lab-a -X PATCH \
  -d '{"description":"Development cluster"}'
printf '%s\n' '{"description":"From stdin"}' | \
  mcm-dev query /1.0/remote-cluster/lab-a -X PATCH -d -

# Inspect the response envelope or wait for a forwarded operation:
mcm-dev query /1.0/remote-cluster --raw
mcm-dev query '/1.0/remote-cluster/lab-a/replicators/backup/state?project=demo' \
  -X PUT -d '{"action":"start"}' --wait
```

Paths, query parameters, and bodies are passed through; resource validation and
PUT/PATCH semantics belong to the API. There is no automatic fetch/merge of PUT
fields. All requests use the imported session, including queries to public
endpoints. No automatic write retries or redirect following are performed.
After an ambiguous write failure, inspect server state before retrying.

Output is formatted JSON metadata, preserving unknown fields, numeric precision,
and array order. `--raw` prints the full decoded response envelope instead.
Synchronous empty responses print `null`. API-returned secrets are not redacted.
There is no table output or `--output` flag. Diagnostics go to stderr.

`--wait` applies to `/1.0/remote-cluster/CLUSTER/...` endpoints: an asynchronous
response is polled through the management operation endpoint for that cluster,
without following a returned URL to another server. Polling is once per second,
with `--wait-timeout` defaulting to `5m`. Synchronous responses print immediately.
Final operation failure returns a nonzero code and prints the final metadata.
Interrupting or timing out a wait does not cancel the server operation.
Without `--wait`, use `--raw` to inspect the operation URL and query its management
endpoint later.

| Exit code | Meaning |
| --- | --- |
| 0 | Success, or async request accepted without waiting |
| 1 | API, transport, protocol, storage, or operation failure |
| 2 | Usage or local validation error |
| 3 | Missing/expired session or HTTP 401 |
| 4 | HTTP 403 |
| 5 | Request/wait deadline exceeded |
| 130 | Interrupted |

## Development and validation

Feature work normally needs no client changes: change the endpoint or JSON in
`query` invocations to match the API. `internal/command` owns argument parsing,
authentication commands, and JSON rendering. `internal/client` owns HTTP/envelope
handling and operation polling; `internal/session` owns cookie persistence.
There are no resource command registrations or shared application model imports
to maintain. Tests under `dev-cli/tests` exercise queries, authentication, TLS
configuration, cookie renewal, deadlines, and operation polling.

```sh
go fmt ./dev-cli/...
go test ./dev-cli/...
go vet ./dev-cli/...
go build -o /tmp/mcm-dev ./dev-cli
```

Tests use an in-memory HTTP transport by default. To exercise real local TLS
where loopback sockets are allowed:

```sh
MCM_DEV_TLS_TESTS=1 go test ./dev-cli/...
```

An optional read-only smoke test uses an already imported development session:

```sh
MCM_DEV_LIVE=1 \
MCM_DEV_SERVER=https://ma.lxd-cm.local:30000 \
MCM_DEV_SESSION_FILE=/path/to/session.json \
MCM_DEV_CA_CERT=/path/to/dev-ca.crt \
go test ./dev-cli/tests -run '^TestLiveSmoke$' -count=1
```

`MCM_DEV_CA_CERT` is the same CLI default used by ordinary commands and is
optional when the system already trusts the API certificate. In a restricted
agent sandbox, use the installed Go binary directly if its launcher is blocked,
and set `GOCACHE` to a writable location. These are environment workarounds, not
CLI requirements.
