<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/logo-dark-transparent.png" />
    <img src="assets/logo.png" alt="ship Logo" width="320" />
  </picture>
</p>

<h1 align="center">ship</h1>

<p align="center"><strong>Infrastructure for AI Coding Agents</strong></p>

<p align="center">
An extremely lightweight infrastructure CLI for provisioning, deploying, tailing logs, and destroying servers.<br/>
One binary. Zero dashboards. A minimal cloud control layer that agents can drive reliably.
</p>

<p align="center">
  <a href="#install">Install</a> &bull;
  <a href="#quick-start">Quick Start</a> &bull;
  <a href="#agent-skill">Agent Skill</a> &bull;
  <a href="https://github.com/basilysf1709/ship">GitHub</a> &bull;
  <a href="https://github.com/basilysf1709/ship/releases">Releases</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/providers-DigitalOcean%20%7C%20Hetzner%20%7C%20Vultr-blue" alt="providers" />
  <img src="https://img.shields.io/badge/license-MIT-green" alt="license" />
  <img src="https://img.shields.io/github/v/release/basilysf1709/ship?color=orange&label=version" alt="version" />
  <img src="https://img.shields.io/badge/language-Go-00ADD8" alt="language" />
</p>

## What is ship?

`ship` is a minimal infrastructure primitive for AI coding agents.

There are too many moments where an agent is working inside the terminal, then suddenly has to break context to provision a server, deploy code, fetch logs, or tear infrastructure down. That context switch is wasteful. `ship` keeps the entire flow inside the CLI so deployment can be handled directly by cloud-capable agents without leaving the terminal.

The goal is simple: give agents a tiny, deterministic interface for infrastructure operations.

**How it works:**

1. **Create** a server with `ship server create`
2. **Deploy** the current project with `ship deploy`
3. **Inspect** the running app with `ship logs`
4. **Destroy** the server with `ship server destroy`

**Key features:**

- **Minimal command surface**: only create, deploy, logs, and destroy
- **Single binary**: build once with `go build -o ship`
- **Provider support**: DigitalOcean, Hetzner, and Vultr
- **Deterministic output**: machine-friendly `KEY=VALUE` responses
- **No dashboard required**: everything happens from the terminal
- **Configurable deploy flow**: use `ship.json` for project-specific deploy steps
- **Local state tracking**: server metadata stored in `.ship/server.json`

## Build

```bash
go build -o ship
```

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/basilysf1709/ship/main/install.sh | sh
```

## Agent Skill

Download the reusable skill file directly:

```bash
curl -O https://raw.githubusercontent.com/basilysf1709/ship/main/Skills.md
```

This file is intended to be dropped into an agent workflow as a concise instruction sheet for using `ship`.

## Quick Start

```bash
export DIGITALOCEAN_TOKEN=...

ship server create --provider digitalocean
ship deploy
ship logs
ship server destroy
```

Provider selection:

```bash
ship server create --provider digitalocean
ship server create --provider hetzner
ship server create --provider vultr
```

## Authentication

`ship` uses provider tokens from environment variables:

- `DIGITALOCEAN_TOKEN`
- `HCLOUD_TOKEN`
- `VULTR_API_KEY`

You only need to set the token for the provider you are using.

## Setup

Before creating a server, make sure this machine has an SSH key available. `ship` will use a local SSH public key and automatically register it with the provider if needed.

Recommended default setup:

```bash
ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519
ssh-add ~/.ssh/id_ed25519
```

If a matching local private key is not available, server creation can succeed but SSH bootstrap and deploy steps will fail.

## Requirements

- Go
- A local SSH key on this machine, either in `~/.ssh/` or loaded in `ssh-agent`
- Docker installed locally if you use the default deploy flow
- A `Dockerfile` in the current project if you use the default deploy flow

## Usage

```bash
ship server create
ship deploy
ship logs
ship server destroy
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `ship server create` | Create a server and install Docker |
| `ship deploy` | Run the project's configured deploy flow, or the default Docker deploy if no config is present |
| `ship logs` | Fetch the last 100 log lines from the app container |
| `ship server destroy` | Destroy the current server and remove local state |

### Flags

| Flag | Description |
|------|-------------|
| `--provider <name>` | Choose provider: `digitalocean`, `hetzner`, or `vultr` |
| `--region <region>` | Override the provider region or location |
| `--size <size>` | Override the provider size, plan, or server type |
| `--image <image>` | Override the provider image |

## Providers

Default create settings by provider:

| Provider | Region | Size | Image |
|---------|--------|------|-------|
| DigitalOcean | `nyc3` | `s-2vcpu-4gb` | `ubuntu-22-04-x64` |
| Hetzner | `nbg1` | `cx22` | `ubuntu-22.04` |
| Vultr | `ewr` | `vc2-2c-4gb` | `Ubuntu 22.04 x64` |

Examples:

```bash
ship server create --provider digitalocean --region sfo3 --size s-1vcpu-2gb --image ubuntu-22-04-x64
ship server create --provider hetzner --region fsn1 --size cpx21 --image ubuntu-24.04
ship server create --provider vultr --region ord --size vc2-1c-2gb --image "Ubuntu 24.04 x64"
```

## Deploy Flow

`ship deploy` looks for a `ship.json` file in the current directory. If it finds a `deploy` block, it runs that deploy recipe. If `ship.json` is missing, it falls back to the default Docker-based flow.

### Configurable deploys

Example `ship.json`:

```json
{
  "deploy": {
    "local_commands": [
      "npm ci",
      "npm run build",
      "tar -czf release.tar.gz dist package.json"
    ],
    "uploads": [
      {
        "source": "release.tar.gz",
        "destination": "/opt/app/release.tar.gz",
        "mode": "0644"
      }
    ],
    "remote_commands": [
      "mkdir -p /opt/app",
      "cd /opt/app && tar -xzf release.tar.gz",
      "cd /opt/app && npm ci --omit=dev",
      "cd /opt/app && pm2 restart app || pm2 start npm --name app -- start"
    ],
    "cleanup_local": [
      "release.tar.gz"
    ]
  }
}
```

Fields:

- `local_commands`: shell commands run on the local machine before upload
- `uploads`: files to copy to the server, with `source`, `destination`, and optional quoted octal `mode`
- `remote_commands`: shell commands run on the server in order
- `cleanup_local`: local files removed after deploy finishes

### Default deploy flow

If no `ship.json` deploy config exists, `ship deploy` assumes a `Dockerfile` exists and runs:

```bash
docker build -t app .
docker save app -o app.tar
```

Then it uploads the image to the server and runs:

```bash
docker load -i /root/app.tar
docker stop app || true
docker rm app || true
docker run -d --name app -p 80:80 app
```

Example output:

```text
STATUS=DEPLOY_COMPLETE
SERVER_IP=1.2.3.4
```

## Server State

Server metadata is stored locally in:

```text
.ship/server.json
```

Example:

```json
{
  "provider": "digitalocean",
  "server_id": "12345",
  "ip": "1.2.3.4",
  "ssh_user": "root"
}
```

## Output Format

All command output is designed to stay clean and machine-parseable:

```text
STATUS=SERVER_CREATED
SERVER_ID=12345
SERVER_IP=1.2.3.4
```

## Why it exists

`ship` is for the gap between coding and infrastructure.

When an AI coding agent is already operating in a terminal, it should not need a browser, a dashboard, or a separate deployment toolchain just to ship code. Provisioning, deployment, log access, and teardown should all be available as a small CLI primitive that agents can call directly and predictably.
