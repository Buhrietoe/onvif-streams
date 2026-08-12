# onvif-streams Agent Guide

## Overview
CLI tool that discovers ONVIF IP cameras via UDP multicast and outputs JSON or table with device metadata and RTSP stream URIs. Single-package Go project: `main.go`, `device.go`, `discovery.go`, `types.go`.

## Commands
```bash
go build .             # Build binary
go run .               # Network discovery mode
go run . -a HOST:PORT  # Target single device
go test ./...          # Unit + integration tests (no network needed)
```

## CLI Flags
- `-a HOST:PORT` — target single device (e.g. 192.168.1.100:8080); omit for auto-discovery
- `-f json|table` — output format (default: json)
- `-n` — names only (implies -f table)
- `-o FILE` — write to file instead of stdout
- `-r N` — SOAP call retries (default: 0)
- `-t DURATION` — UDP discovery timeout (default: 2s)
- `-u USER` / `-p PASS` — credentials (default: admin/admin, overridable via env)

## Environment Variables
- `ONVIF_USER` — default username
- `ONVIF_PASS` — default password
- Flags `-u`/`-p` override env vars; env vars override hardcoded "admin".

## Architecture
- `main.go`: CLI entrypoint. Parses flags into `Config`, routes to discovery or single-device mode, outputs via `json.NewEncoder` or `printTable`.
- `discovery.go`: WS-Discovery probe on UDP `239.255.255.250:3702`. Probes three types: `NetworkVideoTransmitter`, `NetworkVideoRecorder`, `MediaServer`. Parses `XAddrs` from `ProbeMatch` responses. UUID via `github.com/gofrs/uuid` (RFC4122).
- `device.go`: Uses `github.com/use-go/onvif` v0.0.9. Calls SOAP methods via `dev.CallMethod`: `GetDeviceInformation`, `GetProfiles`, `GetStreamUri`. Returns `*DiscoveredDevice`.
- `types.go`: XML envelope structs for ONVIF responses + JSON output structs.

## Critical Patterns
- **Error handling**: `processDevice` returns `nil` only if device info fails. Partial success is intentional: device returned even if profiles/stream URIs fail.
- **XML decoding**: Always check both decode error AND response body nil:
  ```go
  if err := xml.NewDecoder(resp.Body).Decode(&env); err != nil || env.Body.GetDeviceInformationResponse == nil {
      return nil
  }
  ```
- **Stream URI fallback**: `GetStreamUri` failure leaves profile without `StreamURI`; profile still appended.
- **JSON tags**: All output structs use `snake_case` tags (`firmware_version`, `stream_uri`, etc.).

## Tests
- `onvif_streams_test.go` — unit tests: XML unmarshaling, JSON tags, `printTable`, `envDefault`.
- `integration_test.go` — integration tests using fixtures in `testdata/` (no network).
- Run: `go test ./...`

## Dependencies
- `github.com/use-go/onvif` v0.0.9 — ONVIF protocol implementation
- `github.com/gofrs/uuid` v4.4.0 — UUID generation
- Go 1.25.0 required (go.mod)

## Environment
Requires local network access for discovery or targeting:
- UDP multicast: `239.255.255.250:3702` for discovery
- Target device HTTP port for SOAP calls
