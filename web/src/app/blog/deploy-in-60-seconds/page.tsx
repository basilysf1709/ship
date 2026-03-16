import type { Metadata } from "next";
import Link from "next/link";
import { CodeBlock } from "@/components/code-block";
import { BlogPostingJsonLd } from "@/lib/json-ld";

export const metadata: Metadata = {
  title: "Deploy an App in 60 Seconds",
  description:
    "A walkthrough of going from zero to a live server with a single CLI tool.",
  alternates: { canonical: "/blog/deploy-in-60-seconds" },
  openGraph: {
    type: "article",
    title: "Deploy an App in 60 Seconds",
    description:
      "A walkthrough of going from zero to a live server with a single CLI tool.",
    url: "/blog/deploy-in-60-seconds",
    publishedTime: "2026-03-14T00:00:00Z",
    authors: ["Basil Yusuf"],
  },
  twitter: {
    card: "summary_large_image",
    title: "Deploy an App in 60 Seconds",
    description:
      "A walkthrough of going from zero to a live server with a single CLI tool.",
  },
};

export default function DeployPost() {
  return (
    <article className="mx-auto max-w-3xl px-6 pb-20 pt-16">
      <BlogPostingJsonLd
        headline="Deploy an App in 60 Seconds"
        description="A walkthrough of going from zero to a live server with a single CLI tool."
        datePublished="2026-03-14"
        slug="deploy-in-60-seconds"
      />
      <Link
        href="/blog"
        className="text-sm text-muted transition-colors hover:text-foreground"
      >
        &larr; Back to blog
      </Link>
      <time className="mt-6 block text-sm text-muted">March 14, 2026</time>
      <h1 className="mt-2 text-4xl font-bold">Deploy an App in 60 Seconds</h1>

      <div className="prose mt-8 space-y-5 text-muted [&_h2]:mt-10 [&_h2]:text-xl [&_h2]:font-semibold [&_h2]:text-foreground [&_strong]:text-foreground">
        <p>
          Let&apos;s go from nothing to a live app on a cloud server. All you
          need is a project with a Dockerfile and a cloud provider token.
        </p>

        <h2>Step 1: Install Ship</h2>
        <CodeBlock
          language="bash"
          code="curl -fsSL https://raw.githubusercontent.com/basilysf1709/ship/main/install.sh | sh"
        />

        <h2>Step 2: Export Your Token</h2>
        <p>
          Pick your provider and export the API token. Here we&apos;ll use
          DigitalOcean:
        </p>
        <CodeBlock
          language="bash"
          code="export DIGITALOCEAN_TOKEN=dop_v1_your_token_here"
        />

        <h2>Step 3: Create a Server</h2>
        <p>
          This provisions an Ubuntu 22.04 droplet with Docker installed, your
          SSH key registered, and all state saved locally.
        </p>
        <CodeBlock
          language="bash"
          code={`$ ship server create --provider digitalocean
STATUS=SERVER_CREATED
SERVER_ID=412847592
SERVER_IP=143.198.42.17`}
        />

        <h2>Step 4: Deploy</h2>
        <p>
          Ship builds your Docker image, uploads it over SSH, and starts the
          container.
        </p>
        <CodeBlock
          language="bash"
          code={`$ ship deploy
STATUS=DEPLOY_SUCCESS
CONTAINER_ID=a3f8b2c1d4e5`}
        />

        <h2>Step 5: Verify</h2>
        <p>
          Check that everything is running:
        </p>
        <CodeBlock
          language="bash"
          code={`$ ship status
SSH_REACHABLE=true
APP_STATUS=running
HEALTHCHECK=ok
LAST_RELEASE=2026-03-14T10:30:00Z`}
        />

        <h2>That&apos;s It</h2>
        <p>
          Five commands, under 60 seconds of active work. Your app is live on a
          real server with a real IP. From here you can set up a domain with{" "}
          <strong>ship domain setup</strong>, manage secrets with{" "}
          <strong>ship secrets</strong>, and check logs with{" "}
          <strong>ship logs</strong>.
        </p>
        <p>
          When you&apos;re done, <strong>ship server destroy</strong> tears it
          all down cleanly.
        </p>
      </div>
    </article>
  );
}
