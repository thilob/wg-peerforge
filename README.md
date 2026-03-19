# wg-peerforge

Greenfield foundation for a terminal-based WireGuard server and peer manager.

## Scope

The first iteration focuses on:

- creating WireGuard servers from a structured data model
- generating and managing peers
- rendering `wg-quick` compatible configuration
- keeping Linux backend-specific behavior isolated from domain logic

## Layout

- `cmd/wireguard-tui`: CLI entry point
- `internal/domain`: business model and validation
- `internal/app`: use-cases and ports
- `internal/infrastructure`: rendering and system-specific adapters
- `internal/ui`: future terminal UI

## Build

```bash
go build ./cmd/wg-peerforge
```
