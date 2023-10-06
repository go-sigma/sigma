// Copyright 2023 sigma
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

package cronjob

import "time"

const (
	// cronjobIterDuration each job iterate duration
	CronjobIterDuration = time.Second * 30
	// tickNextDuration tick the next runner if current get full of jobs
	TickNextDuration = time.Second * 3
	// maxJob each iterate get the maximum jobs
	MaxJob = 100
)

// Starter ...
var Starter []func()

// Stopper ...
var Stopper []func()

// Initialize ...
func Initialize() {
	for _, start := range Starter {
		start()
	}
}
