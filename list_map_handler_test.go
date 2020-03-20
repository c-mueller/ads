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
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestListFetch(t *testing.T) {
	hostlist, err := os.Open("testdata/test_blocklist")
	defer hostlist.Close()
	assert.NoError(t, err)
	hostlistData, err := ioutil.ReadAll(hostlist)

	handlerFunc := func(w http.ResponseWriter, req *http.Request) {
		w.Write(hostlistData)
		w.Header().Set("Content-Type", "text/plain")
	}

	srv := httptest.NewServer(http.HandlerFunc(handlerFunc))
	defer srv.Close()

	queryUrl := fmt.Sprintf("%s/hosts.txt", srv.URL)

	list, err := GenerateListMapFromHTTPUrls([]string{queryUrl})
	assert.NoError(t, err)

	expectedList, err := os.Open("testdata/test_blocklist_expected_domains")
	defer expectedList.Close()
	assert.NoError(t, err)
	expData, err := ioutil.ReadAll(expectedList)

	for _, url := range strings.Split(string(expData), "\n") {
		t.Logf("Expected QName: %q Found: %v", url, list[url])
		assert.True(t, list[url])
	}
	assert.False(t, list["testme.com"])
}
