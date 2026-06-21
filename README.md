# tdocker

A minimalistic terminal UI for everyday Docker operations. Not a dashboard, not a full Docker management suite - just the handful of things you actually do dozens of times a day, each a single keypress away: tail logs, exec into a shell, restart a container, copy an ID.

![tdocker demo](docs/tdocker.gif)

### Philosophy

`tdocker` is intentionally small. It covers the most common Docker workflows and nothing more. No plugin system, no YAML configs, no container creation wizards. If an operation isn't something you'd do multiple times a week, it probably doesn't belong here.

### Getting started

Install with Homebrew (macOS or Linux):

```
brew install pivovarit/tap/tdocker
```

#### Linux

Grab a prebuilt binary (amd64):

```
curl -sSL https://github.com/pivovarit/tdocker/releases/latest/download/tdocker_linux_amd64.tar.gz | tar xz
sudo mv tdocker /usr/local/bin/
```

For arm64, swap `amd64` for `arm64`.

Or install a native package from the [latest release](https://github.com/pivovarit/tdocker/releases/latest):

| Distro | Package |
|--------|---------|
| Debian / Ubuntu | `.deb` |
| Fedora / RHEL / openSUSE | `.rpm` |
| Alpine | `.apk` |
| Arch | `.pkg.tar.zst` |

#### From source

Install with `go install`:

```
go install github.com/pivovarit/tdocker@latest
```

Make sure `$GOPATH/bin` is on your `$PATH` (the default is `~/go/bin`):

```
export PATH="$PATH:$(go env GOPATH)/bin"
```

Add the line above to your shell profile (`.bashrc`, `.zshrc`, etc.) to make it permanent.

Then launch:

```
tdocker
```

Or run directly from source:

```
git clone https://github.com/pivovarit/tdocker && cd tdocker && go run .
```

### Built & verified on

| Environment | Details |
|-------------|---------|
| macOS | 26.3 (Sequoia), arm64 (Apple Silicon), Docker Desktop 29.2.1 |
| Linux | Ubuntu (CI), unit + Testcontainers integration tests on every push |
| Go | 1.26 |

Released binaries are built for Linux and macOS on both `amd64` and `arm64`.

Clipboard integration is supported on macOS (`pbcopy`), Windows (`clip`), Linux/X11 (`xclip`), Linux/Wayland (`wl-copy`), and SSH/headless via OSC 52.

### Keybindings

| Key | Action |
|-----|--------|
| `↑` / `↓` / `j` / `k` | Navigate |
| `g` / `G` | Jump to top / bottom |
| `→` / `←` | Expand inline details / collapse |
| `/` | Filter containers |
| `A` | Toggle all / running only |
| `l` | View logs |
| `e` | Exec into container (`sh`) |
| `x` | Open debug shell (`docker debug`) |
| `i` | Inspect container (image, env, ports, mounts, networks) |
| `I` | Diagnose container (state, events, logs, health, config) |
| `t` | Show stats |
| `v` | Stream Docker events |
| `c` | Copy container ID to clipboard |
| `S` | Start / Stop (toggles by state) |
| `R` | Restart container |
| `P` | Pause / Unpause |
| `D` | Delete container (stopped only) |
| `N` | Rename container |
| `r` | Refresh |
| `X` | Switch Docker context |
| `?` | Show help |
| `q` / `Ctrl+C` | Quit |

### Tips & Hints

- **Navigate while filtering** - press `↑`/`↓` while typing a filter to accept it and immediately navigate the list
- **`q` clears filters first** - if a filter is active, `q` clears it instead of quitting; press again to exit
- **Inline details** - press `→` on any container to expand port bindings and network info as navigable rows inline; `←` collapses them
- **Compose groups** - `→` and `←` also expand and collapse Compose project groups
- **Auto-scroll in logs** - logs auto-scroll as new lines arrive; scroll up to pause, scroll back to the bottom to resume
- **Smart actions** - `S` stops running containers and starts stopped ones; `R` restarts running containers and starts stopped ones
- **Shell detection** - `e` auto-detects available shells; for distroless/scratch images, use `x` (docker debug) instead
- **`i` vs `I`** - `i` shows the classic inspect summary (image, env, ports, mounts); `I` opens the diagnostic panel with container state, recent events, log tail, and healthcheck results
