package discovery

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/hashicorp/mdns"
)

func (c *Client) runMDNS(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	c.mdnsLookup()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mdnsLookup()
		}
	}
}

func (c *Client) mdnsLookup() {
	entriesCh := make(chan *mdns.ServiceEntry, 8)
	processed := make(chan struct{})
	go func() {
		defer close(processed)
		for entry := range entriesCh {
			// Defensive copy: hashicorp/mdns may reuse the entry struct,
			// causing data races if we read asynchronously.
			entryCopy := *entry
			c.handleMDNSEntry(&entryCopy)
		}
	}()
	_ = mdns.Lookup("_helixplay._tcp", entriesCh)
	close(entriesCh)
	<-processed
}

func (c *Client) handleMDNSEntry(entry *mdns.ServiceEntry) {
	key := entry.AddrV4.String() + ":" + strconv.Itoa(entry.Port)
	localEndpoint := net.JoinHostPort(entry.AddrV4.String(), strconv.Itoa(entry.Port))

	h := Host{
		ID:            key,
		Name:          entry.Host,
		LocalEndpoint: localEndpoint,
		Status:        "online",
		LastSeenAt:    time.Now().UTC().Format(time.RFC3339),
	}

	for _, txt := range entry.InfoFields {
		switch {
		case len(txt) > 7 && txt[:7] == "hostid=":
			h.ID = txt[7:]
		case len(txt) > 5 && txt[:5] == "name=":
			h.Name = txt[5:]
		case len(txt) > 3 && txt[:3] == "os=":
			h.OS = txt[3:]
		case len(txt) > 8 && txt[:8] == "version=":
			h.Version = txt[8:]
		case len(txt) > 7 && txt[:7] == "status=":
			h.Status = txt[7:]
		}
	}

	c.mu.Lock()
	c.mdnsHosts[key] = h
	c.mu.Unlock()

	c.mergeHost(key)
}

func (c *Client) mergeHost(key string) {
	c.mu.Lock()
	mdnsH, hasMDNS := c.mdnsHosts[key]
	rendezvousH, hasRendezvous := c.rendezvousHosts[key]
	c.mu.Unlock()

	if !hasMDNS && !hasRendezvous {
		c.removeHost(key)
		return
	}

	merged := Host{}
	if hasRendezvous {
		merged = rendezvousH
	}
	if hasMDNS {
		if merged.ID == "" {
			merged.ID = mdnsH.ID
		}
		if merged.LocalEndpoint == "" {
			merged.LocalEndpoint = mdnsH.LocalEndpoint
		}
		if merged.Name == "" {
			merged.Name = mdnsH.Name
		}
		if merged.OS == "" {
			merged.OS = mdnsH.OS
		}
		if merged.Version == "" {
			merged.Version = mdnsH.Version
		}
		if merged.Status == "" {
			merged.Status = mdnsH.Status
		}
	}

	c.setHost(key, merged)
}
