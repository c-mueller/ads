// Copyright 2018 Christian Müller <cmueller.dev@gmail.com>
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
	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

var requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: plugin.Namespace,
	Subsystem: "ads",
	Name:      "request_count_total",
	Help:      "Counter of requests made.",
}, []string{"server"})

var blockedRequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: plugin.Namespace,
	Subsystem: "ads",
	Name:      "blocked_request_count_total",
	Help:      "Counter of requests blocked by this plugin.",
}, []string{"server"})

var blockedRequestCountBySource = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: plugin.Namespace,
	Subsystem: "ads",
	Name:      "blocked_request_count",
	Help:      "Counter of requests blocked by this plugin for every source ip.",
}, []string{"server", "source"})

var requestCountBySource = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: plugin.Namespace,
	Subsystem: "ads",
	Name:      "request_count",
	Help:      "Counter of requests piped through this plugin for every source ip.",
}, []string{"server", "source"})

var once sync.Once
