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
	"github.com/Flaque/filet"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_Blockfile_Write_Read(t *testing.T) {
	m := loadBlockMap(t)
	tmpdir := filet.TmpDir(t, "")
	defer filet.CleanUp(t)

	datapath := filepath.Join(tmpdir, "coredns_ads_blockdata.json.gz")
	t.Log(datapath)

	config := StoredListConfiguration{
		Blacklist:       m,
		BlacklistURLs:   []string{"http://localhost:8888/blocklist.txt"},
		UpdateTimestamp: int(time.Now().Unix()),
	}

	err := config.Persist(datapath)
	assert.NoError(t, err)

	reloadedConfig, err := ReadListConfiguration(datapath)
	assert.NoError(t, err)

	assert.Equal(t, config.UpdateTimestamp, reloadedConfig.UpdateTimestamp)
	assert.Equal(t, config.BlacklistURLs, reloadedConfig.BlacklistURLs)
	assert.Equal(t, config.Blacklist, reloadedConfig.Blacklist)
}

func loadBlockMap(t *testing.T) (ListMap) {
	file, err := os.Open("testdata/update_hostlist_test_second_list")
	defer file.Close()
	assert.NoError(t, err)
	data, err := ioutil.ReadAll(file)

	m := make(ListMap, 0)

	parseListFile(data, m)

	return m
}
