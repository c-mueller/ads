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
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

type ListMap map[string]bool

var ValidateQName = regexp.MustCompile("([a-zA-Z0-9]|\\.|-)*").MatchString

func GenerateListMap(urls []string, fetchFunc func(ref string) ([]byte, error)) (ListMap, error) {
	listMap := make(ListMap, 0)
	for _, listUrl := range urls {
		log.Infof("Fetching list %q...", listUrl)

		data, err := fetchFunc(listUrl)
		if err != nil {
			return nil, err
		}
		parseListFile(data, listMap)
	}
	log.Infof("Found %d unique domains to in list", len(listMap))
	return listMap, nil
}

func GenerateListMapFromHTTPUrls(listUrls []string) (ListMap, error) {
	fetcher := func(u string) ([]byte, error) {
		content, err := http.Get(u)
		if err != nil {
			return nil, err
		}
		defer content.Body.Close()

		data, err := ioutil.ReadAll(content.Body)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	return GenerateListMap(listUrls, fetcher)
}

func GenerateListMapFromFileUrls(listUrls []string) (ListMap, error) {
	fetcher := func(u string) ([]byte, error) {
		u = strings.TrimPrefix(u, "file://")
		stream, err := os.Open(u)
		if err != nil {
			return nil, err
		}

		defer stream.Close()

		data, err := ioutil.ReadAll(stream)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	return GenerateListMap(listUrls, fetcher)
}

func parseListFile(data []byte, blockageMap ListMap) {
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
				// Count up to determine last index
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
		if ValidateQName(url) && utf8.Valid([]byte(url)) {
			blockageMap[url] = true
			urlCount++
		}
	}
	log.Infof("Fetched %d entries.", urlCount)
}

func cleanHostsLine(line string) string {
	ln := strings.TrimSuffix(line, " ")
	ln = strings.Replace(line, " ", "\t", -1)
	ln = strings.Replace(ln, "\r", "", -1)
	// Escape quotes to prevent compialtion issues
	// Of course entries containing such data are useless
	ln = strings.Replace(ln, "\"", "\\\"", -1)
	return ln
}
