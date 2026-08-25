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

package build

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
)

func TestCoordinatorMarkCompleted(t *testing.T) {
	ctx := t.Context()
	repo := &testRunnerRepository{}
	coordinator := NewCoordinator(repo, nil)

	require.NoError(t, coordinator.MarkCompleted(ctx, "builder", "runner", 0))
	require.Equal(t, enums.BuildStatusSuccess, repo.lastUpdate(builderRunnerStatusColumn))
	require.NotZero(t, repo.lastUpdate(builderRunnerEndedAtColumn))

	require.NoError(t, coordinator.MarkCompleted(ctx, "builder", "runner", 1))
	require.Equal(t, enums.BuildStatusFailed, repo.lastUpdate(builderRunnerStatusColumn))
}

func TestCoordinatorMarkStopped(t *testing.T) {
	ctx := t.Context()
	repo := &testRunnerRepository{}
	coordinator := NewCoordinator(repo, nil)
	expectedErr := errors.New("No such container")

	require.NoError(t, coordinator.MarkStopped(ctx, "builder", "runner", expectedErr, func(err error) bool {
		return err == expectedErr
	}))
	require.Equal(t, enums.BuildStatusStopped, repo.lastUpdate(builderRunnerStatusColumn))

	require.NoError(t, coordinator.MarkStopped(ctx, "builder", "runner", errors.New("kill failed"), func(error) bool {
		return false
	}))
	require.Equal(t, enums.BuildStatusFailed, repo.lastUpdate(builderRunnerStatusColumn))
}

func TestCoordinatorStoreLogs(t *testing.T) {
	logStore := &testLogStore{}
	coordinator := NewCoordinator(&testRunnerRepository{}, logStore)

	err := coordinator.StoreLogs("builder", "runner", func(stdout, stderr io.Writer) error {
		_, err := stdout.Write([]byte("stdout"))
		require.NoError(t, err)
		_, err = stderr.Write([]byte("stderr"))
		return err
	})

	require.NoError(t, err)
	require.Equal(t, "builder", logStore.builderID)
	require.Equal(t, "runner", logStore.runnerID)
	require.Equal(t, "stdoutstderr", logStore.writer.String())
	require.True(t, logStore.writer.closed)
}

type testRunnerRepository struct {
	updates []map[string]any
}

func (r *testRunnerRepository) UpdateRunner(_ context.Context, _, _ string, updates map[string]any) error {
	clone := make(map[string]any, len(updates))
	for key, value := range updates {
		clone[key] = value
	}
	r.updates = append(r.updates, clone)
	return nil
}

func (r *testRunnerRepository) lastUpdate(key string) any {
	return r.updates[len(r.updates)-1][key]
}

type testLogStore struct {
	builderID string
	runnerID  string
	writer    *testWriteCloser
}

func (s *testLogStore) Write(builderID, runnerID string) io.WriteCloser {
	s.builderID = builderID
	s.runnerID = runnerID
	s.writer = &testWriteCloser{}
	return s.writer
}

func (s *testLogStore) Read(_ context.Context, _ string) (io.Reader, error) {
	return nil, nil
}

type testWriteCloser struct {
	bytes.Buffer
	closed bool
}

func (w *testWriteCloser) Close() error {
	w.closed = true
	return nil
}
