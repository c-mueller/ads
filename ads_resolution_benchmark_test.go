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
	"math/rand"
	"net"
	"testing"
)

const benchmarkSize = 2500000

func BenchmarkBlockSpeed(b *testing.B) {
	log.Infof("Generating %d blockentries for benchmarking", benchmarkSize)
	p := initBenchPlugin(b)
	ctx := context.TODO()

	log.Infof("Generating %d testcases for benchmarking", benchmarkSize)
	testCases := make([]test.Case, 0)
	for i := 0; i < benchmarkSize; i++ {
		tcase := test.Case{
			Qname: fmt.Sprintf("testhost-%09d.local.test.tld", i+1), Qtype: dns.TypeA,
			Answer: []dns.RR{
				test.A(fmt.Sprintf("testhost-%09d.local.test.tld. 3600	IN	A 10.1.33.7", i+1)),
			},
		}
		testCases = append(testCases, tcase)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testCase := testCases[rand.Intn(benchmarkSize)]

		m := testCase.Msg()

		rec := dnstest.NewRecorder(&test.ResponseWriter{})
		_, err := p.ServeDNS(ctx, rec, m)
		if err != nil {
			b.Errorf("Expected no error, got %v\n", err)
			return
		}
	}
}

func initBenchPlugin(t testing.TB) *DNSAdBlock {
	blockmap := make(BlockMap, 0)
	for i := 0; i < benchmarkSize; i++ {
		blockmap[fmt.Sprintf("testhost-%09d.local.test.tld", i+1)] = true
	}
	cfg := defaultConfigWithoutRules
	cfg.TargetIP = net.ParseIP("10.1.33.7")

	p := DNSAdBlock{
		Next:       nxDomainHandler(),
		blockMap:   blockmap,
		BlockLists: []string{"http://localhost:8080/mylist.txt"},
		updater:    nil,
		config:     &cfg,
	}

	return &p
}
