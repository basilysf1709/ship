import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { Navbar } from "@/components/navbar";
import { Footer } from "@/components/footer";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const BASE_URL = "https://shipinfra.dev";

export const metadata: Metadata = {
  metadataBase: new URL(BASE_URL),
  title: {
    default: "Ship - Infrastructure CLI for AI Agents",
    template: "%s | Ship",
  },
  description:
    "Provision, deploy, and manage cloud servers from the terminal. A lightweight infrastructure CLI built for AI coding agents.",
  keywords: [
    "infrastructure CLI",
    "deploy from terminal",
    "AI coding agents",
    "cloud server provisioning",
    "DigitalOcean CLI",
    "Hetzner CLI",
    "Vultr CLI",
    "Docker deploy",
    "ship CLI",
    "devops",
  ],
  authors: [{ name: "Basil Yusuf", url: "https://github.com/basilysf1709" }],
  creator: "Basil Yusuf",
  openGraph: {
    type: "website",
    locale: "en_US",
    url: BASE_URL,
    siteName: "Ship",
    title: "Ship - Infrastructure CLI for AI Agents",
    description:
      "Provision, deploy, and manage cloud servers from the terminal. Built for AI coding agents.",
    images: [
      {
        url: "/og-image.png",
        width: 1200,
        height: 630,
        alt: "Ship - Infrastructure CLI for AI Agents",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Ship - Infrastructure CLI for AI Agents",
    description:
      "Provision, deploy, and manage cloud servers from the terminal. Built for AI coding agents.",
    images: ["/og-image.png"],
  },
  alternates: {
    canonical: "/",
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
};

const jsonLd = [
  {
    "@context": "https://schema.org",
    "@type": "Organization",
    name: "Ship",
    url: BASE_URL,
    logo: `${BASE_URL}/logo.png`,
    sameAs: ["https://github.com/basilysf1709/ship"],
  },
  {
    "@context": "https://schema.org",
    "@type": "WebSite",
    name: "Ship",
    url: BASE_URL,
    description:
      "A lightweight infrastructure CLI built for AI coding agents.",
    publisher: {
      "@type": "Organization",
      name: "Ship",
      logo: { "@type": "ImageObject", url: `${BASE_URL}/logo.png` },
    },
  },
];

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
        />
        <Navbar />
        <main>{children}</main>
        <Footer />
      </body>
    </html>
  );
}
