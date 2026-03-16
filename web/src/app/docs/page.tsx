import type { Metadata } from "next";
import { CodeBlock } from "@/components/code-block";

export const metadata: Metadata = {
  title: "Documentation",
  description:
    "Complete command reference and configuration guide for the Ship infrastructure CLI.",
  alternates: { canonical: "/docs" },
  openGraph: {
    title: "Ship Documentation",
    description:
      "Complete command reference and configuration guide for the Ship infrastructure CLI.",
    url: "/docs",
  },
};

export default function DocsPage() {
  return (
    <div className="mx-auto max-w-3xl px-6 pb-20 pt-16">
      <h1 className="text-4xl font-bold">Documentation</h1>
      <p className="mt-3 text-muted">
        Everything you need to deploy with Ship.
      </p>

      <section className="mt-12">
        <h2 className="text-2xl font-bold">Prerequisites</h2>
        <ul className="mt-4 list-inside list-disc space-y-2 text-sm text-muted">
          <li>macOS or Linux</li>
          <li>
            SSH key (ed25519 recommended):{" "}
            <code className="rounded bg-card px-1.5 py-0.5">
              ssh-keygen -t ed25519
            </code>
          </li>
          <li>Docker (for default deploy flow)</li>
          <li>A Dockerfile or ship.json in your project root</li>
          <li>Cloud provider API token</li>
        </ul>
      </section>

      {/* --- Command reference sections --- */}

      <section id="server-create" className="mt-16 scroll-mt-20">
        <h2 className="text-2xl font-bold">ship server create</h2>
        <p className="mt-2 text-sm text-muted">
          Provision a new Ubuntu 22.04 server with Docker pre-installed. Your
          SSH key is automatically registered with the provider. State is saved
          to <code className="rounded bg-card px-1.5 py-0.5">.ship/server.json</code> and
          the machine-wide inventory at <code className="rounded bg-card px-1.5 py-0.5">~/.ship/servers.json</code>.
        </p>
        <div className="mt-3">
          <CodeBlock
            language="bash"
            code={`# Use defaults
ship server create --provider digitalocean

# Custom options
ship server create --provider hetzner --region fsn1 --size cx22 --image ubuntu-24-04-x64`}
          />
        </div>
        <h3 className="mt-4 text-sm font-semibold">Flags</h3>
        <ul className="mt-2 list-inside list-disc space-y-1 text-sm text-muted">
          <li><code className="rounded bg-card px-1 py-0.5">--provider</code> — digitalocean, hetzner, or vultr (required)</li>
          <li><code className="rounded bg-card px-1 py-0.5">--region</code> — override default region</li>
          <li><code className="rounded bg-card px-1 py-0.5">--size</code> — override default instance size</li>
          <li><code className="rounded bg-card px-1 py-0.5">--image</code> — override default OS image</li>
        </ul>
      </section>

      <section id="server-list" className="mt-16 scroll-mt-20">
        <h2 className="text-2xl font-bold">ship server list</h2>
        <p className="mt-2 text-sm text-muted">
          Show all servers tracked in your machine-wide inventory at{" "}
          <code className="rounded bg-card px-1.5 py-0.5">~/.ship/servers.json</code>.
          Displays provider, IP, region, and project path for each server.
        </p>
        <div className="mt-3">
          <CodeBlock code="ship server list" language="bash" />
        </div>
      </section>

      <section id="server-destroy" className="mt-16 scroll-mt-20">
        <h2 className="text-2xl font-bold">ship server destroy</h2>
        <p className="mt-2 text-sm text-muted">
          Tears down the server at the cloud provider and removes all local
          state files (<code className="rounded bg-card px-1.5 py-0.5">.ship/server.json</code>,
          releases, runtime config). Also removes the entry from the machine-wide inventory.
        </p>
        <div className="mt-3">
          <CodeBlock code="ship server destroy" language="bash" />
        </div>
      </section>

      <section id="deploy" className="mt-16 scroll-mt-20">
        <h2 className="text-2xl font-bold">ship deploy</h2>
        <p className="mt-2 text-sm text-muted">
          Build and deploy your application to the provisioned server. If a{" "}
          <code className="rounded bg-card px-1.5 py-0.5">ship.json</code> deploy config
          exists, Ship runs your custom local commands, uploads files, and
          executes remote commands. Otherwise it uses the default Docker flow:
          build image locally, upload via SSH, and run the container.
        </p>
        <div className="mt-3">
          <CodeBlock code="ship deploy" language="bash" />
        </div>
        <h3 className="mt-6 text-sm font-semibold">Custom Deploy Flow</h3>
        <p className="mt-2 text-sm text-muted">
          Define a <code className="rounded bg-card px-1.5 py-0.5">ship.json</code> with
          local build commands, file uploads, and remote commands.
        </p>
        <div className="mt-3">
          <CodeBlock
            language="json"
            code={`{
  "deploy": {
    "local_commands": ["npm ci", "npm run build"],
    "uploads": [{
      "source": "dist.tar.gz",
      "destination": "/opt/app/release.tar.gz",
      "mode": "0644"
    }],
    "remote_commands": [
      "cd /opt/app && tar -xzf release.tar.gz",
      "pm2 restart app"
    ],
    "cleanup_local": ["dist.tar.gz"]
  }
}`}
          />
        </div>
      </section>

      <section id="status" className="mt-16 scroll-mt-20">
        <h2 className="text-2xl font-bold">ship status</h2>
        <p className="mt-2 text-sm text-muted">
          Report the health of your server and application. Checks SSH
          reachability, app container status, healthcheck endpoint result, and
          last release timestamp. Configure the healthcheck path in ship.json.
        </p>
        <div className="mt-3">
          <CodeBlock
            language="bash"
            code={`$ ship status
SSH_REACHABLE=true
APP_STATUS=running
HEALTHCHECK=ok
LAST_RELEASE=2026-03-14T10:30:00Z`}
          />
        </div>
        <h3 className="mt-4 text-sm font-semibold">Configuration</h3>
        <div className="mt-2">
          <CodeBlock
            language="json"
            code={`{
  "status": {
    "healthcheck_path": "/"
  }
}`}
          />
        </div>
      </section>

      <section id="logs" className="mt-16 scroll-mt-20">
        <h2 className="text-2xl font-bold">ship logs</h2>
        <p className="mt-2 text-sm text-muted">
          Fetch the last 100 lines of logs from the application container on
          your server. Useful for quick debugging without SSHing in manually.
        </p>
        <div className="mt-3">
          <CodeBlock code="ship logs" language="bash" />
        </div>
      </section>

      <section id="exec" className="mt-16 scroll-mt-20">
        <h2 className="text-2xl font-bold">ship exec</h2>
        <p className="mt-2 text-sm text-muted">
          Run an arbitrary command on your remote server over SSH. The command
          runs as the configured SSH user and returns stdout/stderr.
        </p>
        <div className="mt-3">
          <CodeBlock
            language="bash"
            code={`ship exec "docker ps"
ship exec "df -h"
ship exec "cat /var/log/syslog | tail -50"`}
          />
        </div>
      </section>

      <section id="rollback" className="mt-16 scroll-mt-20">
        <h2 className="text-2xl font-bold">ship rollback</h2>
        <p className="mt-2 text-sm text-muted">
          Restore a previous release. Ship tracks every deploy in{" "}
          <code className="rounded bg-card px-1.5 py-0.5">.ship/releases.json</code>.
          Rollback re-uploads the previous release artifact and re-runs the
          remote commands from that release.
        </p>
        <div className="mt-3">
          <CodeBlock
            language="bash"
            code={`# List available releases
ship release list

# Roll back to the previous release
ship rollback`}
          />
        </div>
      </section>

      <section id="secrets" className="mt-16 scroll-mt-20">
        <h2 className="text-2xl font-bold">ship secrets</h2>
        <p className="mt-2 text-sm text-muted">
          Manage environment secrets locally in{" "}
          <code className="rounded bg-card px-1.5 py-0.5">.ship/secrets.env</code>{" "}
          (mode 0600) and sync them to your server.
        </p>
        <div className="mt-3">
          <CodeBlock
            language="bash"
            code={`# Add a secret
ship secrets set DATABASE_URL=postgres://...

# List secrets
ship secrets list

# Sync to server
ship secrets sync`}
          />
        </div>
      </section>

      <section id="domain-setup" className="mt-16 scroll-mt-20">
        <h2 className="text-2xl font-bold">ship domain setup</h2>
        <p className="mt-2 text-sm text-muted">
          Configure a Caddy reverse proxy with automatic TLS certificates. Point
          your DNS to the server IP first, then run the command.
        </p>
        <div className="mt-3">
          <CodeBlock
            language="bash"
            code="ship domain setup --domain example.com --port 3000"
          />
        </div>
        <p className="mt-3 text-sm text-muted">
          Or configure via ship.json:
        </p>
        <div className="mt-3">
          <CodeBlock
            language="json"
            code={`{
  "proxy": {
    "domains": ["example.com", "www.example.com"],
    "app_port": 3000
  }
}`}
          />
        </div>
      </section>

      <section id="init" className="mt-16 scroll-mt-20">
        <h2 className="text-2xl font-bold">ship init</h2>
        <p className="mt-2 text-sm text-muted">
          Generate a starter <code className="rounded bg-card px-1.5 py-0.5">ship.json</code>{" "}
          configuration file. Choose from built-in templates for common project types.
        </p>
        <div className="mt-3">
          <CodeBlock
            language="bash"
            code={`ship init --template docker
ship init --template node
ship init --template go
ship init --template static`}
          />
        </div>
      </section>

      <section id="bootstrap" className="mt-16 scroll-mt-20">
        <h2 className="text-2xl font-bold">ship bootstrap</h2>
        <p className="mt-2 text-sm text-muted">
          Apply packages, proxy configuration, and setup commands defined in
          your ship.json to the server. Runs automatically on{" "}
          <code className="rounded bg-card px-1.5 py-0.5">ship server create</code> if
          a bootstrap config exists, or can be run manually.
        </p>
        <div className="mt-3">
          <CodeBlock code="ship bootstrap" language="bash" />
        </div>
        <h3 className="mt-4 text-sm font-semibold">Configuration</h3>
        <div className="mt-2">
          <CodeBlock
            language="json"
            code={`{
  "bootstrap": {
    "packages": ["nodejs", "pm2"],
    "commands": ["npm install -g pm2"]
  }
}`}
          />
        </div>
      </section>

      {/* --- General reference --- */}

      <section className="mt-16">
        <h2 className="text-2xl font-bold">Output Format</h2>
        <p className="mt-2 text-sm text-muted">
          Ship outputs machine-friendly{" "}
          <code className="rounded bg-card px-1.5 py-0.5">KEY=VALUE</code>{" "}
          pairs by default, ideal for AI agents. Use{" "}
          <code className="rounded bg-card px-1.5 py-0.5">--json</code> for
          structured JSON output.
        </p>
        <div className="mt-3">
          <CodeBlock
            code={`$ ship server create --provider digitalocean
STATUS=SERVER_CREATED
SERVER_ID=12345
SERVER_IP=1.2.3.4

$ ship server create --provider digitalocean --json
{"status":"SERVER_CREATED","server_id":"12345","server_ip":"1.2.3.4"}`}
          />
        </div>
      </section>
    </div>
  );
}
