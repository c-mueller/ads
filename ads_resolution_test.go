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
	"context"
	"fmt"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestLookup_Block_IPv6(t *testing.T) {
	blacklist := make([]string, 0)

	testCases := make([]test.Case, 0)
	for i := 0; i < 10; i++ {
		qname := fmt.Sprintf("testhost-%09d.local.test.tld", i+1)
		blacklist = append(blacklist, qname)

		tcase := test.Case{
			Qname: qname,
			Qtype: dns.TypeAAAA,
			Answer: []dns.RR{
				test.AAAA(fmt.Sprintf("%s. 3600	IN	AAAA fe80::9cbd:c3ff:fe28:e133", qname)),
			},
		}
		testCases = append(testCases, tcase)
	}

	testCases = append(testCases, initAllowedTestCases()[10:]...)

	p := initTestPlugin(t, BuildRuleset(make([]string, 0), blacklist))
	ctx := context.TODO()

	resolveTestCases(testCases, p, ctx, t)
}

func TestLookup_RegexBlacklist(t *testing.T) {
	ruleset := getEmptyRuleset()

	err := ruleset.AddRegexToBlacklist(`(^|\.)local\.c-mueller\.de$`)
	assert.NoError(t, err)

	testCases := make([]test.Case, 0)
	for i := 0; i < 100; i++ {
		qname := fmt.Sprintf("testhost-%09d.local.c-mueller.de", i+1)
		tcase := test.Case{
			Qname: qname,
			Qtype: dns.TypeA,
			Answer: []dns.RR{
				test.A(fmt.Sprintf("%s. 3600	IN	A 10.1.33.7", qname)),
			},
		}
		testCases = append(testCases, tcase)
	}

	for i := 0; i < 10; i++ {
		qname := fmt.Sprintf("testhost-%09d.c-mueller.de", i+1)

		tcase := test.Case{
			Qname: qname,
			Qtype: dns.TypeA,
			Rcode: dns.RcodeNameError,
		}
		testCases = append(testCases, tcase)
	}

	p := initTestPlugin(t, ruleset)
	ctx := context.TODO()

	resolveTestCases(testCases, p, ctx, t)
}

func TestLookup_RegexWhitelist(t *testing.T) {
	ruleset := getEmptyRuleset()

	err := ruleset.AddRegexToWhitelist(`(^|\.)local\.test\.tld$`)
	assert.NoError(t, err)

	testCases := make([]test.Case, 0)
	for i := 0; i < 100; i++ {
		tcase := test.Case{
			Qname: fmt.Sprintf("testhost-%09d.local.test.tld", i+1),
			Qtype: dns.TypeA,
			Rcode: dns.RcodeNameError,
		}
		testCases = append(testCases, tcase)
	}

	p := initTestPlugin(t, ruleset)
	ctx := context.TODO()

	resolveTestCases(testCases, p, ctx, t)
}

func TestLookup_Whitelist(t *testing.T) {
	whitelist := make([]string, 0)

	testCases := make([]test.Case, 0)
	for i := 0; i < 10; i++ {
		qname := fmt.Sprintf("testhost-%09d.local.test.tld", i+1)
		whitelist = append(whitelist, qname)

		tcase := test.Case{
			Qname: qname,
			Qtype: dns.TypeA,
			Rcode: dns.RcodeNameError,
		}
		testCases = append(testCases, tcase)
	}

	testCases = append(testCases, initBlockedTestCases()[10:]...)

	p := initTestPlugin(t, BuildRuleset(whitelist, make([]string, 0)))
	ctx := context.TODO()

	resolveTestCases(testCases, p, ctx, t)
}

func TestLookup_Blacklist(t *testing.T) {
	blacklist := make([]string, 0)

	testCases := make([]test.Case, 0)
	for i := 0; i < 10; i++ {
		qname := fmt.Sprintf("testhost-%09d.local.test.tld", i+1)
		blacklist = append(blacklist, qname)

		tcase := test.Case{
			Qname: qname,
			Qtype: dns.TypeA,
			Answer: []dns.RR{
				test.A(fmt.Sprintf("%s. 3600	IN	A 10.1.33.7", qname)),
			},
		}
		testCases = append(testCases, tcase)
	}

	testCases = append(testCases, initAllowedTestCases()[10:]...)

	p := initTestPlugin(t, BuildRuleset(make([]string, 0), blacklist))
	ctx := context.TODO()

	resolveTestCases(testCases, p, ctx, t)
}

func TestLookup_Block(t *testing.T) {
	p := initTestPlugin(t, getEmptyRuleset())
	ctx := context.TODO()

	testCases := initBlockedTestCases()

	resolveTestCases(testCases, p, ctx, t)
}

func TestLookup_Allow(t *testing.T) {
	p := initTestPlugin(t, getEmptyRuleset())
	ctx := context.TODO()

	testCases := initAllowedTestCases()

	resolveTestCases(testCases, p, ctx, t)
}

func resolveTestCases(testCases []test.Case, p *DNSAdBlock, ctx context.Context, t *testing.T) {
	for _, testCase := range testCases {
		m := testCase.Msg()

		rec := dnstest.NewRecorder(&test.ResponseWriter{})
		_, err := p.ServeDNS(ctx, rec, m)
		if err != nil {
			t.Errorf("Expected no error, got %v\n", err)

		}

		resp := rec.Msg
		err = test.SortAndCheck(resp, testCase)
		assert.NoError(t, err)
	}
}

func initAllowedTestCases() []test.Case {
	testCases := make([]test.Case, 0)
	for i := 0; i < 100; i++ {
		tcase := test.Case{
			Qname: fmt.Sprintf("testhost-%03d.local-a.test.tld", i+1), Qtype: dns.TypeA,
			Rcode: dns.RcodeNameError,
		}
		testCases = append(testCases, tcase)
	}

	return testCases
}

func initBlockedTestCases() []test.Case {
	testCases := make([]test.Case, 0)
	for i := 0; i < 100; i++ {
		tcase := test.Case{
			Qname: fmt.Sprintf("testhost-%09d.local.test.tld", i+1), Qtype: dns.TypeA,
			Answer: []dns.RR{
				test.A(fmt.Sprintf("testhost-%09d.local.test.tld. 3600	IN	A 10.1.33.7", i+1)),
			},
		}
		testCases = append(testCases, tcase)
	}
	return testCases
}

func initTestPlugin(t testing.TB, rs ConfiguredRuleSet) *DNSAdBlock {
	blockmap := make(ListMap, 0)
	for i := 0; i < 100; i++ {
		blockmap[fmt.Sprintf("testhost-%09d.local.test.tld", i+1)] = true
	}

	cfg := defaultConfigWithoutRules
	cfg.TargetIP = net.ParseIP("10.1.33.7")
	cfg.TargetIPv6 = net.ParseIP("fe80::9cbd:c3ff:fe28:e133")
	cfg.EnableLogging = true

	p := DNSAdBlock{
		Next:       nxDomainHandler(),
		blacklist:  blockmap,
		Blacklists: []string{"http://localhost:8080/mylist.txt"},
		RuleSet:    rs,
		updater:    nil,
		config:     &cfg,
	}

	return &p
}

func getEmptyRuleset() ConfiguredRuleSet {
	return BuildRuleset(make([]string, 0), make([]string, 0))
}

func nxDomainHandler() test.Handler {
	return test.HandlerFunc(func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeNameError)
		w.WriteMsg(m)
		return dns.RcodeNameError, nil
	})
}
