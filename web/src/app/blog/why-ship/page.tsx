import type { Metadata } from "next";
import Link from "next/link";
import { BlogPostingJsonLd } from "@/lib/json-ld";

export const metadata: Metadata = {
  title: "Why Ship Exists",
  description:
    "The problem with cloud dashboards and why we built a terminal-first infrastructure CLI.",
  alternates: { canonical: "/blog/why-ship" },
  openGraph: {
    type: "article",
    title: "Why Ship Exists",
    description:
      "The problem with cloud dashboards and why we built a terminal-first infrastructure CLI.",
    url: "/blog/why-ship",
    publishedTime: "2026-03-15T00:00:00Z",
    authors: ["Basil Yusuf"],
  },
  twitter: {
    card: "summary_large_image",
    title: "Why Ship Exists",
    description:
      "The problem with cloud dashboards and why we built a terminal-first infrastructure CLI.",
  },
};

export default function WhyShipPost() {
  return (
    <article className="mx-auto max-w-3xl px-6 pb-20 pt-16">
      <BlogPostingJsonLd
        headline="Why Ship Exists"
        description="The problem with cloud dashboards and why we built a terminal-first infrastructure CLI."
        datePublished="2026-03-15"
        slug="why-ship"
      />
      <Link
        href="/blog"
        className="text-sm text-muted transition-colors hover:text-foreground"
      >
        &larr; Back to blog
      </Link>
      <time className="mt-6 block text-sm text-muted">March 15, 2026</time>
      <h1 className="mt-2 text-4xl font-bold">Why Ship Exists</h1>

      <div className="prose mt-8 space-y-5 text-muted [&_h2]:mt-10 [&_h2]:text-xl [&_h2]:font-semibold [&_h2]:text-foreground [&_strong]:text-foreground">
        <p>
          Every developer has been there. You write your code, it works locally,
          and then you spend the next hour clicking through a cloud
          provider&apos;s dashboard trying to provision a server, configure
          firewalls, set up SSH keys, and get your app deployed.
        </p>

        <h2>The Dashboard Problem</h2>
        <p>
          Cloud dashboards are designed for humans who operate infrastructure
          full-time. They&apos;re feature-rich, deeply nested, and full of
          options you don&apos;t need for a straightforward deploy. For a
          developer — or an AI agent — that just wants to get code running on a
          server, they&apos;re friction.
        </p>
        <p>
          The context switch alone is costly. You leave your terminal, open a
          browser, navigate menus, wait for pages to load, and manually copy
          values back and forth. This breaks flow for humans and is simply
          impossible for AI agents.
        </p>

        <h2>Terminal-First</h2>
        <p>
          Ship takes the opposite approach. Everything happens in the terminal.{" "}
          <strong>One command to create a server.</strong>{" "}
          <strong>One command to deploy.</strong>{" "}
          <strong>One command to check status.</strong> No browser tabs, no
          clicking, no copy-pasting IPs.
        </p>
        <p>
          The output is machine-friendly by design — <code>KEY=VALUE</code>{" "}
          pairs that any script or agent can parse. Add <code>--json</code> when
          you need structured data.
        </p>

        <h2>Built for Agents</h2>
        <p>
          Ship was designed from the start to work with AI coding agents. When
          an agent like Claude Code or Cursor needs to deploy your project, it
          can invoke Ship commands directly. No browser automation, no screen
          scraping, no API wrangling — just clean CLI calls with deterministic
          output.
        </p>
        <p>
          The agent reads the output, understands the state, and moves on. Ship
          keeps the infrastructure layer thin so the agent can focus on what it
          does best: writing and shipping code.
        </p>

        <h2>Minimal by Design</h2>
        <p>
          Ship doesn&apos;t try to be Terraform or Pulumi. It&apos;s not an
          infrastructure-as-code framework. It&apos;s a single binary that does
          exactly what a developer or agent needs to get an app live on a
          server: provision, deploy, monitor, rollback.
        </p>
        <p>
          That&apos;s it. No state files to sync. No plan/apply cycles. No YAML
          schemas. Just commands that do what they say.
        </p>
      </div>
    </article>
  );
}
