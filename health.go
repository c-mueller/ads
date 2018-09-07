package ads

func (e DNSAdBlock) Health() bool {
	// More advanced plugins will check their state, i.e. are they
	// synchronized correctly against their backend etc.

	// Be careful though by making this a single point of failure. I.e. if 5 CoreDNS
	// instances are talking to the same backend and the backend goes down, *all* your
	// instances are unhealthy.

	// This one just returns OK.
	return true
}
