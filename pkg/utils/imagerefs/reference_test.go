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

package imagerefs

import (
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		want2   string
		want3   string
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				name: "gcr.io/kubernetes-helm/tiller:v2.11.0",
			},
			want:    "gcr.io",
			want1:   "kubernetes-helm",
			want2:   "kubernetes-helm/tiller",
			want3:   "v2.11.0",
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				name: "gcr.io/kubernetes-helm/test/tiller:v2.11.0",
			},
			want:    "gcr.io",
			want1:   "kubernetes-helm",
			want2:   "kubernetes-helm/test/tiller",
			want3:   "v2.11.0",
			wantErr: false,
		},
		{
			name: "test3",
			args: args{
				name: "-gcr.io/kubernetes-helm/test/tiller:v2.11.0",
			},
			want:    "",
			want1:   "",
			want2:   "",
			want3:   "",
			wantErr: true,
		},
		{
			name: "test4",
			args: args{
				name: "kubernetes-helm/tiller:v1",
			},
			want:    "docker.io",
			want1:   "kubernetes-helm",
			want2:   "kubernetes-helm/tiller",
			want3:   "v1",
			wantErr: false,
		},
		{
			name: "test5",
			args: args{
				name: "kubernetes-helm/test/tiller:v1",
			},
			want:    "docker.io",
			want1:   "kubernetes-helm",
			want2:   "kubernetes-helm/test/tiller",
			want3:   "v1",
			wantErr: false,
		},
		{
			name: "test6",
			args: args{
				name: "kubernetes-helm/test/tiller@sha256:4ca2a277f1dc3ddd0da33a258096de9a1cf5b9d9bd96b27ee78763ee2248c28c",
			},
			want:    "docker.io",
			want1:   "kubernetes-helm",
			want2:   "kubernetes-helm/test/tiller",
			want3:   "sha256:4ca2a277f1dc3ddd0da33a258096de9a1cf5b9d9bd96b27ee78763ee2248c28c",
			wantErr: false,
		},
		{
			name: "test7",
			args: args{
				name: "kubernetes-helm/test/tiller:v1@sha256:4ca2a277f1dc3ddd0da33a258096de9a1cf5b9d9bd96b27ee78763ee2248c28c",
			},
			want:    "docker.io",
			want1:   "kubernetes-helm",
			want2:   "kubernetes-helm/test/tiller",
			want3:   "v1",
			wantErr: false,
		},
		{
			name: "test8",
			args: args{
				name: "kubernetes-helm/test/tiller",
			},
			want:    "docker.io",
			want1:   "kubernetes-helm",
			want2:   "kubernetes-helm/test/tiller",
			want3:   "latest",
			wantErr: false,
		},
		{
			name: "test9",
			args: args{
				name: "busybox",
			},
			want:    "",
			want1:   "",
			want2:   "",
			want3:   "",
			wantErr: true,
		},
		{
			name: "test10",
			args: args{
				name: "library/busybox",
			},
			want:    "docker.io",
			want1:   "library",
			want2:   "library/busybox",
			want3:   "latest",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, got3, err := Parse(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Parse() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("Parse() got2 = %v, want %v", got2, tt.want2)
			}
			if got3 != tt.want3 {
				t.Errorf("Parse() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}
