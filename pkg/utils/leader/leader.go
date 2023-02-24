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
