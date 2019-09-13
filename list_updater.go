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
	"fmt"
	"time"
)

type ListUpdater struct {
	Enabled        bool
	UpdateInterval time.Duration
	RetryCount     int
	RetryDelay     time.Duration

	Plugin *DNSAdBlock

	persistLists          bool
	persistencePath       string
	lastPersistenceUpdate time.Time

	updateTicker *time.Ticker
	lastUpdate   *time.Time
}

func (u *ListUpdater) Start() {
	log.Info("Registered Update Hook")

	go func() {
		//Sleep 250 MS to ensure coredns is up and running
		time.Sleep(250 * time.Millisecond)

		if !u.persistLists || !exists(u.persistencePath) {
			bm, err := GenerateListMapFromHTTPUrls(u.Plugin.Blacklists)
			if err != nil {
				log.Error(err)
				return
			}
			u.Plugin.blacklist = bm
			persistLoadedBlocklist(u, u.Enabled, u.Plugin.config.BlacklistURLs, bm, u.persistencePath)
		} else {
			storedBlocklist, err := ReadListConfiguration(u.persistencePath)
			if err != nil {
				panic(fmt.Sprintf("Loading persisted blocklist from %q failed", u.persistencePath))
			}
			if storedBlocklist.NeedsUpdate(u.UpdateInterval) && u.Enabled ||
				!validateBlocklistEquality(u.Plugin.config.BlacklistURLs, storedBlocklist.Blocklists) && u.Enabled ||
				!u.Enabled {
				bm, err := GenerateListMapFromHTTPUrls(u.Plugin.config.BlacklistURLs)
				if err != nil {
					log.Error(err)
					return
				}
				u.Plugin.blacklist = bm
				persistLoadedBlocklist(u, u.Enabled, u.Plugin.config.BlacklistURLs, bm, u.persistencePath)
			} else {
				u.Plugin.blacklist = storedBlocklist.BlockedNames

				log.Infof("Loaded Blocklist Length: %d", len(storedBlocklist.BlockedNames))
				log.Infof("Blocklist Length: %d", len(u.Plugin.blacklist))

				u.lastPersistenceUpdate = time.Unix(int64(storedBlocklist.UpdateTimestamp), 0)
			}
		}

		if u.Enabled {
			go u.run()
		}
	}()
}

func (u *ListUpdater) run() {
	log.Info("Running update loop")
	if u.persistLists {
		sleepDuration := u.lastPersistenceUpdate.Add(u.UpdateInterval).Sub(time.Now())
		log.Infof("Scheduled next update in %s", sleepDuration.String())
		time.Sleep(sleepDuration)

		u.handleListUpdate()
	}

	u.updateTicker = time.NewTicker(u.UpdateInterval)

	for range u.updateTicker.C {
		u.handleListUpdate()
		log.Infof("Scheduled next update in %s at %s", u.UpdateInterval.String(), time.Now().Add(u.UpdateInterval).String())
	}
}

func (u *ListUpdater) fetchHTTPLists() (ListMap, ListMap, error) {
	blacklist, err := GenerateListMapFromHTTPUrls(u.Plugin.config.BlacklistURLs)
	if err != nil {
		return nil, nil, err
	}
	whitelist, err := GenerateListMapFromHTTPUrls(u.Plugin.config.WhitelistURLs)
	if err != nil {
		return nil, nil, err
	}
	return whitelist, blacklist, nil
}

func (u *ListUpdater) fetchFileLists() (ListMap, ListMap, error) {
	blacklist, err := GenerateListMapFromFileUrls(u.Plugin.config.BlacklistFiles)
	if err != nil {
		return nil, nil, err
	}
	whitelist, err := GenerateListMapFromFileUrls(u.Plugin.config.WhitelistFiles)
	if err != nil {
		return nil, nil, err
	}
	return whitelist, blacklist, nil
}

func (u *ListUpdater) handleListUpdate() {
	failCount := 0
	for failCount < u.RetryCount {
		log.Infof("Updating lists...")

		whitelist, blacklist, err := u.fetchHTTPLists()
		if err != nil {
			log.Errorf("Attempt %d/%d failed. Error %q", failCount+1, u.RetryCount, err.Error())
			failCount++
			time.Sleep(u.RetryDelay)
			break
		}
		u.Plugin.blacklist = blacklist
		u.Plugin.whitelist = whitelist

		lastUpdate := time.Now()
		u.lastUpdate = &lastUpdate

		if u.persistLists {
			persistedList := StoredListConfiguration{
				UpdateTimestamp: int(time.Now().Unix()),
				Blocklists:      u.Plugin.config.BlacklistURLs,
				BlockedNames:    blacklist,
			}

			err := persistedList.Persist(u.persistencePath)
			if err == nil {
				u.lastPersistenceUpdate = time.Now()
			} else {
				log.Error("Persisting blocklists failed.")
			}
		}
		log.Info("Blocklists have been updated")
	}
}
