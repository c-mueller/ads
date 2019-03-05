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
	"bytes"
	gz "compress/gzip"
	"io/ioutil"
	"os"
)

func validateBlocklistEquality(a, b []string) bool {
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
