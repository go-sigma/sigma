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
				name: "kubernetes-helm/test/tiller",
			},
			want:    "docker.io",
			want1:   "kubernetes-helm",
			want2:   "kubernetes-helm/test/tiller",
			want3:   "latest",
			wantErr: false,
		},
		{
			name: "test7",
			args: args{
				name: "busybox",
			},
			want:    "",
			want1:   "",
			want2:   "",
			want3:   "",
			wantErr: true,
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
