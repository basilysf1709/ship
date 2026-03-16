const BASE_URL = "https://shipinfra.dev";

interface BlogPostingProps {
  headline: string;
  description: string;
  datePublished: string;
  dateModified?: string;
  slug: string;
}

export function BlogPostingJsonLd({
  headline,
  description,
  datePublished,
  dateModified,
  slug,
}: BlogPostingProps) {
  const data = {
    "@context": "https://schema.org",
    "@type": "BlogPosting",
    headline,
    description,
    datePublished,
    dateModified: dateModified ?? datePublished,
    author: {
      "@type": "Person",
      name: "Basil Yusuf",
      url: "https://github.com/basilysf1709",
    },
    publisher: {
      "@type": "Organization",
      name: "Ship",
      logo: { "@type": "ImageObject", url: `${BASE_URL}/logo.png` },
    },
    mainEntityOfPage: {
      "@type": "WebPage",
      "@id": `${BASE_URL}/blog/${slug}`,
    },
    image: `${BASE_URL}/opengraph-image`,
  };

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(data) }}
    />
  );
}
