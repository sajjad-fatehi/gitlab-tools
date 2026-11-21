# gitlab-tools

Command-line client for bulk merge requests and topic browsing on self-hosted GitLab (REST API v4). Written in Go.

## Install

Requires Go (see `go.mod`).

```bash
make install        # builds and copies to $(go env GOPATH)/bin/gitlab-tools
gitlab-tools version
```

Other targets: `make build`, `make test`, `make deps`, `make lint`, `make release`, `make clean`.

Pre-built binaries (linux/darwin amd64 + arm64, windows zip) are on the [releases page](https://github.com/sajjad-fatehi/gitlab-tools/releases):

```bash
curl -L https://github.com/sajjad-fatehi/gitlab-tools/releases/latest/download/gitlab-tools-darwin-arm64.tar.gz | tar xz
sudo mv gitlab-tools-darwin-arm64 /usr/local/bin/gitlab-tools
```

## Config

| Variable | Example |
|---|---|
| `GITLAB_BASE_URL` | `https://gitlab.example.com` |
| `GITLAB_TOKEN` | personal access token with the `api` scope |

Create the token under **User Settings → Access Tokens**. Variables are read from the environment only; there is no env file. `.env.example` shows the values to export.

```bash
export GITLAB_BASE_URL=https://gitlab.example.com
export GITLAB_TOKEN=...
gitlab-tools topics
```

## Global flags

Accepted by `topics`, `projects`, `bulk-mr` and `bulk-mr-topic`:

- `-gitlab-url <url>` GitLab base URL (default: `GITLAB_BASE_URL`)
- `-token <token>` API token (default: `GITLAB_TOKEN`)
- `-verbose` verbose logging

`merge` takes only `-target` and `-topic` and reads `GITLAB_BASE_URL` / `GITLAB_TOKEN` from the environment.

Flags are single-dash Go flags (`-origin`, also accepted as `--origin`).

## Commands

```bash
gitlab-tools topics                                                   # topics with project counts
gitlab-tools topics -page 2 -per-page 20                              # default -per-page 50
gitlab-tools projects -topic backend                                  # projects of a topic; -page, -per-page (default 50)
gitlab-tools bulk-mr -origin op-stage -target op-rc -project group/repo-a -project group/repo-b
gitlab-tools bulk-mr -origin op-stage -target op-rc -group group -project repo-a   # -group: default namespace prefix
gitlab-tools bulk-mr -origin op-stage -target op-rc -project group/repo-a -title "Release" -description "Notes"
gitlab-tools bulk-mr-topic -origin op-stage -target op-rc -topic backend           # -per-page default 100
gitlab-tools merge -target op-stage -topic backend                    # interactive, needs a terminal
gitlab-tools version
gitlab-tools <command> --help
```

`bulk-mr` options: `-origin` and `-target` are required, `-project` is repeatable, `-title` defaults to `Merge <origin> into <target>`, `-description` is auto-generated when omitted. `bulk-mr-topic` takes the same branch flags plus `-topic`, but no `-title` / `-description`.

Per project, `bulk-mr` and `bulk-mr-topic` check that both branches exist, look for an open MR for the same pair, skip when only draft MRs exist, and create an MR otherwise. Reruns are safe. Each project gets one status line, then a summary:

- `CREATED` new MR created
- `SKIPPED_EXISTS` open non-draft MR already exists
- `SKIPPED_DRAFT` only draft MRs exist for the pair
- `SKIPPED_NO_CHANGE` no changes between the branches
- `SKIPPED_NO_BRANCH` origin or target branch missing
- `ERROR` API or network error (exit code 1)

A merge request is a draft if its title starts with `Draft:` or `WIP:` (case-insensitive) or GitLab reports it as draft.

```text
[repo-a] CREATED
  MR !42: https://gitlab.example.com/group/repo-a/-/merge_requests/42
[repo-b] SKIPPED_EXISTS
  Open MR already exists: !15
```

`merge` lists the topic's projects, finds open non-draft MRs targeting `-target`, shows project, title, branches and URL, and asks `Merge this MR? (y/n)` for each. Accepted MRs are merged, then a summary of merged, skipped and failed counts is printed. It is interactive, so run it in a terminal.

## AI agent skill

`skills/gitlab-tools/` is an [Agent Skill](https://docs.claude.com/en/docs/claude-code/skills) (`SKILL.md` plus `references/`) that teaches an AI coding agent to drive this CLI: prerequisite check, read-only discovery, bulk MR recipes, and safety rules (never print the token, run `bulk-mr` / `bulk-mr-topic` only on explicit request, never run the interactive `merge`).

Install for Claude Code, from a clone of this repo. Personal (all projects):

```bash
mkdir -p ~/.claude/skills
cp -R skills/gitlab-tools ~/.claude/skills/gitlab-tools
# or keep it in sync with the repo:
ln -s "$PWD/skills/gitlab-tools" ~/.claude/skills/gitlab-tools
```

Project-local (commit it with the project):

```bash
mkdir -p .claude/skills
cp -R /path/to/gitlab-tools/skills/gitlab-tools .claude/skills/gitlab-tools
```

Restart the agent session so it picks up the skill. It loads when you mention gitlab-tools, bulk merge requests or GitLab topics. The agent still needs the `gitlab-tools` binary on `PATH` and `GITLAB_BASE_URL` / `GITLAB_TOKEN` exported (see [Config](#config)).

Other agents that read `SKILL.md` folders: copy `skills/gitlab-tools` into the skills directory your agent documents. The files are plain Markdown with `name` and `description` frontmatter; for an agent without skill support, paste `SKILL.md` into its instructions file.

## Out of scope

Branch cleanup, MR status reports, batch MR updates (labels, assignees, milestones), pipelines, parallel processing, GitLab versions without API v4.
