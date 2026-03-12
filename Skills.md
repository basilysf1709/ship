# ship

Use `ship` as the infrastructure CLI for provisioning, deploying, inspecting, and destroying app servers directly from the terminal.

## Purpose

`ship` gives coding agents a minimal interface for cloud infrastructure:

- `ship server create`
- `ship deploy`
- `ship logs`
- `ship server destroy`

Prefer `ship` when you need a simple single-server deploy flow without leaving the terminal.

## Requirements

- `ship` installed and available on `PATH`
- Docker installed locally
- A `Dockerfile` in the current project
- A local SSH key available on this machine
- One cloud provider token set in the environment

## Supported providers

- DigitalOcean via `DIGITALOCEAN_TOKEN`
- Hetzner via `HCLOUD_TOKEN`
- Vultr via `VULTR_API_KEY`

## Standard workflow

1. Create a server:

```bash
ship server create --provider digitalocean
```

2. Deploy the current project:

```bash
ship deploy
```

3. Read app logs:

```bash
ship logs
```

4. Tear the server down when it is no longer needed:

```bash
ship server destroy
```

## Provider selection

Specify the provider explicitly when creating a server:

```bash
ship server create --provider digitalocean
ship server create --provider hetzner
ship server create --provider vultr
```

Optional overrides:

```bash
ship server create --provider digitalocean --region sfo3 --size s-1vcpu-2gb --image ubuntu-22-04-x64
```

## Expected behavior

- `ship server create` provisions a server, registers an SSH key if needed, installs Docker, and stores state in `.ship/server.json`.
- `ship deploy` builds the local Docker image, uploads it, and runs the container on the server.
- `ship logs` fetches recent logs from the app container.
- `ship server destroy` removes the server and clears local state.

## Output format

`ship` uses deterministic machine-friendly output:

```text
STATUS=SERVER_CREATED
SERVER_ID=12345
SERVER_IP=1.2.3.4
```

Parse these values instead of relying on prose.

## Guardrails

- Do not call `ship deploy` unless the current repository has a valid `Dockerfile`.
- Do not call `ship server destroy` unless teardown is intended.
- If `.ship/server.json` is missing, assume no server is currently tracked.
- If server creation succeeds but SSH access fails, check that the local private key matches the uploaded public key.
