# GitLab API Examples

This directory contains examples of using yapi to interact with the GitLab REST API.

## Examples

### get-project.yapi.yml
Fetches project information from a public GitLab repository.

```bash
yapi run examples/gitlab/get-project.yapi.yml
```

### list-repository-tree.yapi.yml
Lists files in a GitLab repository recursively, filtering for markdown files.

```bash
yapi run examples/gitlab/list-repository-tree.yapi.yml
```

### get-file-content.yapi.yml
Fetches a file's content from a GitLab repository. The content is base64-encoded and can be decoded using jq.

```bash
yapi run examples/gitlab/get-file-content.yapi.yml
```

### get-commits.yapi.yml
Fetches commit history for a specific file.

```bash
yapi run examples/gitlab/get-commits.yapi.yml
```

## Using with Private Repositories

For private repositories, you'll need to add a GitLab Personal Access Token. Uncomment the headers section in any example and set the `GITLAB_TOKEN` environment variable:

```yaml
headers:
  PRIVATE-TOKEN: ${GITLAB_TOKEN}
```

Then run:

```bash
export GITLAB_TOKEN=your-token-here
yapi run examples/gitlab/get-project.yapi.yml
```

## GitLab API Reference

- [GitLab REST API Documentation](https://docs.gitlab.com/ee/api/rest/)
- [Projects API](https://docs.gitlab.com/ee/api/projects.html)
- [Repository Files API](https://docs.gitlab.com/ee/api/repository_files.html)
- [Commits API](https://docs.gitlab.com/ee/api/commits.html)
