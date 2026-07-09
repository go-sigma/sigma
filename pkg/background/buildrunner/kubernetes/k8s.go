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
	"strings"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/go-sigma/sigma/pkg/background/buildrunner"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	buildrunner.DriverFactories[path.Base(reflect.TypeFor[factory]().PkgPath())] = &factory{}
}

type factory struct{}

var _ buildrunner.Factory = factory{}

// New returns a new filesystem storage driver
func (f factory) New(config *config.Configuration) (buildrunner.Builder, error) {
	i := &instance{
		config: config,
	}

	var err error
	var restConfig *restclient.Config
	if strings.TrimSpace(ptr.To(config.Daemon.Builder.Kubernetes.Kubeconfig)) != "" {
		cfg := clientcmdapi.NewConfig()
		err := yaml.Unmarshal([]byte(ptr.To(config.Daemon.Builder.Kubernetes.Kubeconfig)), &cfg)
		if err != nil {
			return nil, fmt.Errorf("decode kubeconfig failed: %v", err)
		}
		clientConfig := clientcmd.NewDefaultClientConfig(ptr.To(cfg), &clientcmd.ConfigOverrides{})
		restConfig, err = clientConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("get k8s rest config failed: %v", err)
		}
	} else {
		restConfig, err = restclient.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("get k8s client in cluster failed: %v", err)
		}
	}

	i.client, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("get reset client failed: %v", err)
	}

	go i.informer(context.Background())

	return i, nil
}

type instance struct {
	config *config.Configuration
	client *kubernetes.Clientset
}

// Start start a container to build oci image and push to registry
func (i instance) Start(ctx context.Context, builderConfig buildrunner.BuilderConfig) error {
	envs, err := buildrunner.BuildK8sEnv(builderConfig)
	if err != nil {
		return err
	}
	_, err = i.client.CoreV1().Pods(i.config.Daemon.Builder.Kubernetes.Namespace).Create(ctx, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: buildrunner.GenContainerID(builderConfig.BuilderID, builderConfig.RunnerID),
			Labels: map[string]string{
				"oci-image-builder": consts.AppName,
				"builder-id":        builderConfig.BuilderID,
				"runner-id":         builderConfig.RunnerID,
			},
		},
		Spec: corev1.PodSpec{
			HostAliases: buildHostAliases(builderConfig.ExtraHosts),
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
		return fmt.Errorf("create pod failed: %v", err)
	}
	return nil
}

// Stop stop the container
func (i instance) Stop(ctx context.Context, builderID, runnerID string) error {
	podName := buildrunner.GenContainerID(builderID, runnerID)
	return i.client.CoreV1().Pods(i.config.Daemon.Builder.Kubernetes.Namespace).
		Delete(ctx, podName, metav1.DeleteOptions{})
}

// Restart wrap stop and start
func (i instance) Restart(ctx context.Context, builderConfig buildrunner.BuilderConfig) error {
	podName := buildrunner.GenContainerID(builderConfig.BuilderID, builderConfig.RunnerID)
	propagationPolicy := metav1.DeletePropagationForeground
	err := i.client.CoreV1().Pods(i.config.Daemon.Builder.Kubernetes.Namespace).Delete(ctx, podName, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		return err
	}
	return i.Start(ctx, builderConfig)
}

// LogStream get the real time log stream
func (i instance) LogStream(ctx context.Context, builderID, runnerID string, writer io.Writer) error {
	podName := buildrunner.GenContainerID(builderID, runnerID)
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

// buildHostAliases 将 ["hostname:ip"] 格式的 ExtraHosts 转换为 Kubernetes HostAlias
func buildHostAliases(extraHosts []string) []corev1.HostAlias {
	if len(extraHosts) == 0 {
		return nil
	}
	var aliases []corev1.HostAlias
	for _, host := range extraHosts {
		parts := strings.SplitN(host, ":", 2)
		if len(parts) != 2 {
			continue
		}
		aliases = append(aliases, corev1.HostAlias{
			IP:        parts[1],
			Hostnames: []string{parts[0]},
		})
	}
	return aliases
}
