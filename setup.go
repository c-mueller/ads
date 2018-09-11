// Copyright 2018 Christian Müller <cmueller.dev@gmail.com>
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
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/mholt/caddy"
	"net"
	"strings"
	"time"
)

func init() {
	caddy.RegisterPlugin("ads", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	c.Next()

	blocklists := make([]string, 0)
	targetIP := net.ParseIP(defaultResolutionIP)
	logBlocks := false

	enableAutoUpdate := true
	renewalAttemptCount := 5
	failureRetryDelay := time.Minute * 1
	renewalInterval := time.Hour * 24

	for c.NextBlock() {
		value := c.Val()

		switch value {
		case "default-lists":
			blocklists = append(blocklists, defaultBlocklists...)
		case "list":
			if !c.NextArg() {
				return plugin.Error("ads", c.Err("No URL found after list token"))
			}
			url := c.Val()
			if !strings.HasPrefix(url,"http") || !strings.Contains(url,"://") {
				return plugin.Error("ads", c.Err("Invalid url"))
			}
			blocklists = append(blocklists, url)
		case "target":
			if !c.NextArg() {
				return plugin.Error("ads", c.Err("No target IP specified"))
			}
			ip := net.ParseIP(c.Val())
			if ip == nil {
				return plugin.Error("ads", c.Err("Invalid target IP specified"))
			}
			targetIP = ip
		case "disable-auto-update":
			enableAutoUpdate = false
			break
		case "auto-update-interval":
			if !c.NextArg() {
				return plugin.Error("ads", c.Err("No update interval defined"))
			}
			i, err := time.ParseDuration(c.Val())
			if err != nil {
				return plugin.Error("ads", err)
			}
			renewalInterval = i
			break
			//TODO Add Options for Failure Retry interval and Failure retry count
		case "log":
			logBlocks = true
			// Do Nothing in case of { or }
		case "}":
			break
		case "{":
			break
		}
	}

	if len(blocklists) == 0 {
		blocklists = defaultBlocklists
	}

	updater := &BlocklistUpdater{
		RetryCount:     renewalAttemptCount,
		RetryDelay:     failureRetryDelay,
		UpdateInterval: renewalInterval,
		Plugin:         nil,
	}

	c.OnStartup(func() error {
		once.Do(func() {
			metrics.MustRegister(c, requestCount)
			metrics.MustRegister(c, blockedRequestCount)
			metrics.MustRegister(c, requestCountBySource)
			metrics.MustRegister(c, blockedRequestCountBySource)
			if enableAutoUpdate {
				updater.Start()
			}
		})
		return nil
	})

	blockageMap, err := GenerateBlockageMap(blocklists)
	if err != nil {
		return plugin.Error("ads", c.Err("Failed to fetch blocklists"))
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {

		adsPlugin := DNSAdBlock{
			Next:       next,
			BlockLists: blocklists,
			blockMap:   blockageMap,
			TargetIP:   targetIP,
			LogBlocks:  logBlocks,
		}

		updater.Plugin = &adsPlugin

		if !enableAutoUpdate {
			adsPlugin.updater = updater
		}

		return adsPlugin
	})

	return nil
}
