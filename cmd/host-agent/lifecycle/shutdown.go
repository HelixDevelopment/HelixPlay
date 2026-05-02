package lifecycle

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Shutdown gracefully deregisters the host from the coordinator.
func Shutdown(coordinatorURL, hostID, authToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 10 * time.Second}
	url := coordinatorURL + "/api/v1/hosts/" + hostID + "/deregister"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create deregister request: %w", err)
	}
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("deregister request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("deregister returned status %d", resp.StatusCode)
	}

	return nil
}

// ShutdownSequence orchestrates graceful shutdown of multiple subsystems.
// Components are stopped in reverse dependency order.
type ShutdownSequence struct {
	steps []func() error
}

// NewShutdownSequence creates an empty shutdown sequence.
func NewShutdownSequence() *ShutdownSequence {
	return &ShutdownSequence{steps: make([]func() error, 0)}
}

// Add appends a shutdown step. Steps are executed LIFO (last added, first executed).
func (s *ShutdownSequence) Add(name string, fn func() error) {
	s.steps = append(s.steps, fn)
}

// Execute runs all shutdown steps. Errors are collected but execution continues.
func (s *ShutdownSequence) Execute() []error {
	var errs []error
	for i := len(s.steps) - 1; i >= 0; i-- {
		if err := s.steps[i](); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
