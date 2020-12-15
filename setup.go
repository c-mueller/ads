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
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
)

const Version = "0.2.5"

func init() {
	caddy.RegisterPlugin("ads", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	// Modify based on current version
	// Currently no (useful) automated procedure known
	// ToDo Investigate automated options for this
	log.Infof("Initializing CoreDNS 'ads' plugin. Version %s", Version)
	c.Next()
	cfg, err := parsePluginConfiguration(c)
	if err != nil {
		return err
	}

	updater := &ListUpdater{
		Enabled:         cfg.EnableAutoUpdate,
		RetryCount:      cfg.ListRenewalRetryCount,
		RetryDelay:      cfg.ListRenewalRetryInterval,
		UpdateInterval:  cfg.HttpListRenewalInterval,
		Plugin:          nil,
		persistLists:    cfg.EnableListPersistence,
		persistencePath: cfg.ListPersistencePath,
	}

	c.OnStartup(func() error {
		once.Do(func() {
			prometheus.MustRegister(requestCountTotal, blockedRequestCountTotal, blockedRequestCount)
			updater.Start()
		})
		return nil
	})

	bl := make(ListMap, 0)
	wl := make(ListMap, 0)

	ruleset, err := buildRulesetFromConfig(cfg)
	if err != nil {
		return err
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {

		adsPlugin := DNSAdBlock{
			Next:              next,
			ConfiguredRuleSet: *ruleset,
			FileRuleSet:       *NewFileRuleSet(cfg.WhitelistFiles, cfg.BlacklistFiles),
			blacklist:         bl,
			whitelist:         wl,
			config:            cfg,
		}

		updater.Plugin = &adsPlugin

		if !cfg.EnableAutoUpdate {
			adsPlugin.updater = updater
		}

		return &adsPlugin
	})

	return nil
}

func (u *ListUpdater) persistLoadedHttpLists() {
	u.lastPersistenceUpdate = time.Now()
	if u.Enabled {
		persistedBlocklist := StoredListConfiguration{
			UpdateTimestamp: int(time.Now().Unix()),
			BlacklistURLs:   u.Plugin.config.BlacklistURLs,
			WhitelistURLs:   u.Plugin.config.WhitelistURLs,
			Blacklist:       u.Plugin.blacklist,
			Whitelist:       u.Plugin.whitelist,
		}
		persistedBlocklist.Persist(u.persistencePath)
	}
}
