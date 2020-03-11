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
	"bytes"
	gz "compress/gzip"
	"io/ioutil"
	"os"
	"sort"
)

func validateURLListEquality(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	lm := make(map[string]bool, 0)
	for _, v := range a {
		lm[v] = true
	}

	for _, v := range b {
		if !lm[v] {
			return false
		}
	}
	return true
}

func exists(path string) (bool) {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func gzip(data []byte) ([]byte, error) {
	var outputBuffer bytes.Buffer
	compressionWriter := gz.NewWriter(&outputBuffer)
	_, err := compressionWriter.Write(data)
	if err != nil {
		return nil, err
	}
	compressionWriter.Close()

	return outputBuffer.Bytes(), nil
}

func gunzip(data []byte) ([]byte, error) {
	inputBuffer := bytes.NewReader(data)
	compressionReader, err := gz.NewReader(inputBuffer)
	if err != nil {
		return nil, err
	}

	defer compressionReader.Close()

	return ioutil.ReadAll(compressionReader)
}

type pair struct {
	Key   string
	Value int
}
type pairs []pair

func (p pairs) Len() int           { return len(p) }
func (p pairs) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p pairs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func keepHighestValues(input map[string]int, count int) map[string]int {
	if len(input) <= count {
		return input
	}
	elems := make(pairs, len(input))

	idx := 0
	for k, v := range input {
		elems[idx] = pair{k, v}
		idx++
	}

	sort.Sort(elems)

	valmap := make(map[string]int)
	for i := 0; i < count; i++ {
		e := elems[i]
		valmap[e.Key] = e.Value
	}

	return valmap
}
