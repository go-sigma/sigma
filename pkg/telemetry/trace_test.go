// Copyright 2026 sigma
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

package telemetry

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"github.com/go-sigma/sigma/pkg/config"
)

func TestInit(t *testing.T) {
	prevTP := otel.GetTracerProvider()
	prevPropagator := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevPropagator)
	})

	tests := []struct {
		name    string
		cfg     config.ConfigurationTrace
		wantErr bool
		wantNil bool
	}{
		{
			name:    "disabled",
			cfg:     config.ConfigurationTrace{Enabled: false},
			wantNil: true,
		},
		{
			name: "http",
			cfg:  config.ConfigurationTrace{Enabled: true, Protocol: "http", Endpoint: "localhost:4318"},
		},
		{
			name: "http default protocol",
			cfg:  config.ConfigurationTrace{Enabled: true, Endpoint: "localhost:4318"},
		},
		{
			name: "http insecure with custom service name and ratio",
			cfg: config.ConfigurationTrace{
				Enabled:     true,
				Protocol:    "http",
				Endpoint:    "localhost:4318",
				Insecure:    true,
				SampleRatio: 0.5,
				ServiceName: "custom",
			},
		},
		{
			name: "grpc",
			cfg:  config.ConfigurationTrace{Enabled: true, Protocol: "grpc", Endpoint: "localhost:4317", Insecure: true},
		},
		{
			name:    "grpc exporter error",
			cfg:     config.ConfigurationTrace{Enabled: true, Protocol: "grpc", Endpoint: "\x00", Insecure: true},
			wantErr: true,
			wantNil: true,
		},
		{
			name:    "unknown protocol",
			cfg:     config.ConfigurationTrace{Enabled: true, Protocol: "bogus", Endpoint: "localhost:4317"},
			wantErr: true,
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shutdown, err := Init(&config.Configuration{Trace: tt.cfg})
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, shutdown)
				return
			}
			require.NoError(t, err)
			if tt.wantNil {
				require.Nil(t, shutdown)
				return
			}
			require.NotNil(t, shutdown)
			require.NoError(t, shutdown(t.Context()))
		})
	}
}
