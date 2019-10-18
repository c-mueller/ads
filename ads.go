/*
 * Copyright 2018 - 2019 Christian Müller <dev@c-mueller.xyz>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ads

import (
	"fmt"
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
	Next              plugin.Handler
	ConfiguredRuleSet ConfiguredRuleSet
	FileRuleSet       UpdateableRuleset
	blacklist         ListMap
	whitelist         ListMap
	updater           *ListUpdater
	config            *adsPluginConfig
}

func (e *DNSAdBlock) IsWhitelisted(qname string) bool {
	return e.whitelist[qname] || e.ConfiguredRuleSet.IsWhitelisted(qname) || e.FileRuleSet.IsWhitelisted(qname)
}
func (e *DNSAdBlock) IsBlacklisted(qname string) bool {
	return e.blacklist[qname] || e.ConfiguredRuleSet.IsBlacklisted(qname) || e.FileRuleSet.IsBlacklisted(qname)
}
func (e *DNSAdBlock) ShouldBlock(qname string) bool {
	return !e.IsWhitelisted(qname) && e.IsBlacklisted(qname)
}

func (e *DNSAdBlock) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	trimmedQname := state.Name()

	trimmedQname = strings.TrimSuffix(trimmedQname, ".")

	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
	requestCountBySource.WithLabelValues(metrics.WithServer(ctx), state.IP()).Inc()

	if e.ShouldBlock(trimmedQname) {
		blockedRequestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
		blockedRequestCountBySource.WithLabelValues(metrics.WithServer(ctx), state.IP()).Inc()
		var answers []dns.RR
		if e.config.WriteNXDomain {
			answers = nxdomain(state.Name())
		} else if state.QType() == dns.TypeAAAA {
			answers = aaaa(state.Name(), []net.IP{e.config.TargetIPv6})
		} else {
			answers = a(state.Name(), []net.IP{e.config.TargetIP})
		}

		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative, m.RecursionAvailable = true, true
		m.Answer = answers

		w.WriteMsg(m)

		if e.config.EnableLogging {
			log.Infof("Blocked request %q from %q", trimmedQname, state.IP())
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

func nxdomain(zone string) []dns.RR {
	s := fmt.Sprintf("%s 60 IN SOA ns1.%s postmaster.%s 1524370381 14400 3600 604800 60", zone, zone, zone)
	soa, _ := dns.NewRR(s)
	return []dns.RR{soa}
}
