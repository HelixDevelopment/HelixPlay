## 7. API Design & Communication Layer

The communication layer must satisfy four competing requirements: low-latency streaming for controller input and video, reliable request-response semantics for catalog and session management, persistent bidirectional channels for session orchestration, and efficient inter-service messaging for horizontal scaling. This chapter defines the API architecture that addresses each requirement with a specific protocol choice—REST for catalog and session management, WebSockets for real-time session control, Server-Sent Events (SSE) for push notifications, WebRTC for streaming negotiation, and gRPC (via Connect RPC) for internal services—all implemented in Go with Protocol Buffers (protobuf) as the canonical schema language.

### 7.1 REST API Design

#### 7.1.1 Base Conventions

The public REST API follows URL path versioning with the prefix `/api/v1/`, a pattern identified as the most practical for most teams because it is visible, cacheable, and easy to test [^319^]. All bodies use JSON with snake_case field names. Authentication follows OAuth 2.0 Bearer token flow: clients obtain a short-lived access token (15-minute expiration) and a rotating refresh token via `POST /api/v1/auth/token`, then present the access token in the `Authorization: Bearer <token>` header. Refresh tokens are stored in Redis with a 7-day TTL to enable session revocation.

Rate limiting applies at two layers. The API gateway uses `go.uber.org/ratelimit`, implementing a leaky-bucket algorithm via atomic operations without background goroutines [^346^]. Default limits are 100 requests per minute for authenticated users and 20 per minute for unauthenticated endpoints. Application-layer token buckets protect downstream resources.

Framework selection prioritizes ecosystem maturity. Benchmarks show Fiber at 89,247 req/s, Gin at 76,832 req/s, and Echo at 72,156 req/s under synthetic `wrk` load (12 threads, 400 connections) [^290^]. However, with PostgreSQL queries included, all three converge to approximately 3,000–3,250 req/s, as the database becomes the bottleneck [^290^]. Because the platform's REST endpoints are I/O-bound, Gin is selected for its middleware ecosystem and straightforward JWT integration via `appleboy/gin-jwt` [^301^].

#### 7.1.2 Endpoint Definitions

The REST surface exposes five resource collections: games, sessions, hosts, authentication, and user profiles. Catalog endpoints support cursor-based pagination (`page`, `page_size` capped at 100). Session endpoints enforce strict state transitions server-side. Host endpoints support the pairing flow for new hosts joining the fleet.

Protobuf schemas define all request and response types. Packages follow the `cloudstream.{domain}.v1` namespace: `cloudstream.catalog.v1` for games, `cloudstream.session.v1` for sessions, and `cloudstream.host.v1` for hosts. Code generation produces Go structs via `protoc-gen-go`, TypeScript interfaces via `protoc-gen-es`, and Dart classes via `protoc-gen-dart`. Field numbers are reserved on deletion to maintain backward compatibility [^318^].

#### 7.1.3 REST Endpoint Reference

Table 7.1 lists every public REST endpoint with its HTTP method, path, protobuf request and response schema, authentication requirement, and rate limit tier.

| Method | Path | Request Schema | Response Schema | Auth | Rate Limit |
|---|---|---|---|---|---|
| GET | `/api/v1/games` | `catalog.v1.GameListRequest` | `catalog.v1.GameListResponse` | Optional | 100/min |
| GET | `/api/v1/games/{id}` | Path: `game_id` | `catalog.v1.GameDetailsResponse` | Optional | 100/min |
| POST | `/api/v1/sessions` | `session.v1.CreateSessionRequest` | `session.v1.Session` | Required | 30/min |
| GET | `/api/v1/sessions/{id}` | Path: `session_id` | `session.v1.SessionStatusResponse` | Required | 60/min |
| DELETE | `/api/v1/sessions/{id}` | Path: `session_id` | `session.v1.Session` | Required | 30/min |
| GET | `/api/v1/hosts` | `host.v1.HostListRequest` | `host.v1.HostListResponse` | Required | 60/min |
| POST | `/api/v1/hosts/{id}/pair` | Path: `host_id` + `host.v1.PairHostRequest` | `host.v1.Host` | Admin | 10/min |
| POST | `/api/v1/auth/token` | `auth.v1.TokenRequest` | `auth.v1.TokenResponse` | None | 20/min |
| POST | `/api/v1/auth/refresh` | `auth.v1.RefreshRequest` | `auth.v1.TokenResponse` | None | 20/min |

The `GameListResponse` uses `next_page_token` for cursor-based pagination, avoiding `OFFSET` query degradation in PostgreSQL. The `CreateSessionRequest` accepts an optional `host_id`; if omitted, the scheduler selects the closest available host by geographic proximity and load. The `PairHostRequest` endpoint requires the host to sign a nonce with its pre-registered public key to prevent unauthorized enrollment.

### 7.2 WebSocket API for Session Control

#### 7.2.1 Protocol and Authentication

WebSocket connections provide the persistent bidirectional channel required for session control operations that demand lower latency than REST round-trips permit. Clients connect to `/ws/v1/control` and authenticate during the handshake via the `Authorization` header; invalid tokens result in an immediate `1008` close. All messages use JSON with a uniform envelope: `{ "type": "<message_type>", "payload": {}, "timestamp": <unix_ms>, "correlation_id": "<uuid>" }`.

The handler uses the `coder/websocket` library, which achieves approximately 45,000 messages per second at 1,000 concurrent clients with ~19 ms round-trip time [^293^]. It was selected over Gorilla WebSocket due to active maintenance and a cleaner API.

#### 7.2.2 Message Types

Table 7.2 defines the message types exchanged over the WebSocket control channel. Each type follows a request-response or fire-and-forget pattern, with error responses using the same `correlation_id` as the originating message to enable client-side request matching.

| Message Type | Direction | Pattern | Payload Fields | Description |
|---|---|---|---|---|
| `session.create` | Client → Server | Request/Response | `game_id`, `host_id` (opt), `prefs` | Initiates a new gaming session; server responds with `session_id` and `host_endpoint` |
| `session.join` | Client → Server | Request/Response | `session_id`, `user_id` | Joins an existing multiplayer or shared session |
| `session.leave` | Client → Server | Fire-and-forget | `session_id` | Gracefully exits a session; triggers host cleanup |
| `stream.start` | Client → Server | Request/Response | `session_id`, `video_prefs` | Begins WebRTC streaming negotiation; returns SDP offer parameters |
| `stream.pause` | Client → Server | Fire-and-forget | `session_id`, `reason` | Pauses the video stream (backgrounding, network degradation) |
| `stream.resume` | Client → Server | Request/Response | `session_id` | Resumes a paused stream; may renegotiate codec parameters |
| `input.configure` | Client → Server | Request/Response | `controller_type`, `button_map`, `sensitivity` | Submits controller profile for the session; host applies mapping |
| `host.command` | Server → Client | Server Push | `command`, `args` | Host-to-client directives: resolution change, bitrate adaptation, error |
| `session.event` | Server → Client | Server Push | `event_type`, `data` | Lifecycle events: `host_ready`, `game_launched`, `error` |

The `stream.start` message type initiates the WebRTC signaling flow described in Section 7.4. Upon receiving this message, the server creates a Pion `PeerConnection`, generates an SDP offer, and returns it to the client via the response envelope. The client then generates an SDP answer and sends it back via a subsequent `stream.answer` message (not listed, as it is handled within the WebRTC signaling sub-protocol rather than the general control channel).

#### 7.2.3 Scaling: Cross-Server Distribution

Horizontal scaling requires sticky session routing and cross-server message broadcasting. Standard round-robin load balancing fails for WebSockets because the protocol requires connection stickiness—a client may handshake on Server A while subsequent requests route to Server B, which has no record of the connection [^302^]. The platform uses cookie-based session affinity via `nginx.ingress.kubernetes.io/affinity: "cookie"` with `affinity-mode: "persistent"` [^302^].

For cross-server message delivery, Redis Pub/Sub acts as a broadcast layer. Each WebSocket server subscribes to `ws:broadcast:<session_id>`; when a server receives a message requiring fan-out, it publishes to Redis, and all subscribed servers forward to their local connections [^340^]. This pattern adds negligible latency (<1 ms) while enabling elastic scaling without a global connection registry.

#### 7.2.4 Go WebSocket Handler Implementation

The following code example illustrates the Go WebSocket handler using the goroutine-per-connection pattern. Each accepted connection spawns two goroutines: one for reading client messages and dispatching to handlers, and one for writing messages from a buffered outbound channel. A `sync.WaitGroup` coordinates graceful shutdown when the server receives a termination signal.

```go
package ws

import (
    "context"
    "encoding/json"
    "net/http"
    "sync"
    "time"

    "github.com/coder/websocket"
)

// MessageEnvelope defines the uniform JSON envelope for all WS messages.
type MessageEnvelope struct {
    Type          string          `json:"type"`
    Payload       json.RawMessage `json:"payload"`
    Timestamp     int64           `json:"timestamp"`
    CorrelationID string          `json:"correlation_id"`
}

// ControlHandler manages WebSocket connections for session control.
type ControlHandler struct {
    upgrader    websocket.AcceptOptions
    sessions    map[string]*ClientConn
    sessionsMu  sync.RWMutex
    broadcaster BroadcastAdapter // Redis Pub/Sub adapter
    wg          sync.WaitGroup
}

// ClientConn represents a single WebSocket client connection.
type ClientConn struct {
    conn      *websocket.Conn
    userID    string
    sessionID string
    send      chan MessageEnvelope
    ctx       context.Context
    cancel    context.CancelFunc
}

// ServeHTTP upgrades HTTP connections to WebSocket and starts goroutines.
func (h *ControlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Extract and validate JWT from Authorization header
    token := extractBearer(r.Header.Get("Authorization"))
    claims, err := validateJWT(token)
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
        OriginPatterns: []string{"*.cloudstream.example.com"},
    })
    if err != nil {
        return
    }
    defer conn.Close(websocket.StatusNormalClosure, "")

    ctx, cancel := context.WithCancel(r.Context())
    client := &ClientConn{
        conn:   conn,
        userID: claims.UserID,
        send:   make(chan MessageEnvelope, 64),
        ctx:    ctx,
        cancel: cancel,
    }

    h.register(client)

    h.wg.Add(2)
    go h.readPump(client)
    go h.writePump(client)

    <-ctx.Done()
    h.unregister(client)
}

// readPump reads messages from the WebSocket and dispatches to handlers.
func (h *ControlHandler) readPump(c *ClientConn) {
    defer h.wg.Done()
    defer c.cancel()

    for {
        _, data, err := c.conn.Read(c.ctx)
        if err != nil {
            return // Connection closed or error
        }

        var env MessageEnvelope
        if err := json.Unmarshal(data, &env); err != nil {
            continue // Malformed message; skip
        }

        handler := h.route(env.Type)
        if handler != nil {
            resp := handler(c, &env)
            if resp != nil {
                c.send <- *resp
            }
        }
    }
}

// writePump writes messages from the send channel to the WebSocket.
func (h *ControlHandler) writePump(c *ClientConn) {
    defer h.wg.Done()

    ticker := time.NewTicker(30 * time.Second) // keepalive ping
    defer ticker.Stop()

    for {
        select {
        case msg, ok := <-c.send:
            if !ok {
                c.conn.Close(websocket.StatusNormalClosure, "")
                return
            }
            payload, _ := json.Marshal(msg)
            c.conn.Write(c.ctx, websocket.MessageText, payload)

        case <-ticker.C:
            c.conn.Ping(c.ctx)

        case <-c.ctx.Done():
            return
        }
    }
}

// Shutdown waits for all connection goroutines to finish.
func (h *ControlHandler) Shutdown(ctx context.Context) error {
    done := make(chan struct{})
    go func() {
        h.wg.Wait()
        close(done)
    }()
    select {
    case <-done:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

The handler implements a 64-message outbound buffer per connection to absorb temporary backpressure. The 30-second keepalive ping cycle detects half-open connections—a critical requirement for detecting client disconnections in mobile scenarios where the underlying TCP connection may linger after the device loses network coverage. The `Shutdown` method uses a `sync.WaitGroup` to drain active goroutines within a configurable timeout (default 10 seconds) during rolling deployments.

### 7.3 Server-Sent Events (SSE) for Real-Time Updates

#### 7.3.1 Event Types and Delivery Model

Server-Sent Events provide a lightweight, unidirectional push channel for state changes that do not require client-initiated responses. SSE operates over standard HTTP, meaning it requires no special proxy configuration and inherits existing load balancer health-checking and retry semantics [^291^]. The cloud gaming platform uses SSE at the endpoint `/events/v1/stream` for six categories of real-time update.

| Event Name | Payload Fields | Trigger Condition | Client Action |
|---|---|---|---|
| `host.online` | `host_id`, `region`, `capacity` | Host heartbeat first detected | Add host to available pool in UI |
| `host.offline` | `host_id`, `reason`, `last_seen` | Host heartbeat missed for 3 intervals | Remove host from pool; migrate sessions |
| `game.installed` | `game_id`, `host_id`, `install_path` | Game installation completes on host | Enable "Play" button for that game/host |
| `game.updated` | `game_id`, `version`, `patch_notes` | Game update applied to host | Notify user of available update |
| `session.started` | `session_id`, `host_id`, `game_id`, `started_at` | Session transitions to ACTIVE state | Transition UI to streaming view |
| `session.ended` | `session_id`, `duration`, `reason` | Session transitions to TERMINATED | Return to catalog; show session summary |
| `achievement.unlocked` | `achievement_id`, `game_id`, `name` | Game reports achievement unlock | Display overlay notification |

Each SSE event follows the W3C EventSource format: `event: <name>\ndata: <json_payload>\nid: <event_id>\n\n`. The `id` field enables the `Last-Event-ID` header for automatic reconnection—when a client reconnects after a network interruption, it sends the last received event ID, and the server replays any missed events from a ring buffer maintained per user session.

#### 7.3.2 Go Implementation

The SSE implementation uses Go's standard `net/http` package with the `http.Flusher` interface, requiring zero external dependencies [^292^]. The handler sets three critical headers before writing the first event: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, and `Connection: keep-alive`. The `http.Flusher` interface forces immediate delivery of buffered data to the client rather than waiting for the response buffer to fill [^292^]. Client disconnect detection uses `r.Context().Done()`, which fires when the underlying TCP connection closes or the client times out.

The event distribution architecture follows a pub/sub model: internal services publish typed events to NATS JetStream topics (`events.host`, `events.game`, `events.session`, `events.achievement`), and the SSE handler subscribes to a user-specific NATS subject that receives only events relevant to that user's active sessions and hosts. This filtering at the NATS level prevents the SSE handler from processing events for users other than the connected client, reducing CPU and memory overhead.

#### 7.3.3 Client Consumption

Clients consume SSE through platform-specific EventSource implementations. The Angular web client uses the native `EventSource` API. The Flutter client uses the `sse_client` package with Bearer token header support. The iOS Swift client uses `URLSession` with a custom delegate processing SSE chunks as they arrive. All clients implement exponential backoff for reconnection (1–30 seconds, with jitter) to prevent thundering-herd scenarios after server restarts.

### 7.4 WebRTC Signaling Protocol

#### 7.4.1 Signaling Flow and SDP Exchange

WebRTC requires an out-of-band signaling channel to exchange Session Description Protocol (SDP) offers and answers, plus Interactive Connectivity Establishment (ICE) candidates, before a direct peer-to-peer connection is established. The platform uses the existing WebSocket control channel (Section 7.2) for signaling. The flow proceeds as follows: the client sends `stream.start`; the server creates a Pion `PeerConnection` (v4.2.11), generates an SDP offer, and returns it; the client sets the remote description, generates an SDP answer via `stream.answer`; both sides exchange ICE candidates until a path is established.

Pion WebRTC is a pure Go implementation with no CGO, supporting Windows, macOS, Linux, iOS, Android, and WebAssembly [^17^]. Version 4.2.11 adds ICE renomination, `RemoteIPFilter` for security hardening, and `AlwaysNegotiateDataChannels` [^17^]. Connection benchmarks indicate ICE establishment averages ~123 ms and DTLS handshake ~154 ms from the server [^376^], placing total signaling overhead at roughly 300–400 ms—within the sub-500 ms latency budget.

#### 7.4.2 TURN Server Integration

Approximately 15–20% of sessions involve clients or hosts behind symmetric NAT or enterprise firewalls that block direct UDP peer-to-peer connectivity. For these sessions, a TURN (Traversal Using Relays around NAT) server relays media traffic. The platform uses `pion/turn`, a pure Go TURN implementation. Authentication uses ephemeral HMAC credentials: the signaling server generates a time-limited username-password pair (session duration plus 5 minutes) using a shared secret, passed to the client in the `stream.start` response.

ICE candidate gathering follows a priority-ordered process. Both sides gather candidates of three types: `host` (local IP), `srflx` (server-reflexive via STUN), and `relay` (TURN). Pion performs connectivity checks in priority order, preferring direct paths, then server-reflexive, then relay, per RFC 5245.

#### 7.4.3 Pion PeerConnection Setup with DataChannel

The following code example demonstrates the server-side Pion PeerConnection setup with a configured DataChannel for game input. The DataChannel operates in unordered, lossy mode to minimize latency for controller state updates, which tolerate occasional packet loss better than delayed delivery.

```go
package webrtc

import (
    "encoding/json"
    "fmt"
    "time"

    "github.com/pion/webrtc/v4"
)

// InputDataChannelLabel identifies the game input DataChannel.
const InputDataChannelLabel = "game-input"

// StreamNegotiator manages PeerConnection lifecycle for a gaming session.
type StreamNegotiator struct {
    peerConnection *webrtc.PeerConnection
    inputDC        *webrtc.DataChannel
    sessionID      string
    onInputFrame   func([]byte)
}

// NewStreamNegotiator creates a PeerConnection with TURN/STUN servers.
func NewStreamNegotiator(
    sessionID string,
    iceServers []webrtc.ICEServer,
    onInput func([]byte),
) (*StreamNegotiator, error) {
    config := webrtc.Configuration{
        ICEServers: iceServers,
        ICETransportPolicy: webrtc.ICETransportPolicyAll,
        BundlePolicy: webrtc.BundlePolicyMaxBundle,
        RTCPFeedback: webrtc.RTCPFeedback{
            {Type: "goog-remb"},
            {Type: "ccm", Parameter: "fir"},
            {Type: "nack"},
            {Type: "nack", Parameter: "pli"},
        },
    }

    settingEngine := webrtc.SettingEngine{}
    // Enable single-port mode for simplified firewall rules
    settingEngine.SetNetworkTypes([]webrtc.NetworkType{
        webrtc.NetworkTypeUDP4,
        webrtc.NetworkTypeUDP6,
    })

    api := webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine))
    pc, err := api.NewPeerConnection(config)
    if err != nil {
        return nil, fmt.Errorf("peerconnection: %w", err)
    }

    sn := &StreamNegotiator{
        peerConnection: pc,
        sessionID:      sessionID,
        onInputFrame:   onInput,
    }

    // Create the game input DataChannel with low-latency settings
    dcConfig := &webrtc.DataChannelInit{
        Ordered:    boolPtr(false),  // Allow out-of-order delivery
        MaxRetransmits: uint16Ptr(0), // No retransmits for input
    }
    dc, err := pc.CreateDataChannel(InputDataChannelLabel, dcConfig)
    if err != nil {
        pc.Close()
        return nil, fmt.Errorf("datachannel: %w", err)
    }
    sn.inputDC = dc

    dc.OnOpen(func() {
        // DataChannel ready for input transmission
    })
    dc.OnMessage(func(msg webrtc.DataChannelMessage) {
        sn.onInputFrame(msg.Data)
    })

    // Handle ICE connection state changes
    pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
        if state == webrtc.ICEConnectionStateFailed {
            pc.Close()
        }
    })

    // Handle incoming tracks (video from host)
    pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
        // Forward video track to client via RTP relay
    })

    return sn, nil
}

// CreateOffer generates an SDP offer for the client.
func (sn *StreamNegotiator) CreateOffer() (string, error) {
    offer, err := sn.peerConnection.CreateOffer(nil)
    if err != nil {
        return "", err
    }
    if err := sn.peerConnection.SetLocalDescription(offer); err != nil {
        return "", err
    }
    // Wait for ICE gathering to complete or timeout
    gatherComplete := webrtc.GatheringCompletePromise(sn.peerConnection)
    select {
    case <-gatherComplete:
    case <-time.After(5 * time.Second):
        // Return with gathered candidates so far
    }
    return sn.peerConnection.LocalDescription().SDP, nil
}

// SetAnswer processes the client's SDP answer.
func (sn *StreamNegotiator) SetAnswer(sdp string) error {
    answer := webrtc.SessionDescription{
        Type: webrtc.SDPTypeAnswer,
        SDP:  sdp,
    }
    return sn.peerConnection.SetRemoteDescription(answer)
}

// AddICECandidate adds a trickled ICE candidate from the client.
func (sn *StreamNegotiator) AddICECandidate(candidate string, sdpMid string, sdpMLineIndex int) error {
    return sn.peerConnection.AddICECandidate(webrtc.ICECandidateInit{
        Candidate:     candidate,
        SDPMid:        &sdpMid,
        SDPMLineIndex: &sdpMLineIndex,
    })
}

func boolPtr(b bool) *bool       { return &b }
func uint16Ptr(u uint16) *uint16 { return &u }
```

The `StreamNegotiator` configures the DataChannel with `Ordered: false` and `MaxRetransmits: 0` because controller input frames are time-critical—delivering a stale input frame is less valuable than dropping it and receiving the next one. This matches the transport requirements identified in the system architecture: input batching versus immediate transmission is a critical tradeoff, and immediate fire-and-forget delivery via an unreliable DataChannel achieves the lowest possible latency [^385^]. The SCTP (Stream Control Transmission Protocol) layer underlying Pion's DataChannels, enhanced with the RACK (Recent Acknowledgment) loss recovery algorithm, achieves 316.42 Mbps goodput with a p50 latency of 11.86 ms—representing a 34.9% throughput improvement and 27.5% latency reduction over the baseline implementation [^385^].

### 7.5 gRPC for Internal Services

#### 7.5.1 Connect RPC for Multi-Protocol Services

Internal service communication uses gRPC over HTTP/2 for bidirectional streaming and type-safe code generation from `.proto` files [^399^]. However, gRPC's native binary protocol requires a proxy for browser clients (gRPC-Web). The platform uses Connect RPC (connectrpc.com), which supports three protocols from a single handler: native gRPC for internal services, gRPC-Web for browsers, and the Connect protocol over HTTP/1.1 [^341^]. Connect generates idiomatic Go code using only `net/http` [^341^]. Bidirectional streaming requires end-to-end HTTP/2 [^351^].

The service mesh defines five primary services: `SessionService`, `HostService`, `CatalogService`, `StreamService`, and `AnalyticsService`, each in a separate `cloudstream.internal.v1` protobuf package. Handlers implement both gRPC and Connect interfaces, allowing internal microservices to use native gRPC while external clients communicate via Connect without a proxy.

#### 7.5.2 Service Mesh and Security

All internal gRPC communication is secured with mutual TLS (mTLS), in which both the client and server present X.509 certificates validated against a shared Certificate Authority (CA). Service discovery uses HashiCorp Consul, which provides health-checking, DNS-based service resolution, and a key-value store for dynamic configuration. Load balancing uses client-side round-robin selection from the set of healthy service instances returned by Consul, with automatic failover when health checks fail.

Table 7.4 compares the communication protocols used across the platform's different communication patterns, summarizing their roles, transport characteristics, and security models.

| Protocol | Pattern | Transport | Serialization | Latency (typical) | Security | Use Case |
|---|---|---|---|---|---|---|
| REST (Gin) | Request/Response | HTTP/1.1, HTTP/2 | JSON | 50–150 ms | OAuth 2.0 JWT | Catalog queries, session CRUD |
| WebSocket | Bidirectional streaming | HTTP-upgraded TCP | JSON (control), binary (input) | <10 ms | JWT at handshake | Session control, signaling |
| SSE | Server push | HTTP/1.1 | Event stream (text) | <50 ms | OAuth 2.0 JWT | Host status, game events |
| WebRTC (Pion) | Peer-to-peer | UDP (SRTP/DTLS/SCTP) | RTP (video), binary (input) | <5 ms | DTLS + SRTP | Video streaming, input DataChannel |
| gRPC (Connect RPC) | Request/Response + Streaming | HTTP/2 | Protobuf | <10 ms | mTLS + service tokens | Internal microservices |
| NATS Core | Pub/Sub | TCP | Protobuf/JSON | <1 ms | TLS + JWT | Controller input, real-time events |
| NATS JetStream | Persistent streams | TCP | Protobuf/JSON | 1–5 ms | TLS + JWT | Session events, analytics pipeline |

The latency figures reflect documented values from the technology stack: NATS Core delivers sub-millisecond latency to the 99.7th percentile [^299^]; JetStream achieves 3.2 ms p99 at 820,000 messages per second on a 3-node cluster [^343^]; WebRTC DataChannels achieve p50 latency of 11.86 ms at the SCTP layer [^385^]. REST latencies are higher due to connection setup and JSON serialization, acceptable for non-time-critical operations.

The event bus uses NATS Core for latency-critical data (controller input at 60 Hz, signaling messages) and JetStream for durable streams requiring replay (session events, achievements). This balances the <1 ms latency of core NATS against JetStream's durability at 1–5 ms overhead [^343^]. NATS serves as a first-class infrastructure component alongside gRPC: NATS for fire-and-forget events, gRPC for request-response interactions requiring confirmation.
