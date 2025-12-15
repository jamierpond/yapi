import "monaco-editor/min/vs/editor/editor.main.css";
import type { Metadata } from "next";
import { JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { SITE_TITLE, SITE_DESCRIPTION, SITE_URL, OG_BASE_URL } from "@/app/lib/constants";
import { GoogleAnalytics } from "@next/third-parties/google";

const jsonLd = {
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  name: SITE_TITLE,
  description: SITE_DESCRIPTION,
  url: SITE_URL,
  applicationCategory: "DeveloperApplication",
  operatingSystem: "macOS, Linux, Windows",
  offers: {
    "@type": "Offer",
    price: "0",
    priceCurrency: "USD",
  },
  featureList: [
    "HTTP API client",
    "gRPC client",
    "TCP client",
    "YAML configuration",
    "Offline-first",
    "Version control friendly",
    "CLI and web playground",
  ],
};

const jetbrainsMono = JetBrains_Mono({
  variable: "--font-jetbrains-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  metadataBase: new URL(SITE_URL),
  title: {
    default: `${SITE_TITLE} - YAML API Client for HTTP, gRPC & TCP`,
    template: `%s | ${SITE_TITLE}`,
  },
  description: SITE_DESCRIPTION,
  keywords: [
    "API client",
    "YAML",
    "HTTP client",
    "gRPC client",
    "TCP client",
    "Go",
    "API testing",
    "REST API",
    "command line",
    "CLI tool",
    "API workbench",
    "developer tools",
    "offline-first",
  ],
  authors: [{ name: "yapi", url: SITE_URL }],
  creator: "yapi",
  publisher: "yapi",
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
  openGraph: {
    type: "website",
    locale: "en_US",
    url: SITE_URL,
    siteName: SITE_TITLE,
    title: `${SITE_TITLE} - YAML API Client`,
    description: "Offline-first YAML API client for HTTP, gRPC, and TCP",
    images: [`${OG_BASE_URL}/og`],
  },
  twitter: {
    card: "summary_large_image",
    title: `${SITE_TITLE} - YAML API Client`,
    description: "Offline-first YAML API client for HTTP, gRPC, and TCP",
    creator: "@jamierpond",
    images: [`${OG_BASE_URL}/og`],
  },
  alternates: {
    canonical: SITE_URL,
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <GoogleAnalytics gaId="G-RFYSX5CB3L" />
      <body
        className={`${jetbrainsMono.variable} antialiased`}
      >
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
        />
        {children}
      </body>
    </html>
  );
}
