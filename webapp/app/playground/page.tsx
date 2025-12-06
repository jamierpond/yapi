import type { Metadata } from "next";
import Navbar from "../components/Navbar";
import Playground from "../components/Playground";
import ShareButton from "../components/ShareButton";
import { SITE_URL } from "@/app/lib/constants";

export const metadata: Metadata = {
  title: "Playground",
  description:
    "Interactive YAML API playground. Write and test HTTP, gRPC, and TCP requests directly in your browser with real-time validation.",
  alternates: {
    canonical: `${SITE_URL}/playground`,
  },
  openGraph: {
    title: "yapi Playground",
    description:
      "Interactive YAML API playground. Write and test HTTP, gRPC, and TCP requests directly in your browser.",
    url: `${SITE_URL}/playground`,
  },
  twitter: {
    title: "yapi Playground",
    description:
      "Interactive YAML API playground. Write and test HTTP, gRPC, and TCP requests directly in your browser.",
  },
};

export default function PlaygroundPage() {
  return (
    <div className="flex flex-col h-screen">
      <Navbar rightContent={<ShareButton />} />
      <Playground />
    </div>
  );
}
