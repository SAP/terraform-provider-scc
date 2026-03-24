# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terraform Provider for SAP Cloud Connector (`scc`). Manages SAP Cloud Connector instances via their REST API using the HashiCorp Terraform Plugin Framework (Protocol v6). Registry address: `registry.terraform.io/sap/scc`.

## Build & Development Commands

```bash
make build          # Build the provider
make install        # Build and install to GOBIN
make lint           # Run golangci-lint
make fmt            # Format code with gofmt
make generate       # Regenerate docs via terraform-plugin-docs
make test           # Run all tests (parallel=4, timeout=900s)
make testacc        # Run acceptance tests (TF_ACC=1)
```

Run a single test:
```bash
go test -v -run TestResourceSubaccount ./scc/provider/
```

Run a specific sub-test:
```bash
go test -v -run "TestResourceSubaccount/happy_path" ./scc/provider/
```

## Architecture

### Package Layout

- **`main.go`** - Entry point; `go:generate` directives for `terraform-plugin-docs`
- **`scc/provider/`** - All provider logic (resources, data sources, list resources, actions, tests)
- **`internal/api/`** - REST API client (`RestApiClient`) with Basic Auth and mTLS support
- **`internal/api/apiObjects/`** - Go structs mapping to SCC API JSON responses
- **`internal/api/endpoints/`** - URL path builders for each API resource
- **`validation/`** - Custom Terraform validators (UUID, system mapping fields)

### Provider Configuration

The provider (`cloudConnectorProvider`) accepts Basic Auth OR mTLS certificate auth (mutually exclusive). Configuration values resolve from HCL attributes first, then environment variables: `SCC_INSTANCE_URL`, `SCC_USERNAME`, `SCC_PASSWORD`, `SCC_CA_CERTIFICATE`, `SCC_CLIENT_CERTIFICATE`, `SCC_CLIENT_KEY`.

### Resource Implementation Pattern

Each resource follows this structure (example: `resource_subaccount.go`):
1. Struct holding `*api.RestApiClient`
2. `New*Resource()` constructor registered in `provider.go`'s `Resources()` method
3. Implements: `Metadata`, `Schema`, `Configure`, `Create`, `Read`, `Update`, `Delete`, `ImportState`
4. Uses `sendRequest()` or `requestAndUnmarshal[T]()` from `helper.go` for API calls
5. API objects from `internal/api/apiObjects/` and endpoint paths from `internal/api/endpoints/`

Data sources and list resources follow the same pattern but are read-only.

### Testing with VCR (go-vcr)

Tests use HTTP record/replay via `go-vcr`. Fixtures are YAML cassettes stored in `scc/provider/fixtures/`.

- **Replay mode (default):** Tests run offline using recorded fixtures. No env vars needed.
- **Record mode:** Set `TEST_RECORD=true` plus `SCC_USERNAME`, `SCC_PASSWORD`, `SCC_INSTANCE_URL` (and `TF_VAR_*` vars for specific resources). Sensitive data is automatically redacted by hooks in `provider_test.go`.

Key test infrastructure in `provider_test.go`:
- `setupVCR(t, cassetteName)` - Creates recorder, returns `(*recorder.Recorder, User)`
- `getTestProviders(httpClient)` - Builds provider factories for tests
- `providerConfig(user)` - Generates HCL provider block
- `stopQuietly(rec)` - Deferred recorder cleanup

Test pattern:
```go
func TestResource*(t *testing.T) {
    t.Parallel()
    t.Run("happy path", func(t *testing.T) {
        rec, user := setupVCR(t, "fixtures/<cassette_name>")
        defer stopQuietly(rec)
        resource.Test(t, resource.TestCase{
            IsUnitTest:               true,
            ProtoV6ProviderFactories: getTestProviders(rec.GetDefaultClient()),
            Steps: []resource.TestStep{
                { Config: providerConfig(user) + ..., Check: ... },
                { ResourceName: ..., ImportState: true, ImportStateVerify: true },
            },
        })
    })
}
```

### CI Checks

CI validates: build, lint, `go fix`, `go generate` (doc drift), fixture drift detection, and tests across Terraform 1.12-1.14. Fixture changes must be committed; CI detects untracked modifications.

### Documentation Generation

Docs in `docs/` are auto-generated from schema via `terraform-plugin-docs`. Run `make generate` and commit results. Do not edit `docs/` files directly; modify schema `MarkdownDescription` fields or templates in `templates/` instead.

### Dependencies

Vendored in `vendor/`. After modifying `go.mod`, run `go mod vendor` to update.
