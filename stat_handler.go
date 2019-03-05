// Copyright 2018 - 2019 Christian Müller <dev@c-mueller.xyz>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ads

import (
	"time"
)

type Request struct {
	RequestID         string `json:"request_id"`
	RequestedHostname string `json:"requested_hostname"`
	Recipient         string `json:"recipient"`
	Blocked           bool   `json:"blocked"`
	Timestamp         int64  `json:"timestamp"`
}

type Stats struct {
	TotalRequests       int            `json:"total_requests"`
	BlockedRequests     int            `json:"blocked_requests"`
	TopPermittedDomains map[string]int `json:"top_permitted_domains"`
	TopBlockedDomains   map[string]int `json:"top_blocked_domains"`
	TopClients          map[string]int `json:"top_clients"`
}

type StatHandler struct {
	Enabled  bool
	Endpoint string
	Repo     StatRepository
}

type StatRepository interface {
	Insert(request Request) (string, error)
	Delete(uid string) error
	GetRequestsBetween(from, to time.Time) []Request
	GetStats() Stats
	Cleanup() error
	Init() error
	Close() error
}

func (s *StatHandler) Init() error {
	if !s.Enabled {
		return nil
	}

	log.Info("Initializing Stat handler")

	if err := s.Repo.Init(); err != nil {
		return err
	}

	return nil
}
