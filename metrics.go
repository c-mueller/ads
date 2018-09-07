package ads

import (
	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

var requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: plugin.Namespace,
	Subsystem: "example",
	Name:      "request_count_total",
	Help:      "Counter of requests made.",
}, []string{"server"})

var once sync.Once
