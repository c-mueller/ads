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
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

const firstHostlistPath = "testdata/update_hostlist_test_first_list"
const secondHostlistPath = "testdata/update_hostlist_test_second_list"

func TestBlocklistUpdater(t *testing.T) {
	server := initTestServer(t)
	defer server.Close()

	url := fmt.Sprintf("%s/list.txt", server.URL)

	p := initTestPlugin(t, getEmptyRuleset())

	p.config.BlacklistURLs = []string{url}
	p.blacklist = make(ListMap, 0)

	updater := ListUpdater{
		Enabled:        true,
		Plugin:         p,
		UpdateInterval: time.Second * 2,
		RetryCount:     10,
		RetryDelay:     time.Second * 1,
	}

	p.updater = &updater

	p.updater.Start()

	time.Sleep(time.Second * 1)
	assert.Equal(t, 1000, len(p.blacklist))

	time.Sleep(time.Second * 5)
	assert.Equal(t, 2000, len(p.blacklist))

	p.updater.httpUpdateTicker.Stop()
}

func TestBlocklistUpdaterWithBadList(t *testing.T) {
	p := initTestPlugin(t, getEmptyRuleset())

	p.config.BlacklistURLs = []string{"https://badhost/doesnotexist"}
	p.blacklist = make(ListMap, 0)

	updater := ListUpdater{
		Enabled:        false,
		Plugin:         p,
		UpdateInterval: time.Second * 2,
		RetryCount:     10,
		RetryDelay:     time.Second * 1,
	}

	p.updater = &updater
	p.updater.Start()

	// give it time to fail
	time.Sleep(time.Second * 6)
	assert.Equal(t, 0, len(p.blacklist))
}

func initTestServer(t *testing.T) *httptest.Server {
	firstPath, err := os.Open(firstHostlistPath)
	assert.NoError(t, err)
	defer firstPath.Close()
	firstData, err := ioutil.ReadAll(firstPath)

	secondPath, err := os.Open(secondHostlistPath)
	assert.NoError(t, err)
	defer secondPath.Close()
	secondData, err := ioutil.ReadAll(secondPath)

	firstServed := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !firstServed {
			w.Write(firstData)
			firstServed = true
		} else {
			w.Write(secondData)
		}
		w.Header().Set("Content-Type", "text/plain")
	}))

	return server
}
