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
