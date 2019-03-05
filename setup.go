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

	persistBlocklist := false
	persistedBlocklistPath := ""

	enableStats := false
	statEndpoint := ":11022"
	statMode := "inmemory"

	whitelistEntries := make([]string, 0)
	blacklistEntries := make([]string, 0)

	whitelistRegexEntries := make([]string, 0)
	blacklistRegexEntries := make([]string, 0)

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
			if !strings.HasPrefix(url, "http") || !strings.Contains(url, "://") {
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
		case "blocklist-file":
			if !c.NextArg() {
				return plugin.Error("ads", c.Err("No filepath for blocklist persistency defined"))
			}
			if persistBlocklist {
				return plugin.Error("ads", c.Err("Only one filepath for blocklist persistency can be defined"))
			}
			path := c.Val()
			//TODO implement check if path is valid
			persistBlocklist = true
			persistedBlocklistPath = path
			break
		case "log":
			logBlocks = true
		case "whitelist":
			if !c.NextArg() {
				return plugin.Error("ads", c.Err("No name for whitelist entry defined"))
			}
			whitelistEntries = append(whitelistEntries, c.Val())
			break
		case "blacklist":
			if !c.NextArg() {
				return plugin.Error("ads", c.Err("No name for blacklist entry defined"))
			}
			blacklistEntries = append(blacklistEntries, c.Val())
			break
		case "whitelist-regex":
			if !c.NextArg() {
				return plugin.Error("ads", c.Err("No name for whitelist regex entry defined"))
			}
			whitelistRegexEntries = append(whitelistRegexEntries, c.Val())
			break
		case "blacklist-regex":
			if !c.NextArg() {
				return plugin.Error("ads", c.Err("No name for blacklist regex entry defined"))
			}
			blacklistRegexEntries = append(blacklistRegexEntries, c.Val())
			break
		case "stats":
			enableStats = true
			// TODO Implement Endpoint and dbmode configuration
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
		Enabled:           enableAutoUpdate,
		RetryCount:        renewalAttemptCount,
		RetryDelay:        failureRetryDelay,
		UpdateInterval:    renewalInterval,
		Plugin:            nil,
		persistBlocklists: persistBlocklist,
		persistencePath:   persistedBlocklistPath,
	}

	var repo StatRepository
	if statMode == "inmemory" {
		repo = &MemoryStatHandler{}
	} else {
		repo = &MemoryStatHandler{}
	}

	statHandler := StatHandler{
		Enabled:  enableStats,
		Endpoint: statEndpoint,
		Repo:     repo,
	}

	if err := statHandler.Init(); err != nil {
		return err
	}

	c.OnStartup(func() error {
		once.Do(func() {
			metrics.MustRegister(c, requestCount)
			metrics.MustRegister(c, blockedRequestCount)
			metrics.MustRegister(c, requestCountBySource)
			metrics.MustRegister(c, blockedRequestCountBySource)
			updater.Start()
		})
		return nil
	})

	blockageMap := make(BlockMap, 0)

	ruleset := BuildRuleset(whitelistEntries, blacklistEntries)

	for _, v := range whitelistRegexEntries {
		if err := ruleset.AddRegexToWhitelist(v); err != nil {
			return err
		}
	}
	for _, v := range blacklistRegexEntries {
		if err := ruleset.AddRegexToBlacklist(v); err != nil {
			return err
		}
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {

		adsPlugin := DNSAdBlock{
			Next:        next,
			BlockLists:  blocklists,
			blockMap:    blockageMap,
			RuleSet:     ruleset,
			TargetIP:    targetIP,
			LogBlocks:   logBlocks,
			StatHandler: statHandler,
		}

		updater.Plugin = &adsPlugin

		if !enableAutoUpdate {
			adsPlugin.updater = updater
		}

		return &adsPlugin
	})

	return nil
}

func persistLoadedBlocklist(updater *BlocklistUpdater, enableAutoUpdate bool, blocklists []string, blockageMap BlockMap, persistedBlocklistPath string) {
	updater.lastPersistenceUpdate = time.Now()
	if enableAutoUpdate {
		persistedBlocklist := StoredBlocklistConfiguration{
			UpdateTimestamp: int(time.Now().Unix()),
			Blocklists:      blocklists,
			BlockedNames:    blockageMap,
		}
		persistedBlocklist.Persist(persistedBlocklistPath)
	}
}
