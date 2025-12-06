import { NextResponse } from "next/server";

// GraphQL response types
type Asset = {
  name: string;
  downloadCount: number;
};

type Release = {
  name: string;
  releaseAssets: {
    nodes: Asset[];
  };
};

type GraphQLResponse = {
  data: {
    repository: {
      releases: {
        totalCount: number;
        nodes: Release[];
      };
    };
  };
};

const query = `
  query {
    repository(owner: "jamierpond", name: "yapi") {
      releases(first: 100, orderBy: {field: CREATED_AT, direction: DESC}) {
        totalCount
        nodes {
          name
          releaseAssets(first: 100) {
            nodes {
              name
              downloadCount
            }
          }
        }
      }
    }
  }
`;

export async function GET() {
  try {
    if (!process.env.GITHUB_PAT) {
      return NextResponse.json(
        { error: "Server misconfiguration" },
        { status: 500 }
      );
    }

    const res = await fetch("https://api.github.com/graphql", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${process.env.GITHUB_PAT}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ query }),
      next: { revalidate: 3600 },
    });

    if (!res.ok) {
      throw new Error("Failed to fetch releases");
    }

    const { data }: GraphQLResponse = await res.json();

    if (!data?.repository) {
      return NextResponse.json({ total_downloads: 0, total_releases: 0 });
    }

    const releases = data.repository.releases;
    let totalDownloads = 0;

    releases.nodes.forEach((release) => {
      release.releaseAssets.nodes.forEach((asset) => {
        if (asset.name === "checksums.txt") {
          return;
        }
        const realCount = Math.max(0, asset.downloadCount - 1);
        totalDownloads += realCount;
      });
    });

    return NextResponse.json({
      total_downloads: totalDownloads,
      total_releases: releases.totalCount,
    });
  } catch (error) {
    return NextResponse.json(
      { error: "Failed to calculate downloads" },
      { status: 500 }
    );
  }
}
