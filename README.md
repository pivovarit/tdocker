# tdocker

`tdocker` is an extremely lightweight zero-config `docker ps` alternative that you actually reach for dozens of times a day.

The operations that used to require remembering container IDs and chaining CLI commands are now a single keypress: 
- tail logs 
- stop or restart a container
- exec into a shell
- copy an ID to the clipboard
- inspect configuration, 
- check CPU and memory usage

https://github.com/user-attachments/assets/0bc1f303-46e1-4aa4-a755-72e769eeb3fd

### Philosophy

`tdocker` does less on purpose. Every feature exists because it came up during real, day-to-day work with Docker - not to cover edge cases or satisfy a checklist. The goal is a tool you actually reach for, not one you install and forget. If an operation isn't part of a typical workflow, it doesn't belong here.

### Getting started

> Homebrew support is coming as soon as I figure out how to set it up.

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
