/*
 * Copyright 2018 - 2020 Christian Müller <dev@c-mueller.xyz>
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
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"net"
)

func (e *DNSAdBlock) IsWhitelisted(qname string) bool {
	return e.whitelist[qname] || e.ConfiguredRuleSet.IsWhitelisted(qname) || e.FileRuleSet.IsWhitelisted(qname)
}

func (e *DNSAdBlock) IsBlacklisted(qname string) bool {
	return e.blacklist[qname] || e.ConfiguredRuleSet.IsBlacklisted(qname) || e.FileRuleSet.IsBlacklisted(qname)
}

func (e *DNSAdBlock) ShouldBlock(qname string) bool {
	return !e.IsWhitelisted(qname) && e.IsBlacklisted(qname)
}

func (e *DNSAdBlock) onBlock(w dns.ResponseWriter, r *dns.Msg, state *request.Request, trimmedQname string) error {
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

	if e.config.EnableLogging {
		log.Infof("Blocked request %q from %q", trimmedQname, state.IP())
	}
	return w.WriteMsg(m)
}
