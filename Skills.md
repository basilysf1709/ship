# ship

Use `ship` as the infrastructure CLI for provisioning, deploying, inspecting, and destroying app servers directly from the terminal.

## Purpose

`ship` gives coding agents a minimal interface for cloud infrastructure:

- `ship server create`
- `ship deploy`
- `ship status`
- `ship exec`
- `ship logs`
- `ship secrets`
- `ship release list`
- `ship rollback`
- `ship bootstrap`
- `ship domain setup`
- `ship init`
- `ship server list`
- `ship server destroy`

Prefer `ship` when you need a simple single-server deploy flow without leaving the terminal.

## Requirements

- `ship` installed and available on `PATH`
- A local SSH key available on this machine
- One cloud provider token set in the environment
- Docker installed locally if the repo uses the default deploy flow
- A `Dockerfile` in the current project if the repo uses the default deploy flow

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
- `ship status` reports SSH reachability, app status, healthcheck status, and the latest tracked release.
- `ship exec` runs a one-off remote command on the tracked server.
- `ship secrets` manages `.ship/secrets.env` locally and can sync it to `/root/.ship/secrets.env` on the server.
- `ship release list` shows tracked local release history, and `ship rollback` restores a previous release.
- `ship bootstrap` applies package installs, proxy config, and custom remote commands from `ship.json`.
- `ship domain setup` configures Caddy with automatic TLS for configured domains.
- `ship init` writes a starter `ship.json` template for common app shapes.
- `ship deploy` follows the repo's `ship.json` deploy recipe when present; otherwise it uses the default Docker deploy flow.
- `ship server list` shows the locally tracked server inventory for the current machine.
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

- Do not call `ship deploy` unless the current repository has either a valid `ship.json` deploy config or a valid `Dockerfile`.
- Do not call `ship server destroy` unless teardown is intended.
- If `.ship/server.json` is missing, assume no server is currently tracked.
- If server creation succeeds but SSH access fails, check that the local private key matches the uploaded public key.
