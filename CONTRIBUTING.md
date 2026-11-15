# Contributing to qrocodile-cli

Notes for working on the CLI itself.

```bash
make build
make test
make check   # fmt, vet, lint, build, test — what CI runs
```

## Running it locally

```bash
go build -o qrocodile ./cmd/qrocodile
QROCODILE_API_KEY=qk_live_… ./qrocodile generate "https://example.com" -o code.svg
```

Or skip the build step with `go run`:

```bash
QROCODILE_API_KEY=qk_live_… go run ./cmd/qrocodile generate "https://example.com" -o code.svg
```

`QROCODILE_API_BASE_URL` points at a different API instance (a dev/staging deployment, say) instead of the default. The tests read a different pair, `QR_API_KEY`/`QR_API_URL` (see Tests below).

## Layout

| Path                     | What it is                                                                 |
| ------------------------- | --------------------------------------------------------------------------- |
| `cmd/qrocodile/main.go`  | Entry point; calls `cmd.Execute()`. It has its own subdirectory so `go install ./cmd/qrocodile` produces a binary named `qrocodile`: `go install` names the binary after the package directory, and the repository root would give `qrocodile-cli`. |
| `cmd/*.go` (root of `cmd/`) | The command tree (Cobra): `root.go`, `auth*.go`, `generate.go`, `version.go`. A separate package (`cmd`) from `cmd/qrocodile`. |
| `internal/config/`       | The local config file, which holds the stored API key as a typed credential record. |
| `internal/apiclient/`    | Resolves a `qrocodile-api-go` client from the env vars and the config file. |
| `internal/output/`       | The `-o`/no-`-o` writing rules (curl-style binary-to-terminal guard).      |
| `internal/browser/`      | Opens the QR Designer link for `generate --open`.                          |
| `Makefile`                | `make check` is what CI runs; see individual targets for the rest.        |
| `.golangci.yml`          | Excludes `fmt.Fprint*` from `errcheck`; the comment in the file says why. |
| `tools/go.mod`           | `golangci-lint`’s own module (see Linting below).                          |

## Depending on qrocodile-api-go

This module depends on [`qrocodile-api-go`](https://github.com/qrocodile-io/qrocodile-api-go) as an ordinary, published `go.mod` dependency, without a `replace` directive. Bump it the normal way:

```
go get github.com/qrocodile-io/qrocodile-api-go@<version>
```

If you’re changing both repos at once from sibling checkouts, add a local `replace` temporarily (`go mod edit -replace github.com/qrocodile-io/qrocodile-api-go=../qrocodile-api-go`), and remove it again before committing so this module keeps resolving from the published version for everyone else.

## Tests

```bash
make test
```

Two layers:

- **Unit:** table-driven, and an `httptest.Server` for anything that talks to the API (`cmd/generate_run_test.go`). No network by default.
- **Integration** (`cmd/integration_test.go`): hits a live API and self-skips unless `QR_API_KEY` is set:

  ```bash
  QR_API_KEY=qk_live_… make test
  # point at another environment:
  QR_API_KEY=qk_live_… QR_API_URL=http://localhost:3002 make test
  # fail instead of skip when the key is missing (what CI runs outside fork PRs):
  QR_API_KEY=qk_live_… make test-integration
  ```

  `QR_API_KEY`/`QR_API_URL` (not `QROCODILE_API_KEY`/`QROCODILE_API_BASE_URL`) are the test-harness convention shared with `qrocodile-api-go`, so both repos read the same local `.env` during development; the tests translate them into this CLI’s real env vars internally.

## Linting

```bash
make lint
```

`golangci-lint` is a tool dependency pinned in `tools/go.mod`, a separate module, as in `qrocodile-api-go`. That keeps its Go-version requirement from raising the floor this module needs, and keeps its large dependency tree out of this module’s `go.mod`/`go.sum`. `make lint` builds it into `tools/bin/` on first use (or whenever `tools/go.mod`/`tools/go.sum` change) and reuses the binary after that.

## Releasing

1. Move the `## Unreleased` section in `CHANGELOG.md` to a new `## X.Y.Z` heading.
2. Commit that, then tag and push:

   ```bash
   git tag v0.1.0
   git push --tags
   ```

Pushing a `v*` tag triggers `.github/workflows/release.yml`, which re-runs `make check` against that exact commit, then runs [GoReleaser](https://goreleaser.com) (config in `.goreleaser.yml`) to cross-compile binaries for darwin/linux/windows × amd64/arm64, publish a GitHub release with checksummed archives attached, and push an updated formula to [`qrocodile-io/homebrew-tap`](https://github.com/qrocodile-io/homebrew-tap). That last step needs a `HOMEBREW_TAP_GITHUB_TOKEN` repository secret (a fine-grained PAT scoped to that repository’s contents), because the default `GITHUB_TOKEN` has no access outside this repository.

Try a release locally before trusting a real tag to it:

```bash
goreleaser release --snapshot --clean
```

Builds everything into `dist/` (gitignored) without creating a release, pushing a tag, or touching the tap repo — the safe way to catch a config mistake.

Pre-1.0, so a minor bump may carry breaking changes while the surface settles.
