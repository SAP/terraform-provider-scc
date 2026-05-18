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
- **`scc/provider/provider.go`** - Provider registration (`Resources`, `DataSources`, `ListResources`, `Actions`)
- **`scc/provider/resources/`** - Resource implementations; fixtures in `resources/fixtures/`
- **`scc/provider/datasources/`** - Data source implementations; fixtures in `datasources/fixtures/`
- **`scc/provider/listresources/`** - List resource implementations; fixtures in `listresources/fixtures/`
- **`scc/provider/actions/`** - Action implementations (e.g., `scc_generate_csr`)
- **`scc/provider/helpers/`** - Shared API helpers (`helper.go`, `helper_certificate.go`)
- **`scc/provider/model/`** - Terraform model types shared across resources/datasources
- **`scc/provider/tfutils/`** - Test utilities: VCR setup, provider factories, redaction hooks
- **`internal/api/`** - REST API client (`RestApiClient`) with Basic Auth and mTLS support
- **`internal/api/apiObjects/`** - Go structs mapping to SCC API JSON responses
- **`internal/api/endpoints/`** - URL path builders for each API resource
- **`validation/`** - Custom Terraform validators (UUID, system mapping fields)

### Provider Configuration

The provider accepts Basic Auth OR mTLS certificate auth (mutually exclusive). Configuration values resolve from HCL attributes first, then environment variables: `SCC_INSTANCE_URL`, `SCC_USERNAME`, `SCC_PASSWORD`, `SCC_CA_CERTIFICATE`, `SCC_CLIENT_CERTIFICATE`, `SCC_CLIENT_KEY`.

The provider implements `ProviderWithListResources` and `ProviderWithActions` from the Terraform Plugin Framework in addition to the standard resource/datasource interfaces.

### Resource Implementation Pattern

Each resource follows this structure (example: `resource_subaccount.go`):
1. Struct holding `*api.RestApiClient`
2. `New*Resource()` constructor registered in `provider.go`'s `Resources()` method
3. Implements: `Metadata`, `Schema`, `Configure`, `Create`, `Read`, `Update`, `Delete`, `ImportState`
4. Uses helpers from `scc/provider/helpers/` for API calls
5. API objects from `internal/api/apiObjects/` and endpoint paths from `internal/api/endpoints/`
6. Terraform model types from `scc/provider/model/`

Data sources and list resources follow the same pattern but are read-only.

### Testing with VCR (go-vcr)

Tests use HTTP record/replay via `go-vcr`. Fixtures are YAML cassettes stored alongside their test package:
- `scc/provider/resources/fixtures/`
- `scc/provider/datasources/fixtures/`
- `scc/provider/listresources/fixtures/`

- **Replay mode (default):** Tests run offline using recorded fixtures. No env vars needed.
- **Record mode:** Set `TEST_RECORD=true` plus `SCC_USERNAME`, `SCC_PASSWORD`, `SCC_INSTANCE_URL` (and `TF_VAR_*` vars for specific resources). Sensitive data is automatically redacted by hooks in `tfutils/vcr.go`.

Key test infrastructure in `scc/provider/tfutils/`:
- `tfutils.SetupVCR(t, cassetteName)` - Creates recorder, returns `(*recorder.Recorder, User)`
- `tfutils.GetTestProviders(httpClient)` - Builds provider factories for tests
- `tfutils.ProviderConfig(user)` - Generates HCL provider block
- `tfutils.StopQuietly(rec)` - Deferred recorder cleanup

Test pattern:
```go
func TestResource*(t *testing.T) {
    t.Parallel()
    t.Run("happy path", func(t *testing.T) {
        rec, user := tfutils.SetupVCR(t, "fixtures/<cassette_name>")
        defer tfutils.StopQuietly(rec)
        resource.Test(t, resource.TestCase{
            IsUnitTest:               true,
            ProtoV6ProviderFactories: tfutils.GetTestProviders(rec.GetDefaultClient()),
            Steps: []resource.TestStep{
                { Config: tfutils.ProviderConfig(user) + ..., Check: ... },
                { ResourceName: ..., ImportState: true, ImportStateVerify: true },
            },
        })
    })
}
```

### CI Checks

CI validates: build, lint, `go fix`, `go generate` (doc drift), fixture drift detection, and tests across Terraform 1.13-1.15. Fixture changes must be committed; CI detects untracked modifications.

### Documentation Generation

Docs in `docs/` are auto-generated from schema via `terraform-plugin-docs`. Run `make generate` and commit results. Do not edit `docs/` files directly; modify schema `MarkdownDescription` fields or templates in `templates/` instead.

### Dependencies

Vendored in `vendor/`. After modifying `go.mod`, run `go mod vendor` to update.
