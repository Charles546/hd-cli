# hd-cli

Honeydipper CLI tool for deploying, configuring, and managing Honeydipper instances.

## Overview

`hd` is a command-line interface (CLI) tool that simplifies working with Honeydipper,
an event-driven, rule-based orchestration platform. With `hd`, you can quickly deploy
Honeydipper instances, manage configurations, check daemon status, and more.

## Features

- **Easy Deployment**: Deploy Honeydipper via Docker, source, or Kubernetes
- **Configuration Management**: View, edit, validate, and push configurations
- **Status Monitoring**: Check daemon health and status
- **Shell Completions**: Generate completions for bash, zsh, and fish

## Installation

### From Source

```bash
git clone https://github.com/Charles546/hd-cli.git
cd hd-cli
make build
```

### Install to GOPATH

```bash
make install
```

## Usage

```bash
# Initialize a new Honeydipper project
hd init

# Deploy Honeydipper
hd deploy docker
hd deploy source
hd deploy kubernetes

# Manage configuration
hd config show
hd config edit
hd config validate
hd config push
hd config diff

# Check daemon status
hd status

# Destroy instance
hd destroy

# Print version information
hd version

# Generate shell completions
hd completion bash
hd completion zsh
hd completion fish
```

### Global Flags

- `--verbose`: Enable verbose output
- `--config-dir <path>`: Specify configuration directory
- `--output <format>`: Output format (json or text)

## Development

### Prerequisites

- Go 1.25 or later
- Docker or Podman (for deployment features)
- Make

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

### Cleaning

```bash
make clean
```

## License

This project is dual-licensed. By default it is licensed under the GNU Affero General
Public License v3.0. If you have a separate written commercial agreement, you may
use it under those terms instead.

See [LICENSE](./LICENSE) for the full AGPL-3.0 license text and
[LICENSE-COMMERCIAL.md](./LICENSE-COMMERCIAL.md) for commercial licensing information.
