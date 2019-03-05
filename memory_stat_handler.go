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
	"fmt"
	"github.com/google/uuid"
	"time"
)

type MemoryStatHandler struct {
	Requests    map[int64][]*Request
	RequestList map[string]Request
}

func (h *MemoryStatHandler) Init() error {
	h.Requests = make(map[int64][]*Request)
	h.RequestList = make(map[string]Request)
	return nil
}

func (h *MemoryStatHandler) Close() error {

	return nil
}

func (h *MemoryStatHandler) Insert(request Request) (string, error) {
	uid := uuid.New().String()

	request.RequestID = uid

	h.RequestList[uid] = request

	if h.Requests[request.Timestamp] == nil {
		h.Requests[request.Timestamp] = make([]*Request, 0)
	}

	h.Requests[request.Timestamp] = append(h.Requests[request.Timestamp], &request)

	return uid, nil
}

func (h *MemoryStatHandler) Delete(uid string) error {
	if h.RequestList[uid].RequestID != uid {
		return fmt.Errorf("%q not Found", uid)
	}

	r := h.RequestList[uid]

	if len(h.Requests[r.Timestamp]) <= 1 {
		delete(h.Requests, r.Timestamp)
	} else {
		idx := -1
		for k, v := range h.Requests[r.Timestamp] {
			if v.RequestID == uid {
				idx = k
				break
			}
		}
		if idx != -1 {
			h.Requests[r.Timestamp] = append(h.Requests[r.Timestamp][:idx], h.Requests[r.Timestamp][idx+1:]...)
		}
	}

	delete(h.RequestList, uid)
	return nil
}

func (h *MemoryStatHandler) GetRequestsBetween(from, to time.Time) []Request {
	requests := make([]Request, 0)

	fromU := from.Unix()
	toU := to.Unix()

	for time, elems := range h.Requests {
		if time >= fromU && time <= toU {
			for _, v := range elems {
				requests = append(requests, *v)
			}
		}
	}

	return requests
}

func (h *MemoryStatHandler) GetStats() Stats {
	panic("implement me")
}

func (h *MemoryStatHandler) Cleanup() error {

	return nil
}
