# wg-peerforge

Greenfield foundation for a terminal-based WireGuard server and peer manager.

## Scope

The first iteration focuses on:

- creating WireGuard servers from a structured data model
- generating and managing peers
- rendering `wg-quick` compatible configuration
- keeping Linux backend-specific behavior isolated from domain logic

## Layout

- `cmd/wg-peerforge`: CLI entry point
- `internal/domain`: business model and validation
- `internal/app`: use-cases and ports
- `internal/infrastructure`: rendering and system-specific adapters
- `internal/ui`: future terminal UI

## Persistence

Structured server and peer data is stored in JSON files below `data/servers/<server-id>/`.
This keeps the first implementation dependency-free while preserving a clear adapter boundary
behind the `ConfigStore` port.

## Build

```bash
go build ./cmd/wg-peerforge
```

## CLI

```bash
go run ./cmd/wg-peerforge --help
go run ./cmd/wg-peerforge create-server -id alpha -name Alpha -endpoint vpn.example.com
go run ./cmd/wg-peerforge create-peer -server-id alpha -id phone -name "Phone"
go run ./cmd/wg-peerforge tui
go run ./cmd/wg-peerforge list-servers
go run ./cmd/wg-peerforge list-peers -server-id alpha
go run ./cmd/wg-peerforge render-server -server-id alpha
```

Made with love in NRW, powered by VibeCoding.
