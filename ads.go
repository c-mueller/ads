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
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
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

func (e *DNSAdBlock) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := &request.Request{W: w, Req: r}

	trimmedQname := state.Name()
	trimmedQname = strings.TrimSuffix(trimmedQname, ".")

	requestCountTotal.WithLabelValues(metrics.WithServer(ctx)).Inc()

	if e.ShouldBlock(trimmedQname) {
		blockedRequestCountTotal.WithLabelValues(metrics.WithServer(ctx)).Inc()
		blockedRequestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
		e.onBlock(w, r, state, trimmedQname)
		return dns.RcodeSuccess, nil
	} else {
		brw := &BlockingResponseWriter{
			Writer:       w,
			Plugin:       e,
			Request:      r,
			RequestState: state,
		}
		return plugin.NextOrFailure(e.Name(), e.Next, ctx, brw, r)
	}
}

// Name implements the Handler interface.
func (e *DNSAdBlock) Name() string { return "ads" }
