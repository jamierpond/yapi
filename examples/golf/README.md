# Golf Repository Examples

This directory contains examples demonstrating how to use yapi to access markdown files from a public GitLab repository. While named "golf" as a placeholder, these examples can be adapted to work with any GitLab repository containing markdown documentation.

## Examples

### list-markdown-files.yapi.yml
Lists all markdown files in a GitLab repository recursively.

```bash
yapi run examples/golf/list-markdown-files.yapi.yml
```

### get-markdown-file.yapi.yml
Fetches and displays a specific markdown file from the repository.

```bash
yapi run examples/golf/get-markdown-file.yapi.yml
```

## Customizing for Your Repository

To use these examples with your own GitLab repository:

1. Replace the project ID in the URL:
   - Current: `gitlab-org%2Fgitlab-docs`
   - Format: `namespace%2Fproject-name`
   - The `%2F` is a URL-encoded forward slash

2. Update the `ref` parameter to match your default branch (if not `main`)

3. For private repositories, add authentication:
   ```yaml
   headers:
     PRIVATE-TOKEN: ${GITLAB_TOKEN}
   ```

## Example: Adapting for a Golf Score Tracker

If you have a GitLab repository for tracking golf scores with markdown files:

```yaml
yapi: v1
url: https://gitlab.com/api/v4/projects/your-username%2Fgolf-scores/repository/tree

params:
  recursive: "true"
  per_page: "100"

jq_filter: ".[] | select(.type == \"blob\" and (.name | endswith(\".md\"))) | {name, path}"
```
