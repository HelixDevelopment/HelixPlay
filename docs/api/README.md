# HelixPlay API Reference

## Overview

HelixPlay exposes both gRPC and REST APIs for client-server communication.

## gRPC Services

### StreamingControl (`pkg/protocol/v1/streaming.proto`)

Manages streaming session lifecycle.

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| NegotiateStream | stream NegotiateMessage | stream NegotiateMessage | Capability negotiation |
| StartStream | StartStreamRequest | StartStreamResponse | Begin streaming |
| StopStream | StopStreamRequest | StopStreamResponse | End streaming |

### RendezvousService (`pkg/protocol/v1/discovery.proto`)

Host discovery and registration.

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| RegisterHost | RegisterHostRequest | RegisterHostResponse | Host registration |
| ListHosts | ListHostsRequest | ListHostsResponse | Discover available hosts |
| Heartbeat | HeartbeatRequest | HeartbeatResponse | Host health update |

### CatalogService (`pkg/protocol/v1/catalog.proto`)

Game catalog and metadata.

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| ListGames | ListGamesRequest | ListGamesResponse | Paginated game list |
| GetGame | GetGameRequest | Game | Single game lookup |
| SearchGames | SearchGamesRequest | SearchGamesResponse | Full-text search |

## REST Endpoints

### Core Backend

- `GET /health` — Health check
- `GET /api/v1/tenants` — List tenants (admin)
- `POST /api/v1/hosts/register` — Host registration
- `POST /api/v1/hosts/{id}/heartbeat` — Host heartbeat
- `POST /api/v1/hosts/{id}/deregister` — Host deregistration

### Discovery Service

- `GET /health` — Health check
- `POST /register` — Register host beacon
- `POST /deregister` — Deregister host beacon
- `GET /hosts` — List discovered hosts

## Authentication

All endpoints require JWT Bearer tokens (RS256). Device clients use RFC 8628 device authorization grant.

## Rate Limits

- REST: 1000 requests/minute per client
- gRPC streaming: 1 negotiation per 5 seconds per host
