// Copyright 2023 XImager
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

package distribution

import "net/http"

type Transport struct {
	roundTripper http.RoundTripper
	funcs        func(*http.Request)
}

// RoundTrip handles each http request
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.funcs(req)
	return t.roundTripper.RoundTrip(req)
}

// NewTransport creates a new Transport
func NewTransport(funcs func(*http.Request)) http.RoundTripper {
	var tran = &Transport{
		roundTripper: http.DefaultTransport,
		funcs:        funcs,
	}
	return tran
}
