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

package timewheel

import (
	"time"
)

// TimeWheel time wheel
type TimeWheel interface {
	// TickNext the next tick time
	TickNext(ddl time.Duration)
	// AddRunner add runner
	AddRunner(runner Notify)
	// Stop stop the time wheel
	Stop()
}

const (
	maxTicker = time.Minute * 5
)

// Notify is a function that is called when a timer expires.
type Notify func()

type timeWheel struct {
	maxTicker time.Duration

	stop    chan struct{}
	next    chan struct{}
	stopped chan struct{}

	runner []Notify
}

// NewTimeWheel new time wheel
func NewTimeWheel(maxTickers ...time.Duration) TimeWheel {
	t := &timeWheel{
		next:    make(chan struct{}, 1),
		stop:    make(chan struct{}, 1),
		stopped: make(chan struct{}, 1),

		runner: make([]Notify, 0),
	}
	if len(maxTickers) > 0 {
		t.maxTicker = maxTickers[0]
	} else {
		t.maxTicker = maxTicker
	}
	t.runLoop()
	return t
}

// runLoop run loop
func (t *timeWheel) runLoop() {
	go func() {
		ticker := time.NewTicker(t.maxTicker)

		for {
			select {
			case <-ticker.C:
				for _, runner := range t.runner {
					go runner()
				}
			case <-t.next:
				for _, runner := range t.runner {
					go runner()
				}
			case <-t.stop:
				t.stopped <- struct{}{}
				return
			}
		}
	}()
}

// TickNext tick next
func (t *timeWheel) TickNext(ddl time.Duration) {
	go func() {
		<-time.After(ddl)
		t.next <- struct{}{} // tick next
	}()
}

// Stop stop
func (t *timeWheel) Stop() {
	t.stop <- struct{}{}
	<-t.stopped
}

// AddRunner add runner
func (t *timeWheel) AddRunner(runner Notify) {
	t.runner = append(t.runner, runner)
}
