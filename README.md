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

### Getting started

Install with `go install`:

```
go install github.com/pivovarit/tdocker@latest
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
