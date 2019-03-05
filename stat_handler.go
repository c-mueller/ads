package ads

import "time"

type Request struct {
	RequestedHostname string    `json:"requested_hostname"`
	Recipient         string    `json:"recipient"`
	Blocked           bool      `json:"blocked"`
	Timestamp         time.Time `json:"timestamp"`
}

type Stats struct {
	TotalRequests       int            `json:"total_requests"`
	BlockedRequests     int            `json:"blocked_requests"`
	TopPermittedDomains map[string]int `json:"top_permitted_domains"`
	TopBlockedDomains   map[string]int `json:"top_blocked_domains"`
	TopClients          map[string]int `json:"top_clients"`
}

type StatHandler interface {
	Insert(request Request) error
	GetRequestsBetween(from, to time.Time) []Request
	GetStats() Stats
}

