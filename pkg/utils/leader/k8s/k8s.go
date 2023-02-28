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
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/ximager/ximager/pkg/utils/leader"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

const (
	name = "k8s"
)

type k8sLeaderElector struct {
	leader *leaderelection.LeaderElector
}

func init() {
	err := leader.RegisterLeaderFactory(name, &factory{})
	if err != nil {
		panic(fmt.Sprintf("fail to register leader factory: %v", err))
	}
}

type factory struct{}

var _ leader.Factory = &factory{}

// New ...
func (f factory) New(opts leader.Options) (leader.LeaderElector, error) {
	podName := os.Getenv("POD_NAME")
	podNamespace := os.Getenv("POD_NAMESPACE")

	kubeconfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("fail to get kubeconfig: %w", err)
	}

	rcl, err := resourcelock.NewFromKubeconfig(resourcelock.LeasesResourceLock, podNamespace, opts.Name, resourcelock.ResourceLockConfig{Identity: podName}, kubeconfig, opts.RenewDeadline)
	if err != nil {
		return nil, fmt.Errorf("fail to create resource lock: %w", err)
	}
	leaderElector, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:            rcl,
		ReleaseOnCancel: true,
		LeaseDuration:   opts.LeaseDuration,
		RenewDeadline:   opts.RenewDeadline,
		RetryPeriod:     opts.RetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(c context.Context) {},
			OnStoppedLeading: func() {
				log.Info().Msg("no longer the leader, staying inactive")
			},
			OnNewLeader: func(current_id string) {
				if current_id == podName {
					log.Info().Msg("still the leader")
					return
				}
				log.Info().Str("id", current_id).Msg("new leader changed")
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("fail to create leader elector: %w", err)
	}
	go leaderElector.Run(context.Background())

	return k8sLeaderElector{
		leader: leaderElector,
	}, nil
}

// IsLeader returns whether the current pod is the leader
func (l k8sLeaderElector) IsLeader() bool {
	if l.leader == nil {
		return false
	}
	return l.leader.IsLeader()
}
