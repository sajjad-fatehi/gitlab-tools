# Quick Start Guide

Get started with gitlab-tools in 5 minutes.

## Step 1: Build the Tool

```bash
# Clone or navigate to the repository
cd gitlab-tools

# Build the binary
make build

# Or build manually
go build -o gitlab-tools ./cmd/gitlab-tools
```

## Step 2: Set Up GitLab Access

### Create a Personal Access Token

1. Log in to your GitLab instance
2. Navigate to **User Settings → Access Tokens**
3. Create a new token with the `api` scope
4. Save the token securely

### Configure Environment Variables

```bash
# Create a .env file from the example
cp .env.example .env

# Edit .env with your values
export GITLAB_BASE_URL="https://gitlab.example.com"
export GITLAB_TOKEN="your-token-here"

# Load the environment variables
export $(cat .env | xargs)
```

## Step 3: Explore Your GitLab Instance

### List Available Topics

```bash
./gitlab-tools topics
```

This shows all topics with beautiful formatting:

- Topic names with highlighting
- Project counts
- Descriptions

### Find Projects by Topic

```bash
./gitlab-tools projects --topic backend
```

This displays:

- Project names and paths
- Descriptions
- Associated topics
- Direct links to projects

## Step 4: Run Your First Bulk MR

### Example 1: Create MRs for Specific Projects

```bash
./gitlab-tools bulk-mr \
  --origin op-stage \
  --target op-rc \
  --project mygroup/frontend \
  --project mygroup/backend \
  --project mygroup/api
```

### Example 2: Create MRs for All Projects in a Topic

This is the easiest way when you have many projects:

```bash
# First, see what topics are available
./gitlab-tools topics

# Then create MRs for all projects in that topic
./gitlab-tools bulk-mr-topic \
  --origin op-stage \
  --target op-rc \
  --topic backend
```

This will:

- Automatically fetch all projects with the "backend" topic
- Create MRs for each project from op-stage to op-rc
- Skip drafts and existing MRs just like the regular bulk-mr command

### Using Group Prefix

If all your projects are in the same group:

```bash
./gitlab-tools bulk-mr \
  --origin op-stage \
  --target op-rc \
  --group mygroup \
  --project frontend \
  --project backend \
  --project api
```

## Step 5: Understand the Output

The tool will show you the status for each project:

```text
Processing 3 projects...

[mygroup/frontend] ✓ CREATED
  MR !42: https://gitlab.example.com/mygroup/frontend/-/merge_requests/42

[mygroup/backend] → SKIPPED_EXISTS
  Open MR already exists: !15

[mygroup/api] ⊘ SKIPPED_DRAFT
  Draft MR exists: !8 (Draft: Merge op-stage into op-rc)

Summary:
  Total projects: 3
  Created: 1
  Skipped (exists): 1
  Skipped (draft): 1
  Skipped (no branch): 0
  Errors: 0

✓ Completed successfully
```

## Common Scenarios

### Scenario 1: Test with Verbose Output

Enable detailed logging to see API calls:

```bash
./gitlab-tools bulk-mr \
  --origin op-stage \
  --target op-rc \
  --project mygroup/test-repo \
  --verbose
```

### Scenario 2: Override Token for Different User

```bash
./gitlab-tools bulk-mr \
  --origin op-stage \
  --target op-rc \
  --token "different-token" \
  --project mygroup/repo
```

### Scenario 3: Use Different GitLab Instance

```bash
./gitlab-tools bulk-mr \
  --origin feature-branch \
  --target main \
  --gitlab-url "https://other-gitlab.example.com" \
  --project group/repo
```

## Understanding Status Codes

- **✓ CREATED**: New MR successfully created
- **→ SKIPPED_EXISTS**: Open non-draft MR already exists
- **⊘ SKIPPED_DRAFT**: Only draft MRs exist for this branch pair
- **≡ SKIPPED_NO_CHANGE**: No changes between source and target branches
- **⚠ SKIPPED_NO_BRANCH**: Origin or target branch doesn't exist
- **✗ ERROR**: API or network error occurred

## Tips for Success

1. **Test First**: Try with a single project before running on many
2. **Use Verbose Mode**: Add `--verbose` when troubleshooting
3. **Check Branches**: Ensure both origin and target branches exist
4. **Verify Permissions**: Your token needs `api` scope
5. **Idempotent**: Safe to rerun - won't create duplicates

## Troubleshooting

### Error: "GitLab token must be provided"

Solution: Set the `GITLAB_TOKEN` environment variable or use `--token` flag.

### Error: "failed to get project"

Solutions:

- Verify the project path is correct (use `group/repo` format)
- Ensure your token has access to the project
- Check that the project exists on your GitLab instance

### Error: "Target branch 'X' does not exist"

Solution: Create the target branch in the repository first, or verify the branch name.

### No MRs Created (All Skipped)

Check if:

- MRs already exist for the branch pair
- Only draft MRs exist (tool skips these)
- Branches actually exist in each repository

## Next Steps

- Read the full [README](README.md) for detailed documentation
- Check out the [Makefile](Makefile) for build shortcuts
- Explore the code in `internal/` for customization

## Getting Help

For issues or questions:

1. Check the full README documentation
2. Review the command help: `./gitlab-tools bulk-mr --help`
3. Use verbose mode to see detailed API interactions
