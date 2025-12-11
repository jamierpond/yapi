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
      const errorBody = await res.text();
      console.error(`GitHub API error: ${res.status} ${res.statusText}`, {
        status: res.status,
        statusText: res.statusText,
        body: errorBody,
        cursor,
      });
      throw new Error(`GitHub API returned ${res.status}: ${errorBody}`);
    }

    const json = await res.json();
    const { data } = json as GraphQLResponse;

    if (json.errors) {
      console.error("GitHub GraphQL errors:", json.errors);
      throw new Error(`GitHub GraphQL errors: ${JSON.stringify(json.errors)}`);
    }

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
      console.error("GITHUB_PAT environment variable is not set");
      return NextResponse.json(
        { error: "Server misconfiguration" },
        { status: 500 }
      );
    }

    const { nodes, totalCount } = await fetchAllReleases(process.env.GITHUB_PAT);

    let allTimeDownloads = 0;
    let latestReleaseDownloads = 0;

    const latestReleaseName = nodes[0]?.name;

    nodes.forEach((release) => {
      let releaseDownloads = 0;

      release.releaseAssets.nodes.forEach((asset) => {
        if (asset.name === "checksums.txt") {
          return;
        }
        const realCount = Math.max(0, asset.downloadCount - 1);
        releaseDownloads += realCount;
      });

      allTimeDownloads += releaseDownloads;

      if (release.name === latestReleaseName) {
        latestReleaseDownloads = releaseDownloads;
      }
    });

    return NextResponse.json(
      {
        total_downloads: allTimeDownloads,
        active_users_proxy: latestReleaseDownloads,
        total_releases: totalCount,
      },
      {
        headers: {
          "Cache-Control": "public, s-maxage=3600, stale-while-revalidate=59",
        },
      }
    );
  } catch (error) {
    console.error("Failed to fetch download stats:", error);
    return NextResponse.json(
      { error: "Failed to calculate downloads" },
      { status: 500 }
    );
  }
}
