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

package artifact

import (
	"context"

	"github.com/hibiken/asynq"

	"github.com/ximager/ximager/pkg/daemon"
	"github.com/ximager/ximager/pkg/types/enums"
	"github.com/ximager/ximager/pkg/utils"
)

func init() {
	utils.PanicIf(daemon.RegisterTask(enums.DaemonProxyArtifact, runner))
}

// when a new blob is pulled bypass the proxy or pushed a new blob to the registry, the proxy will be notified

func runner(ctx context.Context, _ *asynq.Task) error {
	return nil
}
