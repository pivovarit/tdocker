# tdocker

A minimalistic terminal UI for everyday Docker operations. Not a dashboard, not a full Docker management suite - just the handful of things you actually do dozens of times a day, each a single keypress away: tail logs, exec into a shell, restart a container, copy an ID.

https://github.com/user-attachments/assets/0bc1f303-46e1-4aa4-a755-72e769eeb3fd

### Philosophy

`tdocker` is intentionally small. It covers the most common Docker workflows and nothing more. No plugin system, no YAML configs, no container creation wizards. If an operation isn't something you'd do multiple times a week, it probably doesn't belong here.

### Getting started

Install with Homebrew:

```
brew install pivovarit/tap/tdocker
```

Or install with `go install`:

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

### Tested on

| Component | Version |
|-----------|---------|
| macOS | 26.3 (Sequoia) |
| Docker Desktop | 29.2.1 |
| Go | 1.26 |
| Architecture | arm64 (Apple Silicon) |

Clipboard integration is supported on macOS (`pbcopy`), Windows (`clip`), Linux/X11 (`xclip`), Linux/Wayland (`wl-copy`), and SSH/headless via OSC 52.

### Keybindings

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate |
| `/` | Filter containers |
| `A` | Toggle all / running only |
| `c` | Copy container ID to clipboard |
| `e` | Exec into container (`sh`) |
| `x` | Open debug shell (`docker debug`) |
| `i` | Inspect container |
| `t` | Show stats |
| `l` | View logs |
| `S` | Stop container |
| `s` | Start container |
| `R` | Restart container |
| `D` | Delete container |
| `r` | Refresh |
| `q` / `Ctrl+C` | Quit |
