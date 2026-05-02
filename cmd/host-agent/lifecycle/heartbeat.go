package lifecycle

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// HeartbeatOptions configures the heartbeat loop.
type HeartbeatOptions struct {
	CoordinatorURL string
	HostID         string
	AuthToken      string
	Interval       time.Duration
	HTTPClient     *http.Client
}

// Heartbeater runs a background loop that periodically reports host status.
type Heartbeater struct {
	opts   HeartbeatOptions
	cancel context.CancelFunc
	done   chan struct{}
}

// NewHeartbeater creates a heartbeater with the given options.
func NewHeartbeater(opts HeartbeatOptions) *Heartbeater {
	if opts.Interval <= 0 {
		opts.Interval = 15 * time.Second
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Heartbeater{opts: opts, done: make(chan struct{})}
}

// Start begins the heartbeat loop.
func (h *Heartbeater) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel
	go h.loop(ctx)
}

// Stop halts the heartbeat loop and waits for the current iteration to finish.
func (h *Heartbeater) Stop() {
	if h.cancel != nil {
		h.cancel()
	}
	<-h.done
}

func (h *Heartbeater) loop(ctx context.Context) {
	defer close(h.done)
	ticker := time.NewTicker(h.opts.Interval)
	defer ticker.Stop()

	// Send immediate heartbeat
	h.beat(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.beat(ctx)
		}
	}
}

func (h *Heartbeater) beat(ctx context.Context) {
	payload := map[string]any{
		"host_id":    h.opts.HostID,
		"status":     "online",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	url := h.opts.CoordinatorURL + "/api/v1/hosts/" + h.opts.HostID + "/heartbeat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if h.opts.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.opts.AuthToken)
	}

	resp, err := h.opts.HTTPClient.Do(req)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}

// IsRunning reports whether the heartbeater loop is active.
func (h *Heartbeater) IsRunning() bool {
	select {
	case <-h.done:
		return false
	default:
		return h.cancel != nil
	}
}
