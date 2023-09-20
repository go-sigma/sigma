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

package redis

// mux := asynq.NewServeMux()
// for taskType, handler := range tasks {
// 	topic, ok := topics[taskType]
// 	if !ok {
// 		return fmt.Errorf("topic for daemon task %q not found", taskType)
// 	}
// 	mux.HandleFunc(topic, handler)
// }

// go func() {
// 	err := asyncSrv.Run(mux)
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("srv.Run error")
// 	}
// }()
