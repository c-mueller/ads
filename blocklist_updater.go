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

import "time"

type BlocklistUpdater struct {
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
	u.updateTicker = time.NewTicker(u.UpdateInterval)
	go u.run()
}

func (u *BlocklistUpdater) run() {
	if u.persistBlocklists {
		sleepDuration := u.lastPersistenceUpdate.Add(u.UpdateInterval).Sub(time.Now())
		log.Infof("Scheduled next update in %s", sleepDuration.String())
		time.Sleep(sleepDuration)

		u.handleBlocklistUpdate()
	}
	for range u.updateTicker.C {
		u.handleBlocklistUpdate()
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

		log.Errorf("Attempt %d/%d failed. Error %q%s", failCount+1, u.RetryCount, err.Error(), failCount != u.RetryCount-1)

		failCount++
		time.Sleep(u.RetryDelay)
	}
}
