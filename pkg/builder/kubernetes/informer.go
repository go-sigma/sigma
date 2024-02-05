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

package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func (i *instance) informer(_ context.Context) {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(i.client, time.Second*30,
		informers.WithNamespace(i.config.Daemon.Builder.Kubernetes.Namespace))
	podInformer := informerFactory.Core().V1().Pods().Informer()
	podEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("New Pod created: %s/%s\n", pod.Namespace, pod.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod := oldObj.(*corev1.Pod)
			newPod := newObj.(*corev1.Pod)
			fmt.Printf("Pod updated: %s/%s\n", newPod.Namespace, newPod.Name)
			fmt.Printf("Pod updated: %s/%s\n", oldPod.Namespace, oldPod.Name)
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("Pod deleted: %s/%s\n", pod.Namespace, pod.Name)
		},
	}
	_, err := podInformer.AddEventHandler(podEventHandler)
	if err != nil {
		log.Error().Err(err).Msg("Informer add event handler failed")
	}
}
