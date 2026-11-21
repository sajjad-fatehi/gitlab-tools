# gitlab-tools command reference

Source of truth: `gitlab-tools <command> --help` and `cmd/gitlab-tools/main.go`. Flags are single-dash Go flags; `--flag` also works.

## Contents

- Configuration
- topics
- projects
- bulk-mr
- bulk-mr-topic
- merge
- version
- Draft detection
- Exit codes

## Configuration

| Variable | Meaning |
|---|---|
| `GITLAB_BASE_URL` | Instance URL, e.g. `https://gitlab.example.com` |
| `GITLAB_TOKEN` | Personal access token with the `api` scope; sent as the `PRIVATE-TOKEN` header |

`topics`, `projects`, `bulk-mr`, `bulk-mr-topic` also accept `-gitlab-url` (default: `GITLAB_BASE_URL`) and `-token` (default: `GITLAB_TOKEN`). Prefer the env vars: a `-token` value is visible in shell history and process lists. `merge` reads the env vars only.

Note: `--help` prints the current `GITLAB_BASE_URL` value as the `-gitlab-url` default. The token value is never printed.

## topics

`gitlab-tools topics [-page N] [-per-page N] [-verbose]` - lists topics with project counts. Defaults: `-page 1`, `-per-page 50`. Read-only.

## projects

`gitlab-tools projects -topic <name> [-page N] [-per-page N] [-verbose]` - lists projects of a topic with their `path_with_namespace`, which is the value `bulk-mr -project` expects. `-topic` required. Defaults: `-page 1`, `-per-page 50`. Read-only.

## bulk-mr

`gitlab-tools bulk-mr -origin <branch> -target <branch> -project <path> [-project <path> ...]`

| Flag | Notes |
|---|---|
| `-origin` | required, source branch |
| `-target` | required, target branch |
| `-project` | required, repeatable; full path `group/repo` |
| `-group` | namespace prefix, applied only to `-project` values without a `/` |
| `-title` | default `Merge <origin> into <target>` |
| `-description` | default auto-generated |
| `-gitlab-url`, `-token`, `-verbose` | see Configuration |

Creates merge requests: outward-facing. Per project, in order: check both branches exist (`SKIPPED_NO_BRANCH`), look for open MRs for the pair (`SKIPPED_EXISTS` if a non-draft exists, `SKIPPED_DRAFT` if only drafts), compare branches (`SKIPPED_NO_CHANGE` if no commits differ), create (`CREATED`). Failures give `ERROR`. Safe to rerun.

## bulk-mr-topic

`gitlab-tools bulk-mr-topic -origin <branch> -target <branch> -topic <name> [-per-page N]`

Same as `bulk-mr` but the project list is every project of the topic (paginated, `-per-page` default 100). No `-group`, `-title`, `-description`; uses the default title. Also accepts `-gitlab-url`, `-token`, `-verbose`. Prints `No projects found for topic: <name>` and exits 0 when the topic is empty. Outward-facing.

## merge

`gitlab-tools merge -target <branch> -topic <name>` - both required; only these two flags.

Lists the topic's projects (first 100 only), finds open non-draft MRs targeting `-target`, prints project, title, branches and URL, then asks `Merge this MR? (y/n)` on stdin for each. `y`/`yes` merges, anything else skips. Prints merged / skipped / error counts; exits 1 on errors.

Interactive and irreversible per answer. An agent must not run it or feed answers to stdin. Hand the user this line to run themselves:

```bash
gitlab-tools merge -target <branch> -topic <topic>
```

## version

`gitlab-tools version` - prints `gitlab-tools vX.Y.Z`. `gitlab-tools help` prints the command list.

## Draft detection

An MR counts as draft when GitLab reports `draft`, or its title starts with `Draft:` or `WIP:` (case-insensitive). Drafts are ignored by `merge` and block creation in `bulk-mr*` (`SKIPPED_DRAFT`).

## Exit codes

- `0` completed, including runs with only skips
- `1` any `ERROR` result, missing required flag, missing URL/token, or API failure while listing projects
