---
title: "Verifactu"
description: "VerifactuModule wires VeriFactu dependencies and routes."
type: "module"
author: "Kthulu Core"
stars: 0
icon: "Box"
---

# Verifactu

VerifactuModule wires VeriFactu dependencies and routes.

## Features

- HTTP API
- Domain Logic
- Database Persistence



## Configuration

The module is configured via environment variables:

| Variable | Description |
|----------|-------------|
| `VERIFACTU_SIGN_KEY` | Configuration for VERIFACTU_SIGN_KEY |


## Installation

Add this module to your project:

```bash
kthulu add module verifactu
```

## Components

This module provides the following components to the application:

- **Backend**: Go module wired with Uber Fx.
