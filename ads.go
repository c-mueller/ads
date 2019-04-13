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
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
	"net"
	"strings"
)

var log = clog.NewWithPlugin("ads")

type DNSAdBlock struct {
	Next       plugin.Handler
	BlockLists []string
	RuleSet    RuleSet
	blockMap   BlockMap
	updater    *BlocklistUpdater
	config     *adsPluginConfig
}

func (e *DNSAdBlock) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	qname := state.Name()

	qname = strings.TrimSuffix(qname, ".")

	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
	requestCountBySource.WithLabelValues(metrics.WithServer(ctx), state.IP()).Inc()

	if !e.RuleSet.IsWhitelisted(qname) && (e.blockMap[qname] || e.RuleSet.IsBlacklisted(qname)) {
		var answers []dns.RR
		if state.QType() == dns.TypeAAAA {
			answers = aaaa(state.Name(), []net.IP{e.config.TargetIPv6})
		} else {
			answers = a(state.Name(), []net.IP{e.config.TargetIP})
		}

		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative, m.RecursionAvailable = true, true
		m.Answer = answers

		w.WriteMsg(m)

		blockedRequestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
		blockedRequestCountBySource.WithLabelValues(metrics.WithServer(ctx), state.IP()).Inc()

		if e.config.EnableLogging {
			log.Infof("Blocked request %q from %q", qname, state.IP())
		}

		return dns.RcodeSuccess, nil
	} else {
		return plugin.NextOrFailure(e.Name(), e.Next, ctx, w, r)
	}
}

// Name implements the Handler interface.
func (e *DNSAdBlock) Name() string { return "ads" }

func a(zone string, ips []net.IP) []dns.RR {
	var answers []dns.RR
	for _, ip := range ips {
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeA,
			Class: dns.ClassINET, Ttl: 3600}
		r.A = ip
		answers = append(answers, r)
	}
	return answers
}

func aaaa(zone string, ips []net.IP) []dns.RR {
	var answers []dns.RR
	for _, ip := range ips {
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeAAAA,
			Class: dns.ClassINET, Ttl: 3600}
		r.AAAA = ip
		answers = append(answers, r)
	}
	return answers
}
