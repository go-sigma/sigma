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
	"io"
	"path"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/go-sigma/sigma/pkg/builder"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	builder.DriverFactories[path.Base(reflect.TypeOf(factory{}).PkgPath())] = &factory{}
}

type factory struct{}

var _ builder.Factory = factory{}

// New returns a new filesystem storage driver
func (f factory) New(config configs.Configuration) (builder.Builder, error) {
	i := &instance{}

	var err error
	var restConfig *restclient.Config
	if config.Daemon.Builder.Kubernetes.Kubeconfig != nil {
		cfg := clientcmdapi.NewConfig()
		err := yaml.Unmarshal([]byte(ptr.To(config.Daemon.Builder.Kubernetes.Kubeconfig)), &cfg)
		if err != nil {
			return nil, fmt.Errorf("Decode kubeconfig failed: %v", err)
		}
		clientConfig := clientcmd.NewDefaultClientConfig(ptr.To(cfg), &clientcmd.ConfigOverrides{})
		restConfig, err = clientConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("Get k8s rest config failed: %v", err)
		}
	} else {
		restConfig, err = clientcmd.BuildConfigFromFlags("", "")
		if err != nil {
			return nil, fmt.Errorf("Get k8s client in cluster failed: %v", err)
		}
	}

	i.client, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("Get reset client failed: %v", err)
	}

	go i.informer(context.Background())

	return i, nil
}

type instance struct {
	config configs.Configuration
	client *kubernetes.Clientset
}

// Start start a container to build oci image and push to registry
func (i instance) Start(ctx context.Context, builderConfig builder.BuilderConfig) error {
	envs, err := builder.BuildK8sEnv(builderConfig)
	if err != nil {
		return err
	}
	_, err = i.client.CoreV1().Pods(i.config.Daemon.Builder.Kubernetes.Namespace).Create(ctx, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%d-%d", consts.AppName, builderConfig.BuilderID, builderConfig.RunnerID),
			Labels: map[string]string{
				"oci-image-builder": consts.AppName,
				"builder-id":        strconv.FormatInt(builderConfig.BuilderID, 10),
				"runner-id":         strconv.FormatInt(builderConfig.RunnerID, 10),
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image:   "docker.io/library/builder:dev",
					Command: []string{"sigma-builder"},
					Env:     envs,
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("Create pod failed: %v", err)
	}
	return nil
}

// Stop stop the container
func (i instance) Stop(ctx context.Context, builderID, runnerID int64) error {
	podName := fmt.Sprintf("%s-%d-%d", consts.AppName, builderID, runnerID)
	return i.client.CoreV1().Pods(i.config.Daemon.Builder.Kubernetes.Namespace).
		Delete(ctx, podName, metav1.DeleteOptions{})
}

// Restart wrap stop and start
func (i instance) Restart(ctx context.Context, builderConfig builder.BuilderConfig) error {
	podName := fmt.Sprintf("%s-%d-%d", consts.AppName, builderConfig.BuilderID, builderConfig.RunnerID)
	err := i.client.CoreV1().Pods(i.config.Daemon.Builder.Kubernetes.Namespace).Delete(ctx, podName, metav1.DeleteOptions{
		PropagationPolicy: ptr.Of(metav1.DeletePropagationForeground),
	})
	if err != nil {
		return err
	}
	return i.Start(ctx, builderConfig)
}

// LogStream get the real time log stream
func (i instance) LogStream(ctx context.Context, builderID, runnerID int64, writer io.Writer) error {
	podName := fmt.Sprintf("%s-%d-%d", consts.AppName, builderID, runnerID)
	reader, err := i.client.CoreV1().Pods(i.config.Daemon.Builder.Kubernetes.Namespace).
		GetLogs(podName, &corev1.PodLogOptions{
			Follow: true,
		}).Stream(ctx)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, reader)
	if err != nil {
		return err
	}
	return nil
}
