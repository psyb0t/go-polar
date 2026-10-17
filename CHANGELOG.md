# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.1, 2026-10-17

- Documentation only. The agent skill explains how to look up every endpoint, parameter and model with `go doc` or on pkg.go.dev, and declares its license.
- The skill no longer lists the API credentials under OpenClaw's `requires.env`, which kept OpenClaw from loading it until they were set. It requires only `go`; the credentials are still documented as what your program reads.
- The README points to the full API reference, and the examples in the README and the skill were checked against the current code.

## v0.1.0, 2026-10-17

- Added `polar`, a REST client generated with oapi-codegen from Polar's OpenAPI 3.1.0 document, API version `2026-10`, at https://raw.githubusercontent.com/polarsource/polar/main/docs/openapi/2026-10.openapi.json. It covers all 190 REST operations across 128 paths, including products, customers, subscriptions, checkouts, orders, benefits, usage events and webhook management.
- Number fields are `float64` instead of oapi-codegen's default `float32`.
- `make generate` selects the newest spec version already effective in the current UTC month, normalizes scalar-or-array query filters and regenerates the client. Future versions are excluded, and the spec is not committed.
