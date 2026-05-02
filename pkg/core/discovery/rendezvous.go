package discovery

import (
	"context"
	"time"

	discoveryv1 "github.com/HelixDevelopment/HelixPlay/pkg/protocol/v1/discovery"
)

func (c *Client) runRendezvous(ctx context.Context) {
	c.listRendezvous(ctx)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.listRendezvous(ctx)
		}
	}
}

func (c *Client) listRendezvous(ctx context.Context) {
	resp, err := c.rendezvousClient.ListHosts(ctx, &discoveryv1.ListHostsRequest{
		TenantId: c.tenantID,
		Region:   c.region,
		PageSize: 100,
	})
	if err != nil {
		return
	}

	seen := make(map[string]struct{})
	for _, rh := range resp.Hosts {
		h := rendezvousHostToHost(rh)
		key := h.ID
		if key == "" {
			key = h.PublicEndpoint
		}
		seen[key] = struct{}{}

		c.mu.Lock()
		c.rendezvousHosts[key] = h
		c.mu.Unlock()
		c.mergeHost(key)
	}

	c.mu.Lock()
	for key := range c.rendezvousHosts {
		if _, ok := seen[key]; !ok {
			delete(c.rendezvousHosts, key)
			c.mu.Unlock()
			c.mergeHost(key)
			c.mu.Lock()
		}
	}
	c.mu.Unlock()
}

func rendezvousHostToHost(rh *discoveryv1.RendezvousHost) Host {
	h := Host{
		ID:             rh.HostId,
		Name:           rh.HostId,
		PublicEndpoint: rh.PublicEndpoint,
		LocalEndpoint:  rh.LocalEndpoint,
		OS:             rh.Os,
		Version:        rh.Version,
		Status:         rh.Status,
		ActiveSessions: rh.ActiveSessionCount,
		LastSeenAt:     rh.LastSeenAt,
	}
	if rh.Capabilities != nil {
		h.Capabilities = CapabilitiesSnapshot{
			Codecs:     rh.Capabilities.Codecs,
			Transports: rh.Capabilities.Transports,
			MaxResolution: Resolution{
				Width:  rh.Capabilities.MaxResolution.GetWidth(),
				Height: rh.Capabilities.MaxResolution.GetHeight(),
			},
			MaxRefreshHz: rh.Capabilities.MaxRefreshRate,
			HDRFormats:   rh.Capabilities.HdrFormats,
		}
		for _, g := range rh.Capabilities.Gpus {
			h.Capabilities.GPUs = append(h.Capabilities.GPUs, GPUCapability{
				Index:    g.Index,
				Vendor:   g.Vendor,
				Model:    g.Model,
				Encoder:  g.Encoder,
				Codecs:   g.Codecs,
				VRAMGB:   g.VramGb,
				ThermalC: g.ThermalC,
			})
		}
	}
	return h
}
