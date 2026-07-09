// Copyright 2026 sigma
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

package app

import (
	"sync/atomic"
	"time"
)

// DrainDelay is the time to wait after flipping the draining flag before
// starting http.Server.Shutdown. It gives the load balancer a window to
// propagate the /readyz=503 state and stop routing new connections.
const DrainDelay = 5 * time.Second

var draining atomic.Bool

// SetDraining toggles the draining flag. When true, /readyz starts returning
// 503 so k8s stops routing traffic to this process.
func SetDraining(b bool) { draining.Store(b) }

// IsDraining reports whether the process is draining traffic before shutdown.
func IsDraining() bool { return draining.Load() }
