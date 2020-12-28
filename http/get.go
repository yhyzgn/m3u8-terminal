// Copyright 2020 yhyzgn m3u8
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// author : 颜洪毅
// e-mail : yhyzgn@gmail.com
// time   : 2020-12-28 21:34
// version: 1.0.0
// desc   :

package http

import (
	"io/ioutil"
	"net/http"
)

func Get(url string) (bs []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if nil != err {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if nil != err {
		return
	}
	defer resp.Body.Close()
	bs, err = ioutil.ReadAll(resp.Body)
	return
}
