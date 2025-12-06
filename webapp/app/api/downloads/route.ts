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
        pageInfo: {
          hasNextPage: boolean;
          endCursor: string | null;
        };
        nodes: Release[];
      };
    };
  };
};

const buildQuery = (cursor?: string) => `
  query {
    repository(owner: "jamierpond", name: "yapi") {
      releases(first: 100, orderBy: {field: CREATED_AT, direction: DESC}${cursor ? `, after: "${cursor}"` : ""}) {
        totalCount
        pageInfo {
          hasNextPage
          endCursor
        }
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

async function fetchAllReleases(token: string) {
  let allNodes: Release[] = [];
  let totalCount = 0;
  let cursor: string | undefined;

  do {
    const res = await fetch("https://api.github.com/graphql", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ query: buildQuery(cursor) }),
    });

    if (!res.ok) {
      throw new Error("Failed to fetch releases");
    }

    const { data }: GraphQLResponse = await res.json();

    if (!data?.repository) {
      return { nodes: [], totalCount: 0 };
    }

    const releases = data.repository.releases;
    totalCount = releases.totalCount;
    allNodes = allNodes.concat(releases.nodes);

    if (releases.pageInfo.hasNextPage && releases.pageInfo.endCursor) {
      cursor = releases.pageInfo.endCursor;
    } else {
      break;
    }
  } while (true);

  return { nodes: allNodes, totalCount };
}

export async function GET() {
  try {
    if (!process.env.GITHUB_PAT) {
      return NextResponse.json(
        { error: "Server misconfiguration" },
        { status: 500 }
      );
    }

    const { nodes, totalCount } = await fetchAllReleases(process.env.GITHUB_PAT);

    let totalDownloads = 0;

    nodes.forEach((release) => {
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
      total_releases: totalCount,
    });
  } catch (error) {
    return NextResponse.json(
      { error: "Failed to calculate downloads" },
      { status: 500 }
    );
  }
}
