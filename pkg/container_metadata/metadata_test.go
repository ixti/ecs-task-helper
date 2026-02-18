// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: MIT

package container_metadata

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testMetadata() *Metadata {
	return &Metadata{
		ContainerARN:          "arn:aws:ecs:us-west-2:111122223333:container/0206b271-b33f-47ab-86c6-a0ba208a70a9",
		ContainerName:         "curl",
		ContainerImage:        "111122223333.dkr.ecr.us-west-2.amazonaws.com/curltest:latest",
		TaskARN:               "arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
		TaskDefinitionFamily:  "curltest",
		TaskDefinitionVersion: "24",
		ClusterName:           "default",
	}
}

func expectedOverrides() []string {
	return []string{
		"ECS_CONTAINER_ARN=arn:aws:ecs:us-west-2:111122223333:container/0206b271-b33f-47ab-86c6-a0ba208a70a9",
		"ECS_CONTAINER_NAME=curl",
		"ECS_CONTAINER_IMAGE=111122223333.dkr.ecr.us-west-2.amazonaws.com/curltest:latest",
		"ECS_TASK_ARN=arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
		"ECS_TASK_ID=8f03e41243824aea923aca126495f665",
		"ECS_TASK_DEFINITION_FAMILY=curltest",
		"ECS_TASK_DEFINITION_VERSION=24",
		"ECS_CLUSTER_NAME=default",
	}
}

func TestMetadata_TaskID(t *testing.T) {
	t.Run("with valid TaskARN", func(t *testing.T) {
		metadata := &Metadata{
			ClusterName: "default",
			TaskARN:     "arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
		}

		assert.Equal(t, "8f03e41243824aea923aca126495f665", metadata.TaskID())
	})

	t.Run("with invalid TaskARN", func(t *testing.T) {
		metadata := &Metadata{
			ClusterName: "deadbeef",
			TaskARN:     "arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
		}

		assert.Equal(t, "", metadata.TaskID())
	})

	t.Run("with blank TaskARN", func(t *testing.T) {
		metadata := &Metadata{}
		assert.Equal(t, "", metadata.TaskID())
	})
}

func TestMetadata_ToJSON(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	data, err := json.Marshal(testMetadata())
	require.NoError(err)

	var result map[string]string
	require.NoError(json.Unmarshal(data, &result))

	assert.Equal(map[string]string{
		"containerARN":          "arn:aws:ecs:us-west-2:111122223333:container/0206b271-b33f-47ab-86c6-a0ba208a70a9",
		"containerName":         "curl",
		"containerImage":        "111122223333.dkr.ecr.us-west-2.amazonaws.com/curltest:latest",
		"taskARN":               "arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
		"taskDefinitionFamily":  "curltest",
		"taskDefinitionVersion": "24",
		"clusterName":           "default",
	}, result)
}

func TestMetadata_Environ(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(expectedOverrides(), testMetadata().Environ())
}

func TestMetadata_EnvironWith(t *testing.T) {
	t.Run("with nil base returns only overrides", func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(expectedOverrides(), testMetadata().EnvironWith(nil))
	})

	t.Run("with empty base returns only overrides", func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(expectedOverrides(), testMetadata().EnvironWith([]string{}))
	})

	t.Run("replaces existing metadata env vars", func(t *testing.T) {
		assert := assert.New(t)

		base := []string{
			"ECS_CONTAINER_NAME=old-value",
			"ECS_TASK_ARN=old-task-arn",
			"PATH=/usr/bin",
		}

		env := testMetadata().EnvironWith(base)

		assert.NotContains(env, "ECS_CONTAINER_NAME=old-value")
		assert.NotContains(env, "ECS_TASK_ARN=old-task-arn")
		assert.Contains(env, "ECS_CONTAINER_NAME=curl")
		assert.Contains(env, "ECS_TASK_ARN=arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665")
		assert.Contains(env, "PATH=/usr/bin")
	})

	t.Run("preserves unrelated env vars", func(t *testing.T) {
		assert := assert.New(t)

		base := []string{
			"PATH=/usr/bin",
			"HOME=/home/test",
			"CUSTOM_VAR=custom-value",
			"ECS_SOME_OTHER_VAR=should-remain",
		}

		env := testMetadata().EnvironWith(base)

		assert.Contains(env, "PATH=/usr/bin")
		assert.Contains(env, "HOME=/home/test")
		assert.Contains(env, "CUSTOM_VAR=custom-value")
		assert.Contains(env, "ECS_SOME_OTHER_VAR=should-remain")
	})

	t.Run("overrides appear at end", func(t *testing.T) {
		assert := assert.New(t)

		base := []string{"PATH=/usr/bin", "HOME=/home/test"}
		env := testMetadata().EnvironWith(base)

		overrides := expectedOverrides()
		envLen := len(env)
		overridesLen := len(overrides)

		assert.Equal(envLen, len(base)+overridesLen)
		assert.Equal(overrides, env[envLen-overridesLen:])
	})
}
