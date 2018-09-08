package ads

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func GenerateBlockageMap(blocklistUrls []string) (BlockMap, error) {
	blockageMap := make(BlockMap, 0)
	for _, blocklistURL := range blocklistUrls {
		log.Infof("Fetching blocklist %q...", blocklistURL)
		content, err := http.Get(blocklistURL)
		if err != nil {
			return nil, err
		}

		data, err := ioutil.ReadAll(content.Body)
		if err != nil {
			return nil, err
		}

		urlCount := 0
		for _, line := range strings.Split(string(data), "\n") {
			// Skip lines containing comments
			if strings.Contains(line, "#") {
				continue
			}

			ln := cleanHostsLine(line)
			substrings := strings.Split(ln, "\t")

			url := ""

			if len(substrings) == 0 {
				continue
			} else if len(substrings) == 1 {
				url = substrings[0]
			} else {
				i := 1
				for ; len(substrings[i]) == 0 && i < len(substrings)-1; i++ {

				}

				if len(substrings) == i {
					continue
				}

				url = substrings[i]
			}

			if url == "" {
				continue
			}

			// Enable blocking for url
			blockageMap[url] = true
			urlCount++
		}
		log.Infof("Fetched %d entries.", urlCount)

	}

	log.Infof("Registered %d unique domains to block", len(blockageMap))

	return blockageMap, nil

}

func cleanHostsLine(line string) string {
	ln := strings.TrimSuffix(line, " ")
	ln = strings.Replace(line, " ", "\t", -1)
	ln = strings.Replace(ln, "\r", "", -1)
	return ln
}
