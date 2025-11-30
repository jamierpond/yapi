import "monaco-editor/min/vs/editor/editor.main.css";
import type { Metadata } from "next";
import { JetBrains_Mono } from "next/font/google";
import "./globals.css";

const jetbrainsMono = JetBrains_Mono({
  variable: "--font-jetbrains-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "yapi - YAML API Client for HTTP, gRPC & TCP",
  description: "A small, Bash-powered YAML API client that speaks HTTP, gRPC, and raw TCP. Write clean YAML configs, version control your requests, execute from CLI or web playground.",
  keywords: ["API client", "YAML", "HTTP client", "gRPC client", "TCP client", "Bash", "API testing", "REST API", "command line", "CLI tool"],
  authors: [{ name: "yapi" }],
  openGraph: {
    title: "yapi - YAML API Client",
    description: "Bash-powered YAML API workbench for HTTP, gRPC, and TCP",
    type: "website",
    locale: "en_US",
  },
  twitter: {
    card: "summary_large_image",
    title: "yapi - YAML API Client",
    description: "Bash-powered YAML API workbench for HTTP, gRPC, and TCP",
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
