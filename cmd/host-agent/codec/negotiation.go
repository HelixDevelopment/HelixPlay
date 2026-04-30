package codec

type NegotiationResult struct {
	Agreed []string
	Failed bool
	Reason string
}

type Negotiator struct{}

func NewNegotiator() *Negotiator {
	return &Negotiator{}
}

func (n *Negotiator) Negotiate(clientCodecs, serverCodecs []string) NegotiationResult {
	agreed := []string{}

	// Build server codec set
	serverSet := make(map[string]bool)
	for _, c := range serverCodecs {
		serverSet[c] = true
	}

	// Find intersection
	for _, c := range clientCodecs {
		if serverSet[c] {
			agreed = append(agreed, c)
		}
	}

	if len(agreed) == 0 {
		return NegotiationResult{
			Agreed: agreed,
			Failed: true,
			Reason: "No common codec found",
		}
	}

	return NegotiationResult{
		Agreed: agreed,
		Failed: false,
	}
}
