package ads

import "time"

type BlocklistUpdater struct {
	UpdateInterval time.Duration
	RetryCount     int
	RetryDelay     time.Duration

	Plugin *DNSAdBlock

	updateTicker *time.Ticker
}

func (u *BlocklistUpdater) Start() {
	u.updateTicker = time.NewTicker(u.UpdateInterval)

}

func (u *BlocklistUpdater) run() {
	for range u.updateTicker.C {
		failCount := 0

		for failCount < u.RetryCount {
			blockMap, err := GenerateBlockageMap(u.Plugin.BlockLists)
			if err == nil {

				break
			}

			failCount++
			time.Sleep(u.RetryDelay)
		}
	}
}
