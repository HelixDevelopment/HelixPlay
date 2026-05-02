package capability

type Capabilities struct {
    Codecs           []string
    Resolutions       []string
    MaxFPS           int
    Haptics         bool
    AdaptiveTriggers bool
}

type NegotiationResult struct {
    Success         bool
    AgreedCodecs   []string
    AgreedResolution string
    AgreedFPS       int
    Reason          string
}

type Negotiator struct{}

func NewNegotiator() *Negotiator {
    return &Negotiator{}
}

func (n *Negotiator) Negotiate(client, server Capabilities) NegotiationResult {
    // Find common codecs
    agreedCodecs := []string{}
    serverCodecSet := make(map[string]bool)
    for _, c := range server.Codecs {
        serverCodecSet[c] = true
    }
    for _, c := range client.Codecs {
        if serverCodecSet[c] {
            agreedCodecs = append(agreedCodecs, c)
        }
    }
    
    if len(agreedCodecs) == 0 {
        return NegotiationResult{
            Success: false,
            Reason:  "No common codec found",
        }
    }
    
    // Pick best resolution (intersection of client and server)
    resolution := "1080p" // Default
    clientResSet := make(map[string]bool)
    for _, cr := range client.Resolutions {
        clientResSet[cr] = true
    }
outer:
    for _, r := range []string{"4K", "1440p", "1080p", "720p"} {
        if clientResSet[r] {
            for _, sr := range server.Resolutions {
                if r == sr {
                    resolution = r
                    break outer
                }
            }
        }
    }
    
    // Pick FPS
    fps := client.MaxFPS
    if fps > server.MaxFPS {
        fps = server.MaxFPS
    }
    
    return NegotiationResult{
        Success:         true,
        AgreedCodecs:   agreedCodecs,
        AgreedResolution: resolution,
        AgreedFPS:       fps,
    }
}
