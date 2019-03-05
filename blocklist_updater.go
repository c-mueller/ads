// Copyright 2018 Christian Müller <dev@c-mueller.xyz>
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
	"fmt"
	"time"
)

type BlocklistUpdater struct {
	Enabled        bool
	UpdateInterval time.Duration
	RetryCount     int
	RetryDelay     time.Duration

	Plugin *DNSAdBlock

	persistBlocklists     bool
	persistencePath       string
	lastPersistenceUpdate time.Time

	updateTicker *time.Ticker
	lastUpdate   *time.Time
}

func (u *BlocklistUpdater) Start() {
	log.Info("Registered Update Hook")
	
	go func() {
		//Sleep 5 seconds to ensure coredns is up and running
		time.Sleep(5 * time.Second)

		if !u.persistBlocklists || !exists(u.persistencePath) {
			bm, err := GenerateBlockageMap(u.Plugin.BlockLists)
			if err != nil {
				panic("Failed to fetch blocklists")
			}
			u.Plugin.blockMap = bm
			persistLoadedBlocklist(u, u.Enabled, u.Plugin.BlockLists, bm, u.persistencePath)
		} else {
			storedBlocklist, err := ReadBlocklistConfiguration(u.persistencePath)
			if err != nil {
				panic(fmt.Sprintf("Loading persisted blocklist from %q failed", u.persistencePath))
			}
			if storedBlocklist.NeedsUpdate(u.UpdateInterval) && u.Enabled ||
				!validateBlocklistEquality(u.Plugin.BlockLists, storedBlocklist.Blocklists) && u.Enabled ||
				!u.Enabled {
				bm, err := GenerateBlockageMap(u.Plugin.BlockLists)
				if err != nil {
					panic("Failed to fetch blocklists")
				}
				u.Plugin.blockMap = bm
				persistLoadedBlocklist(u, u.Enabled, u.Plugin.BlockLists, bm, u.persistencePath)
			} else {
				u.Plugin.blockMap = storedBlocklist.BlockedNames

				log.Infof("Loaded Blocklist Length: %d", len(storedBlocklist.BlockedNames))
				log.Infof("Blocklist Length: %d", len(u.Plugin.blockMap))

				u.lastPersistenceUpdate = time.Unix(int64(storedBlocklist.UpdateTimestamp), 0)
			}
		}

		if u.Enabled {
			go u.run()
		}
	}()
}

func (u *BlocklistUpdater) run() {
	log.Info("Running update loop")
	if u.persistBlocklists {
		sleepDuration := u.lastPersistenceUpdate.Add(u.UpdateInterval).Sub(time.Now())
		log.Infof("Scheduled next update in %s", sleepDuration.String())
		time.Sleep(sleepDuration)

		u.handleBlocklistUpdate()
	}

	u.updateTicker = time.NewTicker(u.UpdateInterval)

	for range u.updateTicker.C {
		u.handleBlocklistUpdate()
		log.Infof("Scheduled next update in %s at %s", u.UpdateInterval.String(), time.Now().Add(u.UpdateInterval).String())
	}
}

func (u *BlocklistUpdater) handleBlocklistUpdate() {
	failCount := 0
	for failCount < u.RetryCount {
		log.Infof("Updating blocklists...")

		blockMap, err := GenerateBlockageMap(u.Plugin.BlockLists)
		if err == nil {
			u.Plugin.blockMap = blockMap

			lastUpdate := time.Now()
			u.lastUpdate = &lastUpdate

			if u.persistBlocklists {
				persistedBlocklist := StoredBlocklistConfiguration{
					UpdateTimestamp: int(time.Now().Unix()),
					Blocklists:      u.Plugin.BlockLists,
					BlockedNames:    blockMap,
				}

				err := persistedBlocklist.Persist(u.persistencePath)
				if err == nil {
					u.lastPersistenceUpdate = time.Now()
				} else {
					log.Error("Persisting blocklists failed.")
				}
			}

			log.Info("Blocklists have been updated")

			break
		}

		log.Errorf("Attempt %d/%d failed. Error %q", failCount+1, u.RetryCount, err.Error())

		failCount++
		time.Sleep(u.RetryDelay)
	}
}
