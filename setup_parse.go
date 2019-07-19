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
	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/plugin"
	"net"
	"strings"
	"time"
)

type adsPluginConfig struct {
	BlocklistURLs       []string
	BlacklistRules      []string
	WhitelistRules      []string
	RegexBlacklistRules []string
	RegexWhitelistRules []string

	TargetIP   net.IP
	TargetIPv6 net.IP

	BlocklistRenewalInterval      time.Duration
	BlocklistRenewalRetryCount    int
	BlocklistRenewalRetryInterval time.Duration

	BlocklistPersistencePath string

	EnableLogging              bool
	EnableAutoUpdate           bool
	EnableBlocklistPersistence bool
}

func parsePluginConfiguration(c *caddy.Controller) (*adsPluginConfig, error) {
	config := defaultConfigWithoutRules
	for c.NextBlock() {
		value := c.Val()

		switch value {
		case "default-lists":
			config.BlocklistURLs = append(config.BlocklistURLs, defaultBlocklists...)
		case "list":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No URL found after list token"))
			}
			url := c.Val()
			if !strings.HasPrefix(url, "http") || !strings.Contains(url, "://") {
				return nil, plugin.Error("ads", c.Err("Invalid url"))
			}
			config.BlocklistURLs = append(config.BlocklistURLs, url)
		case "target":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No target IP specified"))
			}
			ip := net.ParseIP(c.Val())
			if ip == nil {
				return nil, plugin.Error("ads", c.Err("Invalid target IP specified"))
			}
			config.TargetIP = ip
		case "target-ipv6":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No target IP specified"))
			}
			ip := net.ParseIP(c.Val())
			if ip == nil {
				return nil, plugin.Error("ads", c.Err("Invalid target IP specified"))
			}
			config.TargetIPv6 = ip
		case "disable-auto-update":
			config.EnableAutoUpdate = false
			break
		case "auto-update-interval":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No update interval defined"))
			}
			i, err := time.ParseDuration(c.Val())
			if err != nil {
				return nil, plugin.Error("ads", err)
			}
			config.BlocklistRenewalRetryInterval = i
			break
			//TODO Add Options for Failure Retry interval and Failure retry count
		case "blocklist-file":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No filepath for blocklist persistency defined"))
			}
			if config.EnableBlocklistPersistence {
				return nil, plugin.Error("ads", c.Err("Only one filepath for blocklist persistency can be defined"))
			}
			path := c.Val()
			//TODO implement check if path is valid
			config.EnableBlocklistPersistence = true
			config.BlocklistPersistencePath = path
			break
		case "log":
			config.EnableLogging = true
		case "whitelist":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No name for whitelist entry defined"))
			}
			config.WhitelistRules = append(config.WhitelistRules, c.Val())
			break
		case "blacklist":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No name for blacklist entry defined"))
			}
			config.BlacklistRules = append(config.BlacklistRules, c.Val())
			break
		case "whitelist-regex":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No name for whitelist regex entry defined"))
			}
			config.RegexBlacklistRules = append(config.RegexWhitelistRules, c.Val())
			break
		case "blacklist-regex":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No name for blacklist regex entry defined"))
			}
			config.RegexBlacklistRules = append(config.RegexBlacklistRules, c.Val())
			break
		case "}":
			break
		case "{":
			break
		}
	}

	if len(config.BlocklistURLs) == 0 {
		config.BlocklistURLs = defaultBlocklists
	}
	return &config, nil
}

func buildRulesetFromConfig(cfg *adsPluginConfig) (*RuleSet, error) {
	ruleset := BuildRuleset(cfg.WhitelistRules, cfg.BlacklistRules)

	for _, v := range cfg.RegexWhitelistRules {
		if err := ruleset.AddRegexToWhitelist(v); err != nil {
			return nil, err
		}
	}
	for _, v := range cfg.RegexBlacklistRules {
		if err := ruleset.AddRegexToBlacklist(v); err != nil {
			return nil, err
		}
	}
	return &ruleset, nil
}
