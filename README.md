<p align="center">
<pre>
              ██████╗ ███████╗ ██████╗
             ██╔════╝ ██╔════╝██╔═══██╗
             ██║      ███████╗██║   ██║
             ██║      ╚════██║██║▄▄ ██║
             ╚██████╗ ███████║╚██████╔╝
              ╚═════╝ ╚══════╝ ╚══▀▀═╝
       Multi-Project Claude Squad Launcher
</pre>
</p>

<p align="center">
  <a href="#install">Install</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#how-it-works">How It Works</a> •
  <a href="#commands">Commands</a> •
  <a href="#configuration">Configuration</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go 1.21+">
  <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey?style=flat" alt="macOS | Linux">
  <img src="https://img.shields.io/badge/license-MIT-blue?style=flat" alt="MIT License">
  <img src="https://img.shields.io/badge/PRs-welcome-brightgreen?style=flat" alt="PRs Welcome">
</p>

---

`csq` wraps [Claude Squad](https://github.com/smtg-ai/claude-squad) (`cs`) to add **per-project state isolation**, **fuzzy project search**, and **automatic worktree bootstrapping**.

### The Problem

Claude Squad stores all state in `~/.claude-squad/`. When you work across many projects with parallel sessions, state from different repos gets mixed together. Switching projects means losing track of which sessions belong where.

### The Solution

`csq` gives each project its own isolated `HOME` directory, so Claude Squad thinks it's running independently per project. Your sessions, worktrees, and state stay cleanly separated.

---

## Install

### Prerequisites

`csq` is written in Go. If you don't have Go installed:

<details>
<summary><strong>macOS</strong> (Homebrew)</summary>

```bash
brew install go
```
</details>

<details>
<summary><strong>Linux</strong> (apt/dnf)</summary>

```bash
# Debian / Ubuntu
sudo apt update && sudo apt install -y golang

# Fedora / RHEL
sudo dnf install -y golang
```
</details>

<details>
<summary><strong>Any platform</strong> (official installer)</summary>

Download from [go.dev/dl](https://go.dev/dl/) and follow the instructions for your OS.

After installing, verify with:

```bash
go version   # should print go1.21 or later
```
</details>

You'll also need [Claude Squad](https://github.com/smtg-ai/claude-squad) (`cs`) installed and available in your `PATH`.

### Build from source

```bash
git clone https://github.com/f3r/csq.git
cd csq
make install   # builds the binary and copies it to ~/bin/csq
```

> **Note:** Make sure `~/bin` is in your `PATH`. Add this to your `~/.zshrc` or `~/.bashrc` if needed:
> ```bash
> export PATH="$HOME/bin:$PATH"
> ```

---

## Quick Start

```bash
# 1. One-time setup — creates ~/.csq/ and installs the bootstrap hook
csq init

# 2. Launch the fuzzy project picker
csq

# 3. Jump straight to a project (fuzzy match)
csq my-api

# 4. List all projects with session counts
csq list

# 5. See all active sessions across every project
csq status
```

---

## How It Works

### Project Discovery

`csq` recursively scans your code directories for git repositories. By default it looks in `~/code/` up to 3 levels deep. Results are cached for 5 minutes.

### State Isolation

When you select a project, `csq` creates an isolated home directory at `~/.csq/homes/<project>/` and launches `cs` with `HOME` pointed there. Each project gets its own `.claude-squad/` directory with independent session state.

Essential dotfiles (`.gitconfig`, `.ssh`, `.claude`, etc.) are symlinked back to your real home so everything works normally.

### Worktree Bootstrapping

A `SessionStart` hook (installed by `csq init`) runs whenever Claude Code starts a new session. If it detects a git worktree, it automatically:

1. Copies `.env*` files from the main checkout
2. Copies files listed in `.csq-copy` (see [below](#csq-copy-file))
3. Initializes git submodules
4. Installs dependencies (detects npm/yarn/pnpm/go/bundler/pip)

This runs once per worktree and is fully idempotent.

---

## Commands

### `csq`

Opens an interactive fuzzy picker showing all discovered projects. Type to filter, arrow keys to navigate, Enter to select.

```
  Search: web█

  > acme/web-app       [2 sessions]  ~/code/acme/web-app
    acme/web-docs                     ~/code/acme/web-docs

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
PROJECT              TITLE                     STATUS   BRANCH
acme/api             fix auth middleware        running  cs/fix-auth
acme/api             add rate limiting          paused   cs/rate-limit
personal/blog        migrate to astro           running  cs/astro-migration

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

---

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

---

## `.csq-copy` File

Place a `.csq-copy` file in your project root to list additional gitignored files that should be copied into worktrees. One path per line:

```
# .csq-copy
global-bundle.pem
config/local.yml
.env.local
```

Lines starting with `#` are ignored. Paths are relative to the project root.

---

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

---

## Development

```bash
make build        # compile binary
make test         # run all tests
make test-cover   # run tests with coverage report
make install      # build + copy to ~/bin
make clean        # remove build artifacts
```

---

## License

MIT
