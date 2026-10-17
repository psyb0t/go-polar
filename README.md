# go-polar

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/go-polar.svg)](https://pkg.go.dev/github.com/psyb0t/go-polar)
[![CI](https://github.com/psyb0t/go-polar/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/go-polar/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/go-polar/badges/version.svg)](https://github.com/psyb0t/go-polar/tags)
[![license](https://raw.githubusercontent.com/psyb0t/go-polar/badges/license.svg)](LICENSE)
[![imported by](https://raw.githubusercontent.com/psyb0t/go-polar/badges/importers.svg)](https://github.com/psyb0t/go-polar/blob/badges/importers.md)

Unofficial Go client for the [Polar API](https://polar.sh/docs/api-reference), generated with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) from the OpenAPI document Polar publishes. I am not affiliated with Polar.

The package is `polar`, in the repository root. Everything in `client_*.gen.go` is generated. Don't edit it by hand, regenerate it.

## Install

```sh
go get github.com/psyb0t/go-polar
```

## Usage

Every REST endpoint in the spec has a typed method on `ClientWithResponses`. Responses decode into `JSON200`, `JSON201` and the other status fields.

```go
package main

import (
	"context"
	"net/http"
	"os"

	"github.com/psyb0t/go-polar"
)

func main() {
	auth := polar.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set(
				"Authorization",
				"Bearer "+os.Getenv("POLAR_ACCESS_TOKEN"),
			)

			return nil
		},
	)

	client, err := polar.NewClientWithResponses(
		"https://api.polar.sh",
		auth,
	)
	if err != nil {
		panic(err)
	}

	res, err := client.ProductsListWithResponse(
		context.Background(),
		&polar.ProductsListParams{},
	)
	if err != nil {
		panic(err)
	}

	_ = res.JSON200
}
```

The same code is in `example_test.go`, so `make test` checks that it compiles. Check the response status and success field before using the decoded body. HTTP error statuses do not automatically become Go errors.

The full API is in the [Go reference](https://pkg.go.dev/github.com/psyb0t/go-polar); from inside your module, `go doc github.com/psyb0t/go-polar` prints the same.

Method names come from operation IDs: `products:list` becomes `ProductsListWithResponse`, and `customers:create` becomes `CustomersCreateWithResponse`.

## Auth

Set `Authorization: Bearer <token>` through `WithRequestEditorFn`, as in the example. Use a personal access token, an organization access token or an OAuth access token with the scopes required by the operation. Customer portal and member operations use their corresponding session tokens. The generated client does not fetch or refresh tokens.

| Environment | Base URL |
|---|---|
| Production | `https://api.polar.sh` |
| Sandbox | `https://sandbox-api.polar.sh` |

Paths already include `/v1`; pass the host URL without adding that prefix.

## Regenerating

The spec is not committed. `make generate` downloads the newest effective OpenAPI document into the gitignored `api/` directory and regenerates the client:

```sh
make generate
```

`gen.go` declares the steps:

- `scripts/fetch-spec.sh` lists the date-versioned JSON files in [Polar's spec directory](https://github.com/polarsource/polar/tree/main/docs/openapi). It selects the greatest `YYYY-MM` version that is no later than the current UTC month. Future versions are excluded.
- `scripts/patch-spec.sh` maps scalar-or-array query filters to slices. The generator otherwise emits anonymous structs with inaccessible union fields. A one-element slice sends one value; longer slices send repeated query keys. It also normalizes provider descriptions for generated comments. It preserves every REST operation.
- oapi-codegen runs through `go tool` with `config.yaml`. Number fields map to `float64`. `scripts/split-gen.sh` splits the output at declarations and runs the pinned `goimports` tool.

Empty JSON response schemas map to `json.RawMessage`, preserving their bytes without decoding an unspecified structure.

`.github/workflows/regen.yml` does the same every day in CI. When the new spec changes the client, it runs `make lint`, `make test` and `make sec`. If a check fails, Claude edits the code and the checks run again, up to three times, and a review step reads the result. A passing change is committed with a CHANGELOG entry and tagged as the next release: a patch version for fixes and additions, a minor version for breaking changes.

## Agent integrations

The [skill](.agents/skills/go-polar) teaches a coding agent how to use go-polar when it writes Go. It works in any agent that reads `.agents/skills/`, and installs natively in the clients below.

### Claude Code

```bash
claude plugin marketplace add psyb0t/agents
claude plugin install go-polar@psyb0t
```

### Codex

```bash
codex plugin marketplace add psyb0t/agents
codex plugin add go-polar@psyb0t
```

Codex also picks the skill up with no install in any checkout of this repo.

### OpenClaw

The skill is published to ClawHub on every release:

```bash
openclaw skills install @psyb0t/go-polar
```

## Development

Every command runs in the `Dockerfile.dev` image. The host needs only Docker and Make.

```sh
make help      # list targets
make dep       # go mod tidy + vendor
make lint      # shellcheck, shfmt, go fix, golangci-lint
make test      # go test -race
make sec       # gitleaks, govulncheck, semgrep
make generate  # download the spec and regenerate
```

## License

[MIT](LICENSE). The generated code comes from Polar's published API specification. The specification itself is not redistributed here. The spec does not declare a license or terms URL.
