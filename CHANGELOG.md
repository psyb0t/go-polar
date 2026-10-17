# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.0, 2026-10-17

- Added `polar`, a REST client generated with oapi-codegen from Polar's OpenAPI 3.1.0 document, API version `2026-10`, at https://raw.githubusercontent.com/polarsource/polar/main/docs/openapi/2026-10.openapi.json. It covers all 190 REST operations across 128 paths, including products, customers, subscriptions, checkouts, orders, benefits, usage events and webhook management.
- Number fields are `float64` instead of oapi-codegen's default `float32`.
- `make generate` selects the newest spec version already effective in the current UTC month, normalizes scalar-or-array query filters and regenerates the client. Future versions are excluded, and the spec is not committed.
