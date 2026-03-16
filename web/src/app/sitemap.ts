import type { MetadataRoute } from "next";

export default function sitemap(): MetadataRoute.Sitemap {
  const base = "https://shipinfra.dev";

  return [
    {
      url: base,
      lastModified: new Date("2026-03-16"),
      changeFrequency: "weekly",
      priority: 1.0,
    },
    {
      url: `${base}/docs`,
      lastModified: new Date("2026-03-16"),
      changeFrequency: "weekly",
      priority: 0.9,
    },
    {
      url: `${base}/blog`,
      lastModified: new Date("2026-03-15"),
      changeFrequency: "weekly",
      priority: 0.7,
    },
    {
      url: `${base}/blog/why-ship`,
      lastModified: new Date("2026-03-15"),
      changeFrequency: "monthly",
      priority: 0.6,
    },
    {
      url: `${base}/blog/deploy-in-60-seconds`,
      lastModified: new Date("2026-03-14"),
      changeFrequency: "monthly",
      priority: 0.6,
    },
    {
      url: `${base}/blog/ai-agents-need-infra`,
      lastModified: new Date("2026-03-13"),
      changeFrequency: "monthly",
      priority: 0.6,
    },
  ];
}
