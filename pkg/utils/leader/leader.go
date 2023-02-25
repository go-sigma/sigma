// The MIT License (MIT)
//
// Copyright © 2023 Tosone <i@tosone.cn>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
