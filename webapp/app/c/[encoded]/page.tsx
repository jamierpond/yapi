import Playground from "../../components/Playground";
import { yapiDecode } from "../../_lib/yapi-encode";
import type { Metadata } from "next";

type Props = {
  params: Promise<{ encoded: string }>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  try {
    const { encoded } = await params;
    const decoded = yapiDecode(encoded);
    const preview = decoded.length > 200 ? decoded.slice(0, 200) + "..." : decoded;

    return {
      title: "yapi playground",
      description: preview,
      openGraph: {
        title: "yapi playground",
        description: preview,
        type: "website",
      },
      twitter: {
        card: "summary_large_image",
        title: "yapi playground",
        description: preview,
      },
    };
  } catch (e) {
    return {
      title: "yapi playground",
      description: "compiler explorer for APIs",
    };
  }
}

export default function Home() {
  return <Playground />;
}
