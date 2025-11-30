import "monaco-editor/min/vs/editor/editor.main.css";
import type { Metadata } from "next";
import { JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { SITE_TITLE, SITE_DESCRIPTION, SITE_URL } from "@/app/lib/constants";

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
    "Bash",
    "API testing",
    "REST API",
    "command line",
    "CLI tool",
    "API workbench",
    "developer tools",
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
    description: "Bash-powered YAML API workbench for HTTP, gRPC, and TCP",
  },
  twitter: {
    card: "summary_large_image",
    title: `${SITE_TITLE} - YAML API Client`,
    description: "Bash-powered YAML API workbench for HTTP, gRPC, and TCP",
    creator: "@jamierpond",
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
      <body
        className={`${jetbrainsMono.variable} antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
