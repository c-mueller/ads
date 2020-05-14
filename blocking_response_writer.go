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
	"strings"
)

type BlockingResponseWriter struct {
	Writer       dns.ResponseWriter
	Plugin       *DNSAdBlock
	Request      *dns.Msg
	RequestState *request.Request
}

func (b *BlockingResponseWriter) LocalAddr() net.Addr {
	return b.Writer.LocalAddr()
}

func (b *BlockingResponseWriter) RemoteAddr() net.Addr {
	return b.Writer.RemoteAddr()
}

func (b *BlockingResponseWriter) WriteMsg(msg *dns.Msg) error {
	for _, rr := range msg.Answer {
		host := ""
		switch v := rr.(type) {
		case *dns.CNAME:
			host = strings.TrimSuffix(v.Target, ".")
		case *dns.A:
			host = strings.TrimSuffix(v.Hdr.Name, ".")
		case *dns.AAAA:
			host = strings.TrimSuffix(v.Hdr.Name, ".")
		default:
			continue
		}
		if b.Plugin.ShouldBlock(host) {
			return b.Plugin.onBlock(b, b.Request, b.RequestState, host)
		}
	}
	return b.Writer.WriteMsg(msg)
}

func (b *BlockingResponseWriter) Write(bytes []byte) (int, error) {
	log.Warning("'ads' called with Write: CNAME blocking therefore does not work")
	return b.Writer.Write(bytes)
}

func (b *BlockingResponseWriter) Close() error {
	return b.Writer.Close()
}

func (b *BlockingResponseWriter) TsigStatus() error {
	return b.Writer.TsigStatus()
}

func (b *BlockingResponseWriter) TsigTimersOnly(b2 bool) {
	b.Writer.TsigTimersOnly(b2)
}

func (b *BlockingResponseWriter) Hijack() {
	b.Writer.Hijack()
}
