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

package leader

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Options struct {
	Name          string
	LeaseDuration time.Duration
	RenewDeadline time.Duration
	RetryPeriod   time.Duration
}

// LeaderElector is the interface for the leader elector
type LeaderElector interface {
	IsLeader() bool
}

// Factory is the interface for the storage driver factory
type Factory interface {
	New(opts Options) (LeaderElector, error)
}

var leaderFactories = make(map[string]Factory)

// RegisterLeaderFactory registers a leader factory by name.
func RegisterLeaderFactory(name string, factory Factory) error {
	if _, ok := leaderFactories[name]; ok {
		return fmt.Errorf("leader %q already registered", name)
	}
	leaderFactories[name] = factory
	return nil
}

// Leader is the leader elector
var Leader LeaderElector

func Initialize() error {
	typ := viper.GetString("leader.type")
	factory, ok := leaderFactories[typ]
	if !ok {
		return fmt.Errorf("leader %q not registered", typ)
	}

	var opts = Options{
		Name:          viper.GetString("leader.name"),
		LeaseDuration: viper.GetDuration("leader.leaseDuration"),
		RenewDeadline: viper.GetDuration("leader.renewDeadline"),
		RetryPeriod:   viper.GetDuration("leader.retryPeriod"),
	}

	if opts.Name == "" {
		return fmt.Errorf("leader name is empty")
	}

	var err error
	Leader, err = factory.New(opts)
	if err != nil {
		return err
	}
	return nil
}
