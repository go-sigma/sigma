package compress

import (
	"reflect"
	"testing"
)

func TestCompress(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "TestCompress-1",
			args: args{
				src: "test.txt",
			},
			want:    []byte{31, 139, 8, 0, 0, 0, 0, 0, 4, 255, 0, 17, 0, 238, 255, 104, 101, 108, 108, 111, 32, 120, 105, 109, 97, 103, 101, 114, 33, 33, 33, 10, 1, 0, 0, 255, 255, 170, 130, 92, 143, 17, 0, 0, 0},
			wantErr: false,
		},
		{
			name: "TestCompress-2",
			args: args{
				src: "no-exist.txt",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Compress(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Compress() = %v, want %v", got, tt.want)
			}
		})
	}
}
