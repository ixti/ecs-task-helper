// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: MIT

package container_metadata

import "strings"

type Metadata struct {
	ContainerARN          string `json:"containerARN"`
	ContainerName         string `json:"containerName"`
	ContainerImage        string `json:"containerImage"`
	TaskARN               string `json:"taskARN"`
	TaskDefinitionFamily  string `json:"taskDefinitionFamily"`
	TaskDefinitionVersion string `json:"taskDefinitionVersion"`
	ClusterName           string `json:"clusterName"`
}

// TaskID returns TaskID part of TaskARN.
func (m *Metadata) TaskID() string {
	if m.ClusterName == "" {
		return ""
	}

	if _, id, ok := strings.Cut(m.TaskARN, ":task/"+m.ClusterName+"/"); ok {
		return id
	}

	return ""
}

// EnvironWith returns ECS metadata as environment variables.
// If base is nil, returns only the ECS metadata variables.
// If base is provided, returns base with ECS metadata variables merged in
// (overriding any existing).
func (m *Metadata) EnvironWith(base []string) []string {
	overrides := []string{
		"ECS_CONTAINER_ARN=" + m.ContainerARN,
		"ECS_CONTAINER_NAME=" + m.ContainerName,
		"ECS_CONTAINER_IMAGE=" + m.ContainerImage,
		"ECS_TASK_ARN=" + m.TaskARN,
		"ECS_TASK_ID=" + m.TaskID(),
		"ECS_TASK_DEFINITION_FAMILY=" + m.TaskDefinitionFamily,
		"ECS_TASK_DEFINITION_VERSION=" + m.TaskDefinitionVersion,
		"ECS_CLUSTER_NAME=" + m.ClusterName,
	}

	if base == nil {
		return overrides
	}

	keys := make(map[string]struct{}, len(overrides))
	for _, v := range overrides {
		key, _, _ := strings.Cut(v, "=")
		keys[key] = struct{}{}
	}

	merged := make([]string, 0, len(base)+len(overrides))
	for _, v := range base {
		key, _, _ := strings.Cut(v, "=")
		if _, exists := keys[key]; !exists {
			merged = append(merged, v)
		}
	}

	return append(merged, overrides...)
}

// Environ returns only the ECS metadata environment variables.
// Equivalent to EnvironWith(nil).
func (m *Metadata) Environ() []string {
	return m.EnvironWith(nil)
}
