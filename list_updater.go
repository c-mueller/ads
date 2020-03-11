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

	httpUpdateTicker *time.Ticker
	fileUpdateTicker *time.Ticker
	lastUpdate       *time.Time
}

func (u *ListUpdater) Start() {
	log.Info("Initializing CoreDNS 'ads' list update routines...")

	go func() {
		//Sleep 250 MS to ensure coredns is up and running
		time.Sleep(250 * time.Millisecond)

		if !u.persistLists || !exists(u.persistencePath) {
			wl, bl, err := u.fetchHTTPLists()
			if err != nil {
				log.Error(err)
				return
			}
			u.Plugin.blacklist = bl
			u.Plugin.whitelist = wl
			u.persistLoadedHttpLists()
		} else {
			storedListSet, err := ReadListConfiguration(u.persistencePath)
			if err != nil {
				panic(fmt.Sprintf("Loading persisted blocklist from %q failed", u.persistencePath))
			}
			if storedListSet.NeedsUpdate(u.UpdateInterval) && u.Enabled ||
				!validateURLListEquality(u.Plugin.config.BlacklistURLs, storedListSet.BlacklistURLs) && u.Enabled ||
				!validateURLListEquality(u.Plugin.config.WhitelistURLs, storedListSet.WhitelistURLs) && u.Enabled ||
				!u.Enabled {
				wl, bl, err := u.fetchHTTPLists()
				if err != nil {
					log.Error(err)
					return
				}
				u.Plugin.blacklist = bl
				u.Plugin.whitelist = wl
				u.persistLoadedHttpLists()
			} else {
				u.Plugin.whitelist = storedListSet.Whitelist
				u.Plugin.blacklist = storedListSet.Blacklist

				log.Infof("Loaded Whitelist (HTTP) Length: %d", len(storedListSet.Whitelist))
				log.Infof("Loaded Blacklist (HTTP) Length: %d", len(storedListSet.Blacklist))

				u.lastPersistenceUpdate = time.Unix(int64(storedListSet.UpdateTimestamp), 0)
			}
		}

		go u.runFileUpdater()

		if u.Enabled {
			go u.runHttpUpdater()
		}
	}()
}

func (u *ListUpdater) runFileUpdater() {
	u.fileUpdateTicker = time.NewTicker(u.Plugin.config.FileListRenewalInterval)
	u.handleFileUpdate()

	for range u.fileUpdateTicker.C {
		u.handleFileUpdate()
	}
}

func (u *ListUpdater) handleFileUpdate() {
	log.Info("Updating lists from Local files...")

	wl, bl, err := u.fetchFileLists()
	if err != nil {
		log.Errorf("Loading File lists has failed. Error message: %q", err.Error())
		return
	}
	u.Plugin.FileRuleSet.Whitelist = wl
	u.Plugin.FileRuleSet.Blacklist = bl
}

func (u *ListUpdater) runHttpUpdater() {
	log.Info("Updating lists from HTTP URLs...")
	if u.persistLists {
		sleepDuration := u.lastPersistenceUpdate.Add(u.UpdateInterval).Sub(time.Now())
		log.Infof("Scheduled next update in %s", sleepDuration.String())
		time.Sleep(sleepDuration)

		u.handleHTTPListUpdate()
	}

	u.httpUpdateTicker = time.NewTicker(u.UpdateInterval)

	for range u.httpUpdateTicker.C {
		u.handleHTTPListUpdate()
		log.Infof("Scheduled next update of HTTP lists in %s at %s", u.UpdateInterval.String(), time.Now().Add(u.UpdateInterval).String())
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
	log.Infof("[HTTP Update] Loaded %d entries into Blacklist and %d entries into whitelist", len(blacklist), len(whitelist))
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
	log.Infof("[File Update] Loaded %d entries into Blacklist and %d entries into whitelist", len(blacklist), len(whitelist))
	return whitelist, blacklist, nil
}

func (u *ListUpdater) handleHTTPListUpdate() {
	log.Infof("Updating and Persisting HTTP lists...")
	failCount := 0
	for failCount < u.RetryCount {

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
				BlacklistURLs:   u.Plugin.config.BlacklistURLs,
				WhitelistURLs:   u.Plugin.config.WhitelistURLs,
				Blacklist:       blacklist,
				Whitelist:       whitelist,
			}

			err := persistedList.Persist(u.persistencePath)
			if err == nil {
				u.lastPersistenceUpdate = time.Now()
			} else {
				log.Error("Persisting HTTP Lists failed.")
			}
		}
		log.Info("Lists with HTTP URLs have been updated")
	}
}
