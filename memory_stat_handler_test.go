// Copyright 2018 - 2019 Christian Müller <dev@c-mueller.xyz>
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
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	e := MemoryStatHandler{}
	assert.NoError(t, e.Init())

	insertTestData(e, t)

	assert.Equal(t, 36*3600, len(e.RequestList))
}

func TestGetRange(t *testing.T) {
	e := MemoryStatHandler{}
	assert.NoError(t, e.Init())

	td := time.Now()
	insertTestData(e, t)

	elems := e.GetRequestsBetween(td.Add(24*-1*time.Hour), td.Add(12*-1*time.Hour))

	assert.True(t, len(elems) >= 43195 && len(elems) <= 43205)
}

func TestDelete(t *testing.T) {
	e := MemoryStatHandler{}
	assert.NoError(t, e.Init())

	insertTestData(e, t)

	cnt := 0
	for k, _ := range e.RequestList {
		assert.NoError(t, e.Delete(k))
		cnt++
		if cnt >= 500 {
			break
		}
	}

	assert.Equal(t, 36*3600-500, len(e.RequestList))

	reqcnt := 0
	for _, v := range e.Requests {
		for range v {
			reqcnt++
		}
	}

	assert.Equal(t, 36*3600-500, reqcnt)
}

func TestCleanup(t *testing.T) {
	e := MemoryStatHandler{}
	assert.NoError(t, e.Init())

	insertTestData(e, t)

	t.Logf("Before: %d", len(e.RequestList))

	assert.NoError(t, e.Cleanup())

	ln := len(e.RequestList)

	t.Logf("After: %d", ln)

	assert.True(t, ln >= 86390 && ln <= 86410, "Length not in Range")
}

func TestStats(t *testing.T) {
	e := MemoryStatHandler{}
	assert.NoError(t, e.Init())

	insertTestData(e, t)

	data, err := json.MarshalIndent(e.GetStats(), "", "  ")
	assert.NoError(t, err)
	fmt.Println(string(data))
}

func TestDelete_EmptyInput(t *testing.T) {
	e := MemoryStatHandler{}
	assert.NoError(t, e.Init())
	assert.Error(t, e.Delete(""))
}

func TestDelete_UnknownInput(t *testing.T) {
	e := MemoryStatHandler{}
	assert.NoError(t, e.Init())
	assert.Error(t, e.Delete("idk"))
}

func insertTestData(e MemoryStatHandler, t *testing.T) {
	for i := 0; i < 36*3600; i++ {
		r := Request{
			Blocked:           i%2 == 0,
			Timestamp:         time.Now().Add(time.Duration(i) * -1 * time.Second).Unix(),
			Recipient:         fmt.Sprintf("10.10.10.%d", i%256),
			RequestedHostname: fmt.Sprintf("www.my-%09d.xyz", i%(rand.Intn(20)+1)),
		}
		uid, err := e.Insert(r)
		assert.NoError(t, err)
		assert.NotEmpty(t, uid)
	}
}
