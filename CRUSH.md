# onvif-streams Development Guide

## Build / Test / Lint

- **Build**: `go build .` or `go run .`
- **Test**: `go test ./...` – run a specific test with `-run`.
- **Golden**: `go test ./... -update` to regenerate golden files.
- **Lint**: Run your linter directly (e.g., `golangci-lint run`).
- **Dev**: `go run . ./dev` (profiling enabled).

## Code Style

- Imports: use `goimports` to sort stdlib, external, internal.
- Formatting: `gofmt`.
- Naming: exported – PascalCase, unexported – camelCase.
- Types: explicit, use type aliases.
- Errors: return explicitly, wrap with `fmt.Errorf`.
- Context: first argument for operations.
- Interfaces: minimal, defined in consuming packages.
- Consts: typed, iota.
- JSON tags: snake_case.
- File perms: octal (0o755, 0o644).
- Comments: end with period.

## Testing

- Use `require` from testify.
- Parallel tests: `t.Parallel()`.
- Env vars: `t.SetEnv`.
- Temp dirs: `t.TempDir()`.

## Mock Providers

```go
func TestYourFunction(t *testing.T) {
    // Enable mock providers for testing
    orig := config.UseMockProviders
    config.UseMockProviders = true
    defer func() {
        config.UseMockProviders = orig
        config.ResetProviders()
    }()
    config.ResetProviders()
    providers := config.Providers()
    // test logic
}
```

No cursor or Copilot rules.
