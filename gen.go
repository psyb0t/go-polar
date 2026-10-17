// Package polar is a Go client for the Polar API, generated from its
// public OpenAPI document with oapi-codegen.
package polar

//go:generate bash scripts/fetch-spec.sh
//go:generate bash scripts/patch-spec.sh
//go:generate go tool oapi-codegen --config=config.yaml api/openapi.patched.json
//go:generate bash scripts/split-gen.sh
