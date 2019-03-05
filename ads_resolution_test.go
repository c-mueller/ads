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
	"context"
	"fmt"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestLookup_Block(t *testing.T) {
	p := initTestPlugin(t)
	ctx := context.TODO()

	testCases := initBlockedTestCases()

	for _, testCase := range testCases {
		m := testCase.Msg()

		rec := dnstest.NewRecorder(&test.ResponseWriter{})
		_, err := p.ServeDNS(ctx, rec, m)
		if err != nil {
			t.Errorf("Expected no error, got %v\n", err)
			return
		}

		resp := rec.Msg
		err = test.SortAndCheck( resp, testCase)
		assert.NoError(t, err)
	}
}

func TestLookup_Allow(t *testing.T) {
	p := initTestPlugin(t)
	ctx := context.TODO()

	testCases := initAllowedTestCases()

	for _, testCase := range testCases {
		m := testCase.Msg()

		rec := dnstest.NewRecorder(&test.ResponseWriter{})
		_, err := p.ServeDNS(ctx, rec, m)
		if err != nil {
			t.Errorf("Expected no error, got %v\n", err)
			return
		}

		resp := rec.Msg
		err = test.SortAndCheck( resp, testCase)
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

func initTestPlugin(t testing.TB) *DNSAdBlock {
	blockmap := make(BlockMap, 0)
	for i := 0; i < 100; i++ {
		blockmap[fmt.Sprintf("testhost-%09d.local.test.tld", i+1)] = true
	}
	p := DNSAdBlock{
		Next:       nxDomainHandler(),
		blockMap:   blockmap,
		BlockLists: []string{"http://localhost:8080/mylist.txt"},
		updater:    nil,
		LogBlocks:  true,
		TargetIP:   net.ParseIP("10.1.33.7"),
	}

	return &p
}

func nxDomainHandler() test.Handler {
	return test.HandlerFunc(func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeNameError)
		w.WriteMsg(m)
		return dns.RcodeNameError, nil
	})

}
