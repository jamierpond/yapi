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
      next: { revalidate: 3600 },
    });

    if (!res.ok) {
      throw new Error(`GitHub API returned ${res.status}`);
    }

    const json = await res.json();
    const { data } = json as GraphQLResponse;

    if (json.errors || !data?.repository) {
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

export async function getGitHubStars(): Promise<number | null> {
  try {
    const res = await fetch("https://api.github.com/repos/jamierpond/yapi", {
      next: { revalidate: 3600 },
    });
    if (!res.ok) return null;
    const data = await res.json();
    return data.stargazers_count;
  } catch {
    return null;
  }
}

export async function getTotalDownloads(): Promise<number | null> {
  try {
    const token = process.env.GITHUB_PAT;
    if (!token) return null;

    const { nodes } = await fetchAllReleases(token);

    let total = 0;
    nodes.forEach((release) => {
      release.releaseAssets.nodes.forEach((asset) => {
        if (asset.name !== "checksums.txt") {
          total += asset.downloadCount;
        }
      });
    });

    return total;
  } catch {
    return null;
  }
}
