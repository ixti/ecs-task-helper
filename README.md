# ECS Task Helper Utilities

[![CI](https://github.com/ixti/ecs-task-helper/actions/workflows/ci.yml/badge.svg)](https://github.com/ixti/ecs-task-helper/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/ixti/ecs-task-helper/graph/badge.svg?token=toUztJc66F)](https://codecov.io/gh/ixti/ecs-task-helper)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A lightweight CLI tool providing utilities for AWS ECS containers.
It fetches ECS task metadata and exposes it as environment variables,
making container introspection simple.


## Features

- Fetch ECS container metadata from the [Task Metadata Endpoint V4](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v4.html)
- Export metadata as environment variables or JSON
- Execute commands with metadata automatically injected into the environment
- Lightweight HTTP health check utility
- Minimal Docker image built from scratch (~2MB)
- Multi-architecture support (linux/amd64, linux/arm64)


## Installation

### Docker

```sh
docker pull ghcr.io/ixti/ecs-task-helper:latest
```

### From Source

```sh
go install github.com/ixti/ecs-task-helper@latest
```

## Usage

### `metadata` - Print ECS Metadata

Fetches and prints ECS container metadata in the specified format.

```sh
# Print as environment variables (default)
ecs-task-helper metadata

# Print as JSON
ecs-task-helper metadata --format json
```

**Output environment variables:**

| Environment Variable            | JSON Key                  | Description                   |
| ------------------------------- | ------------------------- | ----------------------------- |
| `ECS_CONTAINER_ARN`             | `containerArn`            | ARN of the container          |
| `ECS_CONTAINER_NAME`            | `containerName`           | Name of the container         |
| `ECS_CONTAINER_IMAGE`           | `containerImage`          | Container image               |
| `ECS_TASK_ARN`                  | `taskArn`                 | ARN of the ECS task           |
| `ECS_TASK_ID`                   | -                         | ID of the ECS task            |
| `ECS_TASK_DEFINITION_FAMILY`    | `taskDefinitionFamily`    | Task definition family name   |
| `ECS_TASK_DEFINITION_VERSION`   | `taskDefinitionVersion`   | Task definition version       |
| `ECS_CLUSTER_NAME`              | `clusterName`             | Name of the ECS cluster       |

### `exec` - Execute with Metadata Environment

Executes a command with ECS metadata automatically injected as environment variables.
Uses `execve(2)` to replace the process, making it ideal for container entrypoints.

```sh
# Run your application with ECS metadata in the environment
ecs-task-helper exec /app/myservice

# Pass arguments to the command
ecs-task-helper exec /app/myservice --port 8080
```

If the ECS metadata endpoint is not available (e.g., running locally),
the command executes with the current environment and logs a warning.

### `check` - HTTP Health Check

A lightweight HTTP client for health checks. Returns exit code 0 on success, 1 on failure.

```sh
# Basic health check (expects HTTP 200)
ecs-task-helper check http://localhost:8080/health

# Custom timeout
ecs-task-helper check --timeout 5s http://localhost:8080/health

# Accept multiple status codes
ecs-task-helper check --status 200,204 http://localhost:8080/health

# Quiet mode (suppress response body)
ecs-task-helper check --quiet http://localhost:8080/health
```

**Flags:**

| Flag          | Default   | Description                                                  |
| ------------- | --------- | ------------------------------------------------------------ |
| `--timeout`   | `1s`      | Request timeout                                              |
| `--status`    | `200`     | Expected HTTP status codes (can be specified multiple times) |
| `--quiet`     | `false`   | Suppress response body output                                |

## Configuration

| Environment Variable                    | Default      | Description                   |
| --------------------------------------- | ------------ | ----------------------------- |
| `ECS_CONTAINER_METADATA_URI_V4`         | (set by ECS) | Metadata endpoint URL         |
| `ECS_CONTAINER_METADATA_URI_V4_TIMEOUT` | `5s`         | Timeout for metadata requests |

## Example: ECS Task Definition

Using `ecs-task-helper` as an entrypoint wrapper:

```json
{
  "containerDefinitions": [
    {
      "name": "myapp",
      "image": "myapp:latest",
      "entryPoint": ["/ecs-task-helper", "exec"],
      "command": ["/app/myservice"],
      "healthCheck": {
        "command": ["CMD", "/ecs-task-helper", "check", "--quiet", "http://localhost:8080/health"],
        "interval": 30,
        "timeout": 5,
        "retries": 3
      }
    }
  ]
}
```

To include the binary in your image, use a multi-stage build:

```dockerfile
FROM ghcr.io/ixti/ecs-task-helper:latest AS ecs-task-helper

FROM your-base-image
COPY --from=ecs-task-helper /ecs-task-helper /ecs-task-helper
# ... rest of your Dockerfile
```

## Building

```sh
# Build binary
go build -o ecs-task-helper .

# Run tests
go test ./...

# Build Docker image
docker build -t ecs-task-helper .
```
