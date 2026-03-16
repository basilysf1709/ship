import Image from "next/image";
import Link from "next/link";
import { CodeBlock } from "@/components/code-block";
import { Step } from "@/components/step";

const softwareJsonLd = {
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  name: "Ship",
  description:
    "A lightweight infrastructure CLI for provisioning, deploying, and managing cloud servers. Built for AI coding agents.",
  url: "https://shipinfra.dev",
  applicationCategory: "DeveloperApplication",
  operatingSystem: "macOS, Linux",
  offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
  author: {
    "@type": "Person",
    name: "Basil Yusuf",
    url: "https://github.com/basilysf1709",
  },
  codeRepository: "https://github.com/basilysf1709/ship",
  license: "https://opensource.org/licenses/MIT",
};

export default function Home() {
  return (
    <div className="mx-auto max-w-3xl px-6">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(softwareJsonLd) }}
      />
      {/* Hero */}
      <section className="pb-16 pt-24 text-center">
        <Image
          src="/logo.png"
          alt="Ship"
          width={220}
          height={220}
          className="mx-auto mb-6"
          priority
        />
        <h1 className="text-5xl font-bold tracking-tight">Ship</h1>
        <p className="mt-4 text-xl text-muted">
          Deploy from terminal. No dashboards. No context-switching.
        </p>
        <p className="mt-3 text-sm text-muted">
          A lightweight infrastructure CLI built for AI coding agents.
          <br />
          Provision servers, deploy apps, manage secrets — all from one command.
        </p>
      </section>

      {/* Install */}
      <section className="pb-16">
        <h2 className="mb-4 text-2xl font-bold">Install</h2>
        <CodeBlock
          code="curl -fsSL https://raw.githubusercontent.com/basilysf1709/ship/main/install.sh | sh"
          language="bash"
        />
        <p className="mt-3 text-sm text-muted">
          Downloads the latest binary for your platform. Requires macOS or
          Linux.
        </p>
      </section>

      {/* Quick Start */}
      <section className="pb-16">
        <h2 className="mb-6 text-2xl font-bold">Quick Start</h2>
        <div className="space-y-8">
          <Step number={1} title="Set your provider token">
            Export your cloud provider API key. Ship supports DigitalOcean,
            Hetzner, and Vultr.
            <div className="mt-3">
              <CodeBlock code="export DIGITALOCEAN_TOKEN=dop_v1_..." />
            </div>
          </Step>
          <Step number={2} title="Create a server">
            Provisions an Ubuntu server with Docker pre-installed and your SSH
            key registered.
            <div className="mt-3">
              <CodeBlock code="ship server create --provider digitalocean" />
            </div>
          </Step>
          <Step number={3} title="Deploy your app">
            Builds your Docker image, uploads it, and runs the container on your
            server.
            <div className="mt-3">
              <CodeBlock code="ship deploy" />
            </div>
          </Step>
          <Step number={4} title="Check status">
            Verify your app is live with SSH reachability, healthcheck, and
            release info.
            <div className="mt-3">
              <CodeBlock code="ship status" />
            </div>
          </Step>
        </div>
      </section>

      {/* Providers */}
      <section className="pb-16">
        <h2 className="mb-6 text-2xl font-bold">Supported Providers</h2>
        <div className="grid gap-4 sm:grid-cols-3">
          {[
            {
              name: "DigitalOcean",
              env: "DIGITALOCEAN_TOKEN",
              region: "nyc3",
            },
            { name: "Hetzner", env: "HCLOUD_TOKEN", region: "nbg1" },
            { name: "Vultr", env: "VULTR_API_KEY", region: "ewr" },
          ].map((p) => (
            <div
              key={p.name}
              className="rounded-lg border border-border bg-card p-4"
            >
              <h3 className="font-semibold">{p.name}</h3>
              <p className="mt-1 text-xs text-muted">
                <code className="rounded bg-background px-1 py-0.5">
                  {p.env}
                </code>
              </p>
              <p className="mt-1 text-xs text-muted">
                Default region: {p.region}
              </p>
            </div>
          ))}
        </div>
      </section>

      {/* Commands */}
      <section className="pb-16">
        <h2 className="mb-6 text-2xl font-bold">Commands</h2>
        <div className="overflow-hidden rounded-lg border border-border">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border bg-card">
                <th className="px-4 py-3 text-left font-medium text-muted">
                  Command
                </th>
                <th className="px-4 py-3 text-left font-medium text-muted">
                  Description
                </th>
              </tr>
            </thead>
            <tbody>
              {[
                ["ship server create", "Provision a new cloud server", "server-create"],
                ["ship server list", "Show all tracked servers", "server-list"],
                ["ship server destroy", "Tear down server and clean state", "server-destroy"],
                ["ship deploy", "Build and deploy your application", "deploy"],
                ["ship status", "Check server and app health", "status"],
                ["ship logs", "Fetch container logs", "logs"],
                ["ship exec", "Run remote commands on server", "exec"],
                ["ship rollback", "Restore a previous release", "rollback"],
                ["ship secrets", "Manage and sync environment secrets", "secrets"],
                ["ship domain setup", "Configure reverse proxy with auto-TLS", "domain-setup"],
                ["ship init", "Generate starter ship.json config", "init"],
                ["ship bootstrap", "Apply packages and setup commands", "bootstrap"],
              ].map(([cmd, desc, anchor]) => (
                <tr key={cmd} className="border-b border-border last:border-0">
                  <td className="px-4 py-2.5">
                    <Link href={`/docs#${anchor}`} className="hover:underline">
                      <code className="text-accent">{cmd}</code>
                    </Link>
                  </td>
                  <td className="px-4 py-2.5 text-muted">{desc}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      {/* Configuration */}
      <section className="pb-20">
        <h2 className="mb-4 text-2xl font-bold">Configuration</h2>
        <p className="mb-4 text-sm text-muted">
          Optionally create a{" "}
          <code className="rounded bg-card px-1.5 py-0.5 text-accent">
            ship.json
          </code>{" "}
          for custom deploy flows.
        </p>
        <CodeBlock
          language="json"
          code={`{
  "deploy": {
    "local_commands": ["npm ci", "npm run build"],
    "uploads": [{
      "source": "dist.tar.gz",
      "destination": "/opt/app/release.tar.gz"
    }],
    "remote_commands": ["cd /opt/app && tar -xzf release.tar.gz && npm start"]
  },
  "proxy": {
    "domains": ["example.com"],
    "app_port": 3000
  }
}`}
        />
      </section>
    </div>
  );
}
