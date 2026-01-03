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

export async function getGitHubStats(): Promise<{ stars: number | null; forks: number | null }> {
  try {
    const token = process.env.GITHUB_PAT;
    const res = await fetch("https://api.github.com/repos/jamierpond/yapi", {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
      next: { revalidate: 3600 },
    });
    if (!res.ok) {
      console.error(`[getGitHubStats] GitHub API error: ${res.status} ${res.statusText}`);
      return { stars: null, forks: null };
    }
    const data = await res.json();
    return {
      stars: data.stargazers_count ?? null,
      forks: data.forks_count ?? null,
    };
  } catch (err) {
    console.error("[getGitHubStats] Error:", err);
    return { stars: null, forks: null };
  }
}

// Keep for backwards compatibility
export async function getGitHubStars(): Promise<number | null> {
  const { stars } = await getGitHubStats();
  return stars;
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

// VS Code Marketplace stats
type VSCodeStatistic = {
  statisticName: string;
  value: number;
};

type VSCodeExtension = {
  statistics?: VSCodeStatistic[];
};

type VSCodeQueryResponse = {
  results?: Array<{
    extensions?: VSCodeExtension[];
  }>;
};

export async function getVSCodeInstalls(): Promise<number | null> {
  try {
    const res = await fetch(
      "https://marketplace.visualstudio.com/_apis/public/gallery/extensionquery?api-version=7.2-preview.1",
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          filters: [
            {
              criteria: [{ filterType: 7, value: "yapi.yapi-extension" }],
              pageSize: 1,
            },
          ],
          flags: 914,
        }),
        next: { revalidate: 3600 },
      }
    );

    if (!res.ok) return null;

    const data: VSCodeQueryResponse = await res.json();
    const stats = data.results?.[0]?.extensions?.[0]?.statistics;
    const installStat = stats?.find((s) => s.statisticName === "install");
    return installStat?.value ?? null;
  } catch {
    return null;
  }
}

export async function getOpenVSXDownloads(): Promise<number | null> {
  try {
    const res = await fetch(
      "https://open-vsx.org/api/yapi/yapi-extension",
      { next: { revalidate: 3600 } }
    );

    if (!res.ok) return null;

    const data = await res.json();
    return data.downloadCount ?? null;
  } catch {
    return null;
  }
}

