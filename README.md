# GitLab Tools

A command-line toolkit for managing branches and merge requests on self-hosted GitLab instances.

## Features

- **Bulk Merge Request Creation**: Create merge requests from an origin branch to a target branch across multiple repositories
  - Draft-aware: Skips projects where source branch MRs are already in draft
  - Idempotent: Safe to rerun without creating duplicates
  - Per-project feedback with clear status reporting
  - Bulk create MRs for all projects in a topic with a single command

- **Topics Management**: Browse and explore GitLab topics
  - List all available topics with project counts
  - View all projects associated with a specific topic
  - Beautiful terminal UI with colors and formatting

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/sajjad-fatehi/gitlab-tools/releases).

#### Linux (AMD64)

```bash
curl -L https://github.com/sajjad-fatehi/gitlab-tools/releases/latest/download/gitlab-tools-linux-amd64.tar.gz | tar xz
sudo mv gitlab-tools-linux-amd64 /usr/local/bin/gitlab-tools
chmod +x /usr/local/bin/gitlab-tools
```

#### Linux (ARM64)

```bash
curl -L https://github.com/sajjad-fatehi/gitlab-tools/releases/latest/download/gitlab-tools-linux-arm64.tar.gz | tar xz
sudo mv gitlab-tools-linux-arm64 /usr/local/bin/gitlab-tools
chmod +x /usr/local/bin/gitlab-tools
```

#### macOS (Intel)

```bash
curl -L https://github.com/sajjad-fatehi/gitlab-tools/releases/latest/download/gitlab-tools-darwin-amd64.tar.gz | tar xz
sudo mv gitlab-tools-darwin-amd64 /usr/local/bin/gitlab-tools
chmod +x /usr/local/bin/gitlab-tools
```

#### macOS (Apple Silicon)

```bash
curl -L https://github.com/sajjad-fatehi/gitlab-tools/releases/latest/download/gitlab-tools-darwin-arm64.tar.gz | tar xz
sudo mv gitlab-tools-darwin-arm64 /usr/local/bin/gitlab-tools
chmod +x /usr/local/bin/gitlab-tools
```

#### Windows

Download the appropriate `.zip` file for your architecture from the releases page and extract it to a directory in your PATH.

### Build from Source

If you prefer to build from source:

```bash
git clone https://github.com/sajjad-fatehi/gitlab-tools.git
cd gitlab-tools
go build -o gitlab-tools ./cmd/gitlab-tools
```

**Requirements:**

- Go 1.23 or later
- Access to a self-hosted GitLab instance
- GitLab personal access token with API access

## Configuration

### Environment Variables

- `GITLAB_BASE_URL`: Base URL for your GitLab instance (e.g., `https://gitlab.example.com`)
- `GITLAB_TOKEN`: Personal access token for GitLab API authentication

### Personal Access Token

Create a token in GitLab with the following scopes:

- `api`: Full API access (required for reading projects, branches, and creating merge requests)

Navigate to: **User Settings → Access Tokens**

## Usage

### List All Topics

View all available GitLab topics with their project counts:

```bash
export GITLAB_BASE_URL="https://gitlab.example.com"
export GITLAB_TOKEN="your-personal-access-token"

./gitlab-tools topics
```

With pagination:

```bash
./gitlab-tools topics --page 2 --per-page 20
```

### List Projects by Topic

View all projects that belong to a specific topic:

```bash
./gitlab-tools projects --topic backend
```

With pagination:

```bash
./gitlab-tools projects --topic backend --page 1 --per-page 30
```

### Bulk Merge Request Creation

#### For Specific Projects

Create merge requests from an origin branch to a target branch across multiple projects:

```bash
export GITLAB_BASE_URL="https://gitlab.example.com"
export GITLAB_TOKEN="your-personal-access-token"

./gitlab-tools bulk-mr \
  --origin op-stage \
  --target op-rc \
  --project group/repo-a \
  --project group/repo-b
```

#### For All Projects in a Topic

Create merge requests for all projects that have a specific topic:

```bash
./gitlab-tools bulk-mr-topic \
  --origin op-stage \
  --target op-rc \
  --topic backend
```

This automatically:

- Fetches all projects with the specified topic
- Creates MRs for each project
- Provides the same draft-aware and idempotent behavior

#### Command Options

- `--origin`: Source branch name (required)
- `--target`: Target branch name (required)
- `--project`: Project path (can be repeated for multiple projects)
- `--group`: Default group/namespace prefix (optional)

### Interactive Merge Command

Interactively merge open, non-draft merge requests targeting a specific branch across all projects in a topic:

```bash
./gitlab-tools merge \
  --target op-stage \
  --topic backend
```

This command will:

1. **Fetch all projects** with the specified topic
2. **Find open MRs** targeting the specified branch in each project
3. **Filter out draft MRs** automatically
4. **Display MR details** including:
   - Project name
   - MR title
   - Source and target branches
   - Web URL for reference
5. **Prompt for confirmation** for each MR (y/n)
6. **Merge accepted MRs** automatically
7. **Show a summary** of merged, skipped, and failed operations

#### Example Session

```console
📦 Fetching projects for topic: backend
✓ Found 12 projects

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Project: example/backend-api
MR Title: feat: Add new authentication endpoint
Branches: feature/auth → op-stage
URL: https://git.example.com/example/backend-api/-/merge_requests/42
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Merge this MR? (y/n): y
✓ Successfully merged!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Summary
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ Merged:  3
⊘ Skipped: 1
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

#### Merge Command Options

- `--target`: Target branch to merge into (required)
- `--topic`: Topic to filter projects (required)
- `--gitlab-url`: Override GitLab base URL (optional, uses `GITLAB_BASE_URL` env var)
- `--token`: Override GitLab token (optional, uses `GITLAB_TOKEN` env var)
- `--verbose`: Enable detailed logging (optional)

#### Behavior

The tool will:

1. **Validate branches**: Check that both origin and target branches exist
2. **Check existing MRs**: Look for open merge requests with the same source/target pair
3. **Skip draft contexts**: If only draft MRs exist, skip creation
4. **Create MRs**: Only create when no open MR exists or only closed/merged MRs exist
5. **Report results**: Show per-project status and summary

#### Status Codes

- `CREATED`: New merge request successfully created
- `SKIPPED_EXISTS`: Open non-draft MR already exists
- `SKIPPED_DRAFT`: Only draft MRs exist for this branch pair
- `SKIPPED_NO_CHANGE`: No changes between source and target branches
- `SKIPPED_NO_BRANCH`: Origin or target branch doesn't exist
- `ERROR`: API or network error occurred

#### Example Output

```text
Processing 3 projects...

[repo-a] CREATED
  MR !42: https://gitlab.example.com/group/repo-a/-/merge_requests/42

[repo-b] SKIPPED_EXISTS
  Open MR already exists: !15

[repo-c] SKIPPED_DRAFT
  Draft MR exists: !8 (Draft: Merge op-stage into op-rc)

Summary:
  Created: 1
  Skipped (exists): 1
  Skipped (draft): 1
  Skipped (no changes): 0
  Skipped (no branch): 0
  Errors: 0

✓ Completed successfully
```

### Draft Detection

A merge request is considered a draft if:

- The title starts with `Draft:` or `WIP:` (case-insensitive)
- GitLab API indicates draft status via the `draft` field

## Project Structure

```text
gitlab-tools/
├── cmd/
│   └── gitlab-tools/
│       └── main.go           # CLI entry point
├── internal/
│   ├── gitlab/
│   │   ├── client.go         # GitLab API client
│   │   └── types.go          # Domain models
│   └── bulkmr/
│       └── service.go        # Bulk MR creation logic
├── go.mod
└── README.md
```

## Development

### Running Tests

```bash
go test ./...
```

### Running with Verbose Output

```bash
./gitlab-tools bulk-mr --verbose --origin op-stage --target op-rc --project mygroup/myrepo
```

## Limitations

- Requires GitLab API v4 (most modern self-hosted instances)
- Sequential processing (no parallelization in initial version)
- Basic MR configuration (no custom labels, assignees, or milestones)

## Future Enhancements

Planned features for this toolkit:

- Branch cleanup commands
- MR status reporting
- Batch MR updates (labels, assignees)
- Pipeline management

## License

MIT License - see [LICENSE](LICENSE) file for details.
