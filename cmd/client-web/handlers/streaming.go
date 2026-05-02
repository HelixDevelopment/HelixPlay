package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/HelixDevelopment/HelixPlay/pkg/core/input"
	"github.com/HelixDevelopment/HelixPlay/pkg/core/streaming"
)

// Streaming holds streaming-related HTTP handlers.
type Streaming struct {
	streamCtrl *streaming.SessionController
}

// RegisterRoutes registers streaming endpoints on the supplied mux.
func (s *Streaming) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/streaming/negotiate", s.Negotiate)
	mux.HandleFunc("/api/streaming/telemetry", s.Telemetry)
	mux.HandleFunc("/api/streaming/state", s.State)
	mux.HandleFunc("/api/streaming/controller", s.ControllerUpdate)
}

// Negotiate handles session negotiation requests.
func (s *Streaming) Negotiate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		HostAddress string `json:"host_address"`
		ClientID    string `json:"client_id"`
		UserID      string `json:"user_id"`
		AuthToken   string `json:"auth_token"`
		GameID      string `json:"game_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg := streaming.Config{
		HostAddress:     req.HostAddress,
		ClientID:        req.ClientID,
		UserID:          req.UserID,
		AuthToken:       req.AuthToken,
		RequestedGameID: req.GameID,
	}

	s.streamCtrl = streaming.NewSessionController(cfg)
	ctx := r.Context()
	if err := s.streamCtrl.Connect(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "connecting",
		"state":  string(s.streamCtrl.State()),
	})
}

// Telemetry receives client telemetry.
func (s *Streaming) Telemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// State returns the current streaming state.
func (s *Streaming) State(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	state := "idle"
	if s.streamCtrl != nil {
		state = string(s.streamCtrl.State())
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"state": state})
}

// ControllerUpdate accepts controller state from the browser Gamepad API.
func (s *Streaming) ControllerUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var st input.ControllerState
	if err := json.NewDecoder(r.Body).Decode(&st); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if s.streamCtrl != nil {
		_ = s.streamCtrl.SendControllerState(st)
	}

	w.WriteHeader(http.StatusNoContent)
}
