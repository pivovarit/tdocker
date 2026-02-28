# tdocker

A terminal UI for Docker container management. It displays your running (or all) containers in an interactive table, lets you navigate, filter, and inspect them, and provides quick actions - exec into a shell, view logs, start/stop/delete containers - all without leaving the terminal.

### Keybindings

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate |
| `/` | Filter containers |
| `a` | Toggle all / running only |
| `c` | Copy container ID to clipboard |
| `e` | Exec into container (`sh`) |
| `x` | Open debug shell (`docker debug`) |
| `i` | Inspect container |
| `t` | Show stats |
| `l` | View logs |
| `s` | Stop container |
| `S` | Start container |
| `d` | Delete container |
| `r` | Refresh |
| `q` / `Ctrl+C` | Quit |
