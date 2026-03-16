import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Blog",
  description:
    "Articles about Ship, infrastructure automation, and AI-driven deployment.",
  alternates: { canonical: "/blog" },
  openGraph: {
    title: "Ship Blog",
    description:
      "Articles about Ship, infrastructure automation, and AI-driven deployment.",
    url: "/blog",
  },
};

const posts = [
  {
    slug: "why-ship",
    title: "Why Ship Exists",
    description:
      "The problem with cloud dashboards and why we built a terminal-first infrastructure CLI.",
    date: "2026-03-15",
  },
  {
    slug: "deploy-in-60-seconds",
    title: "Deploy an App in 60 Seconds",
    description:
      "A walkthrough of going from zero to a live server with a single CLI tool.",
    date: "2026-03-14",
  },
  {
    slug: "ai-agents-need-infra",
    title: "AI Agents Need Infrastructure Tools",
    description:
      "Why AI coding agents struggle with deployment and how Ship solves it.",
    date: "2026-03-13",
  },
];

export default function BlogPage() {
  return (
    <div className="mx-auto max-w-3xl px-6 pb-20 pt-16">
      <h1 className="text-4xl font-bold">Blog</h1>
      <p className="mt-3 text-muted">
        Updates, guides, and thoughts on terminal-first infrastructure.
      </p>

      <div className="mt-10 space-y-8">
        {posts.map((post) => (
          <Link
            key={post.slug}
            href={`/blog/${post.slug}`}
            className="block rounded-lg border border-border bg-card p-6 transition-colors hover:border-accent/40"
          >
            <time className="text-xs text-muted">{post.date}</time>
            <h2 className="mt-1 text-xl font-semibold">{post.title}</h2>
            <p className="mt-2 text-sm text-muted">{post.description}</p>
          </Link>
        ))}
      </div>
    </div>
  );
}
