# csq — Multi-Project Claude Squad Launcher

`csq` wraps [Claude Squad](https://github.com/smtg-ai/claude-squad) (`cs`) to add per-project state isolation, fuzzy project search, and automatic worktree bootstrapping.

**The problem:** Claude Squad stores all state in `~/.claude-squad/`. When you work across many projects with parallel sessions, state from different repos gets mixed together. Switching projects means losing track of which sessions belong where.

**The solution:** `csq` gives each project its own isolated `HOME` directory, so Claude Squad thinks it's running independently per project. Your sessions, worktrees, and state stay cleanly separated.

## Install

```bash
# Requires Go 1.21+
git clone https://github.com/f3r/csq.git
cd csq
make install   # builds and copies binary to ~/bin/csq
```

Make sure `~/bin` is in your `PATH`.

## Quick Start

```bash
# One-time setup: creates ~/.csq/ and installs the bootstrap hook
csq init

# Launch the fuzzy project picker
csq

# Jump straight to a project (fuzzy match)
csq my-api

# List all projects with session counts
csq list

# See all active sessions across every project
csq status
```

## How It Works

### Project Discovery

`csq` recursively scans your code directories for git repositories. By default it looks in `~/code/` up to 3 levels deep. Results are cached for 5 minutes.

### State Isolation

When you select a project, `csq` creates an isolated home directory at `~/.csq/homes/<project>/` and launches `cs` with `HOME` pointed there. Each project gets its own `.claude-squad/` directory with independent session state.

Essential dotfiles (`.gitconfig`, `.ssh`, `.claude`, etc.) are symlinked back to your real home so everything works normally.

### Worktree Bootstrapping

A `SessionStart` hook (installed by `csq init`) runs whenever Claude Code starts a new session. If it detects a git worktree, it automatically:

1. Copies `.env*` files from the main checkout
2. Copies files listed in `.csq-copy` (see below)
3. Initializes git submodules
4. Installs dependencies (detects npm/yarn/pnpm/go/bundler/pip)

This runs once per worktree and is fully idempotent.

## Commands

### `csq`

Opens an interactive fuzzy picker showing all discovered projects. Type to filter, arrow keys to navigate, Enter to select.

```
  Search: web█

  > acme/web-app       [2 sessions]  ~/code/acme/web-app
    acme/web-docs                       ~/code/acme/web-docs

  15 projects · 2 matching
```

### `csq <name>`

Fuzzy-matches the project name. If there's exactly one match, launches immediately. If multiple match, opens the picker pre-filtered.

```bash
csq my-api           # exact or unique match → launches directly
csq web              # ambiguous → opens picker filtered to "web"
```

### `csq list` (alias: `csq ls`)

Shows all discovered projects with their session counts.

```
PROJECT              SESSIONS  PATH
acme/api             2         ~/code/acme/api
acme/web-app         -         ~/code/acme/web-app
personal/dotfiles    1         ~/code/personal/dotfiles
```

Use `-r` to refresh the project cache.

### `csq status`

Aggregates all active Claude Squad sessions across every project.

```
PROJECT          TITLE                     STATUS   BRANCH
acme/api   fix auth middleware        running  cs/fix-auth
acme/api   add rate limiting          paused   cs/rate-limit
personal/blog    migrate to astro           running  cs/astro-migration

3 session(s) across 2 project(s)
```

### `csq init`

One-time setup. Creates `~/.csq/`, writes default config, generates the bootstrap script, and installs the `SessionStart` hook in `~/.claude/settings.json`.

### `csq config`

View or edit configuration.

```bash
csq config                          # show full config
csq config roots                    # show specific key
csq config max_depth 4              # set a value
csq config roots ~/code,~/work      # comma-separated for arrays
```

### Passing flags to `cs`

Use `--` to pass arguments through to Claude Squad:

```bash
csq my-api -- -y
```

## Configuration

Config lives at `~/.csq/config.json`:

```json
{
  "roots": ["~/code"],
  "max_depth": 3,
  "cs_binary": "cs",
  "home_base": "~/.csq/homes",
  "symlink_dotfiles": [
    ".gitconfig", ".ssh", ".claude", ".config",
    ".zshrc", ".zshenv", "bin", ".nvm", ".cargo"
  ]
}
```

| Key | Description | Default |
|-----|-------------|---------|
| `roots` | Directories to scan for git repos | `["~/code"]` |
| `max_depth` | How deep to recurse when scanning | `3` |
| `cs_binary` | Path to the `cs` binary | `"cs"` |
| `home_base` | Where isolated home dirs are created | `"~/.csq/homes"` |
| `symlink_dotfiles` | Files/dirs symlinked from real home into each project home | see above |

## `.csq-copy` File

Place a `.csq-copy` file in your project root to list additional gitignored files that should be copied into worktrees. One path per line:

```
# .csq-copy
global-bundle.pem
config/local.yml
.env.local
```

Lines starting with `#` are ignored. Paths are relative to the project root.

## File Layout

```
~/.csq/
├── config.json          # csq configuration
├── bootstrap.sh         # worktree bootstrap script (SessionStart hook)
├── cache.json           # project discovery cache (auto-generated)
└── homes/
    ├── acme--api/
    │   ├── .claude-squad/   ← independent state per project
    │   ├── .gitconfig → ~/.gitconfig
    │   ├── .ssh → ~/.ssh
    │   └── .claude → ~/.claude
    └── personal--blog/
        ├── .claude-squad/
        └── ...
```
