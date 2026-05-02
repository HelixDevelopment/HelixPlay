package capability

import (
    "sync"
)

type Endpoint struct {
    Type     string
    Address  string
    Port     int
    Metadata map[string]string
}

type Discovery struct {
    mu        sync.RWMutex
    serviceType string
    address    string
}

func NewDiscovery(serviceType, address string) *Discovery {
    return &Discovery{
        serviceType: serviceType,
        address:    address,
    }
}

func (d *Discovery) Discover() ([]Endpoint, error) {
    d.mu.RLock()
    defer d.mu.RUnlock()
    return []Endpoint{
        {Type: d.serviceType, Address: d.address, Port: 50051, Metadata: map[string]string{"version": "1.0"}},
    }, nil
}

func (d *Discovery) DiscoverByType(serviceType string) []Endpoint {
    d.mu.RLock()
    defer d.mu.RUnlock()
    if serviceType == d.serviceType {
        return []Endpoint{
            {Type: d.serviceType, Address: d.address, Port: 50051},
        }
    }
    return nil
}
