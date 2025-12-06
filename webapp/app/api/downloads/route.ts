import { NextResponse } from "next/server";

// Define the shape of your data based on the GitHub REST API response
type Asset = {
  name: string;
  download_count: number;
};

type Release = {
  name: string;
  assets: Asset[];
};

export async function GET() {
  try {
    // 1. Fetch the data
    // Replace this URL with your actual data source (e.g., GitHub API)
    const res = await fetch("https://api.github.com/repos/jamierpond/yapi/releases", {
      // next: { revalidate: 3600 }, // Cache for 1 hour
    });

    if (!res.ok) {
      throw new Error("Failed to fetch releases");
    }

    const releases: Release[] = await res.json();

    let totalDownloads = 0;

    // 2. Iterate and Calculate
    releases.forEach((release) => {
      release.assets.forEach((asset) => {
        // Filter: Ignore checksum files, we only want binary installs
        if (asset.name === "checksums.txt") {
          return;
        }

        // Logic: The API inflates counts by 1.
        // We subtract 1, ensuring we don't go below zero.
        const realCount = Math.max(0, asset.download_count - 1);

        totalDownloads += realCount;
      });
    });

    // 3. Return the clean number
    return NextResponse.json({
      total_real_downloads: totalDownloads,
      meta: {
        note: "Excludes checksums.txt and adjusts for API inflation (-1 per asset).",
      },
    });

  } catch (error) {
    return NextResponse.json(
      { error: "Failed to calculate downloads" },
      { status: 500 }
    );
  }
}
