---
name: gitlab-tools
description: Drive the `gitlab-tools` CLI for self-hosted GitLab - list topics, list the projects of a topic, and open merge requests from one branch to another across many projects (`bulk-mr`, `bulk-mr-topic`), or guide an interactive bulk merge (`merge`). Use this skill whenever the user mentions gitlab-tools, bulk merge requests, opening MRs for every project in a topic, promoting a branch (e.g. stage to rc) across repositories, GitLab topics, or merging open MRs by topic and target branch, even if they do not name the tool.
---

# gitlab-tools

CLI for bulk merge requests and topic browsing on self-hosted GitLab (REST API v4). Full flag tables and output formats: `references/commands.md`.

## Before any command

Run the prerequisite check once per session:

```bash
command -v gitlab-tools && gitlab-tools version
[ -n "$GITLAB_BASE_URL" ] && echo "GITLAB_BASE_URL set" || echo "GITLAB_BASE_URL missing"
[ -n "$GITLAB_TOKEN" ] && echo "GITLAB_TOKEN set" || echo "GITLAB_TOKEN missing"
```

- Binary missing: tell the user to install it (`make install`, or a release binary from the GitHub releases page). Do not build or download it unasked.
- Variable missing: ask the user to export it (`GITLAB_BASE_URL=https://gitlab.example.com`, `GITLAB_TOKEN` = personal access token with the `api` scope). The tool reads the environment only, no env file.
- Never print, echo, log or commit the token. Never pass it as `-token` on a command line (it lands in shell history and process lists); rely on the env var. Check variables only with the `-n` test above, never `echo $GITLAB_TOKEN` or `env`.

## Safety tiers

| Tier | Commands | Rule |
|---|---|---|
| Read-only | `topics`, `projects`, `version`, any `--help` | Run freely. |
| Outward-facing | `bulk-mr`, `bulk-mr-topic` | Creates real merge requests on GitLab. Run only when the user explicitly asked for it in this conversation. Otherwise show the exact command and ask. |
| Interactive | `merge` | Merges MRs after a per-MR `y/n` prompt on stdin. Never run it, and never pipe `yes`/`y` into it: that would merge every open MR without a human decision. Print the exact command for the user to run in their own terminal. |

Why: MR creation and merging are visible to the whole team and hard to undo; the user must own that decision.

## Workflow

1. Discover scope (read-only) when the user is not sure what to target:
   - `gitlab-tools topics` - topics with project counts.
   - `gitlab-tools projects -topic <topic>` - project paths of a topic.
2. Confirm the plan back to the user before creating MRs: origin branch, target branch, which projects (explicit list or whole topic).
3. Create MRs (only on explicit request):
   - Chosen projects: `gitlab-tools bulk-mr -origin <branch> -target <branch> -project <path> [-project <path> ...]`
   - Whole topic: `gitlab-tools bulk-mr-topic -origin <branch> -target <branch> -topic <topic>`
4. Report the per-project status lines and the summary exactly as printed. Do not paraphrase `ERROR` lines away.

Flags are single-dash Go flags (`-origin`); double-dash (`--origin`) also works.

## Recipes

```bash
# list topics, then projects of one topic (paginate with -page / -per-page, default 50)
gitlab-tools topics
gitlab-tools projects -topic backend

# MRs for chosen projects (full paths)
gitlab-tools bulk-mr -origin op-stage -target op-rc -project group/repo-a -project group/repo-b

# short names plus a namespace prefix
gitlab-tools bulk-mr -origin op-stage -target op-rc -group group -project repo-a -project repo-b

# custom title / description (bulk-mr only)
gitlab-tools bulk-mr -origin op-stage -target op-rc -project group/repo-a -title "Release" -description "Notes"

# MRs for every project in a topic (no -title / -description support)
gitlab-tools bulk-mr-topic -origin op-stage -target op-rc -topic backend

# interactive merge: give this to the user, do not run it
gitlab-tools merge -target op-rc -topic backend
```

## Reading results

`bulk-mr` and `bulk-mr-topic` are idempotent: rerunning is safe. Per project they check both branches exist, look for an open MR for the pair, skip when only drafts exist, and create otherwise. Status per project:

- `CREATED` new MR (URL on the next line)
- `SKIPPED_EXISTS` open non-draft MR already there
- `SKIPPED_DRAFT` only draft MRs exist for the pair
- `SKIPPED_NO_CHANGE` no commits between the branches
- `SKIPPED_NO_BRANCH` origin or target missing in that project
- `ERROR` API/network failure; the command exits 1

Output contains ANSI colour codes and icons; read the status words, ignore the escapes.

## Common failures

- `Error: GitLab URL must be provided ...` / `GitLab token must be provided ...`: env vars missing; see the prerequisite check.
- `SKIPPED_NO_BRANCH` for many projects: wrong branch name or wrong topic; re-check with the user rather than retrying.
- `ERROR` with 401/403: token invalid or lacks `api` scope; ask the user to fix it. Do not retry in a loop.
- Unsure of a flag: run `gitlab-tools <command> --help`. Note `merge --help` prints only `-target` and `-topic`.
