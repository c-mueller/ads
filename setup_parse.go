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
	"fmt"
	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/plugin"
	"net"
	"net/url"
	"time"
)

type adsPluginConfig struct {
	BlacklistURLs       []string
	WhitelistURLs       []string
	BlacklistFiles      []string
	WhitelistFiles      []string
	BlacklistRules      []string
	WhitelistRules      []string
	RegexBlacklistRules []string
	RegexWhitelistRules []string

	TargetIP   net.IP
	TargetIPv6 net.IP

	HttpListRenewalInterval  time.Duration
	FileListRenewalInterval  time.Duration
	ListRenewalRetryCount    int
	ListRenewalRetryInterval time.Duration

	ListPersistencePath string

	EnableLogging         bool
	EnableAutoUpdate      bool
	EnableListPersistence bool

	WriteNXDomain bool
}

func parsePluginConfiguration(c *caddy.Controller) (*adsPluginConfig, error) {
	config := defaultConfigWithoutRules
	for c.NextBlock() {
		value := c.Val()

		switch value {
		case "default-lists":
			config.BlacklistURLs = append(config.BlacklistURLs, defaultBlacklists...)
		case "strict-default-lists":
			config.BlacklistURLs = append(config.BlacklistURLs, strictDefaultBlacklists...)
			config.WhitelistURLs = append(config.WhitelistURLs, strictDefaultWhitelists...)
		case "unfiltered-strict-default-lists":
			config.BlacklistURLs = append(config.BlacklistURLs, strictDefaultBlacklists...)
		case "blacklist":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No URL found after list token"))
			}
			parsedUrl, err := url.Parse(c.Val())
			if err != nil {
				return nil, plugin.Error("ads", c.Err(fmt.Sprintf("Invaild URL. Got error while parsing %s", err.Error())))
			} else if parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https" && parsedUrl.Scheme != "file" {
				return nil, plugin.Error("ads", c.Err(fmt.Sprintf("Invaild URL. The scheme %s is not supported!", parsedUrl.Scheme)))
			} else if parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https" {
				config.BlacklistURLs = append(config.BlacklistURLs, c.Val())
			} else {
				config.BlacklistFiles = append(config.BlacklistFiles, parsedUrl.Path)
			}
		case "whitelist":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No URL found after list token"))
			}
			parsedUrl, err := url.Parse(c.Val())
			if err != nil {
				return nil, plugin.Error("ads", c.Err(fmt.Sprintf("Invaild URL. Got error while parsing %s", err.Error())))
			} else if parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https" && parsedUrl.Scheme != "file" {
				return nil, plugin.Error("ads", c.Err(fmt.Sprintf("Invaild URL. The scheme %s is not supported!", parsedUrl.Scheme)))
			} else if parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https" {
				config.WhitelistURLs = append(config.WhitelistURLs, c.Val())
			} else {
				config.WhitelistFiles = append(config.WhitelistFiles, parsedUrl.Path)
			}
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
			config.ListRenewalRetryInterval = i
			break
			//TODO Add Options for Failure Retry interval and Failure retry count
		case "list-store":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No filepath for blocklist persistency defined"))
			}
			if config.EnableListPersistence {
				return nil, plugin.Error("ads", c.Err("Only one filepath for blocklist persistency can be defined"))
			}
			path := c.Val()
			//TODO implement check if path is valid
			config.EnableListPersistence = true
			config.ListPersistencePath = path
			break
		case "log":
			config.EnableLogging = true
		case "block":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No name for blacklist (block) entry defined"))
			}
			config.BlacklistRules = append(config.BlacklistRules, c.Val())
			break
		case "block-regex":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No name for blacklist regex (block-regex) entry defined"))
			}
			config.RegexBlacklistRules = append(config.RegexBlacklistRules, c.Val())
			break
		case "permit":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No name for whitelist (permit) entry defined"))
			}
			config.WhitelistRules = append(config.WhitelistRules, c.Val())
			break
		case "permit-regex":
			if !c.NextArg() {
				return nil, plugin.Error("ads", c.Err("No name for whitelist regex (permit-regex) entry defined"))
			}
			config.RegexWhitelistRules = append(config.RegexWhitelistRules, c.Val())
			break
		case "nxdomain":
			config.WriteNXDomain = true
			break
		case "}":
			break
		case "{":
			break
		}
	}

	if len(config.BlacklistURLs) == 0 {
		config.BlacklistURLs = defaultBlacklists
	}
	return &config, nil
}

func buildRulesetFromConfig(cfg *adsPluginConfig) (*ConfiguredRuleSet, error) {
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
