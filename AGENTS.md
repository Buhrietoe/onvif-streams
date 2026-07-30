# onvif-streams Agent Guide

## Overview
CLI tool that discovers ONVIF IP cameras via UDP multicast and outputs JSON with device metadata and RTSP stream URIs. Single-package Go project: `main.go`, `device.go`, `discovery.go`, `types.go`.

## Commands
```bash
go build .             # Build binary
go run .                # Network discovery mode
go run . -a HOST:PORT   # Target single device (default user/pass: admin/admin)
go test ./...           # No tests currently exist
```

## Architecture
- `main.go`: CLI entrypoint. Two modes: discovery (`discoverONVIFDevices`) or targeted (`processDevice`). Outputs JSON array via `json.NewEncoder`.
- `discovery.go`: WS-Discovery probe on UDP `239.255.255.250:3702`. Fixed 2s timeout. Parses `XAddrs` from `ProbeMatch` responses.
- `device.go`: Uses `github.com/use-go/onvif` v0.0.9. Calls SOAP methods: `GetDeviceInformation`, `GetProfiles`, `GetStreamUri`. Returns `*DiscoveredDevice`.
- `types.go`: XML envelope structs for ONVIF responses + JSON output structs.

## Critical Patterns
- **Error handling**: `processDevice` returns `nil` on failure; caller skips device. Partial success is intentional: device info returned even if profiles fail.
- **XML decoding** (device.go:29): Always check both decode error AND response body nil:
  ```go
  if err := xml.NewDecoder(resp.Body).Decode(&envInfo); err != nil || envInfo.Body.GetDeviceInformationResponse == nil {
      return nil
  }
  ```
- **Stream URI fallback** (device.go:68-75): `GetStreamUri` failure leaves profile without `StreamURI`; profile still appended.
- **JSON tags**: All output structs use `snake_case` tags (`firmware_version`, `stream_uri`, etc.).

## Gotchas
- `generateUUID()` (discovery.go:79) uses Unix nanoseconds, not RFC4122. Works for WS-Discovery but fragile with strict implementations.
- UDP timeout is hardcoded via `discoverONVIFDevices(2 * time.Second)` in main.go:23.
- Discovery only probes `NetworkVideoTransmitter` type (discovery.go:23).
- No tests exist. Adding tests requires mocking ONVIF client or testing XML unmarshaling in isolation.

## Dependencies
- `github.com/use-go/onvif` v0.0.9 - ONVIF protocol implementation
- Go 1.24+ required (go.mod)

## Environment
Requires local network access:
- UDP multicast: `239.255.255.250:3702` for discovery
- Target device HTTP port for SOAP calls
