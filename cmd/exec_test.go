// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ixti/ecs-task-helper/pkg/container_metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecCommand(t *testing.T) {
	t.Run("with successful fetch executes with merged environ", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		var capturedEnv []string

		deps := &execCmdDeps{
			metadataCmdDeps: metadataCmdDeps{
				FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
					return testMetadata(), nil
				},
				Timeout: 5 * time.Second,
			},
			Environ:  func() []string { return []string{"PATH=/usr/bin", "HOME=/home/test"} },
			LookPath: func(file string) (string, error) { return "/bin/" + file, nil },
			Exec: func(argv0 string, argv []string, envv []string) error {
				capturedEnv = envv
				return nil
			},
		}

		cmd := NewExecCommand(deps)
		cmd.SetArgs([]string{"--", "sh", "-c", "echo hello"})

		err := cmd.Execute()

		require.NoError(err)
		assert.Contains(capturedEnv, "PATH=/usr/bin")
		assert.Contains(capturedEnv, "HOME=/home/test")
		assert.Contains(capturedEnv, "ECS_CONTAINER_NAME=curl")
		assert.Contains(capturedEnv, "ECS_CLUSTER_NAME=default")
	})

	t.Run("with missing metadata URI uses empty metadata", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		var capturedEnv []string

		deps := &execCmdDeps{
			metadataCmdDeps: metadataCmdDeps{
				FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
					return nil, container_metadata.ErrMissingMetadataURI
				},
				Timeout: 5 * time.Second,
			},
			Environ:  func() []string { return []string{"PATH=/usr/bin", "CUSTOM=value"} },
			LookPath: func(file string) (string, error) { return "/bin/" + file, nil },
			Exec: func(argv0 string, argv []string, envv []string) error {
				capturedEnv = envv
				return nil
			},
		}

		cmd := NewExecCommand(deps)
		cmd.SetArgs([]string{"sh"})

		err := cmd.Execute()

		require.NoError(err)
		assert.Contains(capturedEnv, "PATH=/usr/bin")
		assert.Contains(capturedEnv, "CUSTOM=value")
		assert.Contains(capturedEnv, "ECS_CONTAINER_NAME=")
		assert.Contains(capturedEnv, "ECS_CLUSTER_NAME=")
	})

	t.Run("with fetch error continues with empty metadata", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		var capturedEnv []string

		deps := &execCmdDeps{
			metadataCmdDeps: metadataCmdDeps{
				FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
					return nil, errors.New("network error")
				},
				Timeout: 5 * time.Second,
			},
			Environ:  func() []string { return []string{"PATH=/usr/bin", "CUSTOM=value"} },
			LookPath: func(file string) (string, error) { return "/bin/" + file, nil },
			Exec: func(argv0 string, argv []string, envv []string) error {
				capturedEnv = envv
				return nil
			},
		}

		cmd := NewExecCommand(deps)
		cmd.SetArgs([]string{"sh"})

		err := cmd.Execute()

		require.NoError(err)
		assert.Contains(capturedEnv, "PATH=/usr/bin")
		assert.Contains(capturedEnv, "CUSTOM=value")
		assert.Contains(capturedEnv, "ECS_CONTAINER_NAME=")
		assert.Contains(capturedEnv, "ECS_CLUSTER_NAME=")
	})

	t.Run("with LookPath error returns error", func(t *testing.T) {
		assert := assert.New(t)

		lookPathErr := errors.New("executable not found")
		deps := &execCmdDeps{
			metadataCmdDeps: metadataCmdDeps{
				FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
					return testMetadata(), nil
				},
				Timeout: 5 * time.Second,
			},
			Environ:  func() []string { return nil },
			LookPath: func(file string) (string, error) { return "", lookPathErr },
			Exec: func(argv0 string, argv []string, envv []string) error {
				return nil
			},
		}

		cmd := NewExecCommand(deps)
		cmd.SetArgs([]string{"nonexistent"})

		err := cmd.Execute()

		assert.ErrorIs(err, lookPathErr)
	})

	t.Run("with Exec error returns error", func(t *testing.T) {
		assert := assert.New(t)

		execErr := errors.New("exec failed")
		deps := &execCmdDeps{
			metadataCmdDeps: metadataCmdDeps{
				FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
					return testMetadata(), nil
				},
				Timeout: 5 * time.Second,
			},
			Environ:  func() []string { return nil },
			LookPath: func(file string) (string, error) { return "/bin/" + file, nil },
			Exec: func(argv0 string, argv []string, envv []string) error {
				return execErr
			},
		}

		cmd := NewExecCommand(deps)
		cmd.SetArgs([]string{"sh"})

		err := cmd.Execute()

		assert.ErrorIs(err, execErr)
	})

	t.Run("passes correct argv to Exec", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		var capturedArgv0 string
		var capturedArgv []string

		deps := &execCmdDeps{
			metadataCmdDeps: metadataCmdDeps{
				FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
					return testMetadata(), nil
				},
				Timeout: 5 * time.Second,
			},
			Environ:  func() []string { return nil },
			LookPath: func(file string) (string, error) { return "/usr/bin/" + file, nil },
			Exec: func(argv0 string, argv []string, envv []string) error {
				capturedArgv0 = argv0
				capturedArgv = argv
				return nil
			},
		}

		cmd := NewExecCommand(deps)
		cmd.SetArgs([]string{"--", "python", "-c", "print('hello')"})

		err := cmd.Execute()

		require.NoError(err)
		assert.Equal("/usr/bin/python", capturedArgv0)
		assert.Equal([]string{"/usr/bin/python", "-c", "print('hello')"}, capturedArgv)
	})

	t.Run("requires at least one argument", func(t *testing.T) {
		assert := assert.New(t)

		deps := &execCmdDeps{
			metadataCmdDeps: metadataCmdDeps{
				FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
					return testMetadata(), nil
				},
				Timeout: 5 * time.Second,
			},
			Environ:  func() []string { return nil },
			LookPath: func(file string) (string, error) { return "/bin/" + file, nil },
			Exec: func(argv0 string, argv []string, envv []string) error {
				return nil
			},
		}

		cmd := NewExecCommand(deps)
		cmd.SetArgs([]string{})

		err := cmd.Execute()

		assert.Error(err)
	})

	t.Run("with nil deps uses defaults", func(t *testing.T) {
		cmd := NewExecCommand(nil)

		assert.Equal(t, "exec command [args...]", cmd.Use)
		assert.NotNil(t, cmd.RunE)
	})
}
