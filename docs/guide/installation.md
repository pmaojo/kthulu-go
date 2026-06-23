---
title: Installation
description: Get up and running with the Kthulu CLI and development environment.
---


Kthulu is a Go-based CLI tool. To get started, you'll need a few prerequisites.

## Prerequisites

- **Go 1.21+**: Kthulu is built with Go. [Download Go](https://go.dev/dl/)
- **Node.js 18+**: Required for frontend development and the Hub. [Download Node.js](https://nodejs.org/)
- **Docker** (Optional): Recommended for running local databases and services.

## Installing the CLI

You can install the Kthulu CLI directly using `go install`.

```bash
go install github.com/pmaojo/kthulu-go/cmd/kthulu@latest
```

Ensure your `$(go env GOPATH)/bin` is in your system `PATH`.

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

## Verifying Installation

Run the following command to verify that Kthulu is installed correctly:

```bash
kthulu version
```

You should see output similar to:

```text
Kthulu CLI v1.3.5
```

## Setup & Configuration

After installation, initialize your local configuration:

```bash
kthulu init
```

This will set up the `~/.kthulu` directory for plugins and local cache.

## Next Steps

Now that you have the CLI installed, take the [Quick Start Tour](/docs/guide/quick-start) to build your first application.
