# tdocker

`tdocker` is an extremely lightweight zero-config `docker ps` alternative that you actually reach for dozens of times a day.

The operations that used to require remembering container IDs and chaining CLI commands are now a single keypress: 
- tail logs 
- stop or restart a container
- exec into a shell
- copy an ID to the clipboard
- inspect configuration, 
- check CPU and memory usage

https://github.com/user-attachments/assets/743ed68b-a001-4f3a-99a7-cd2e342ae9f6

### Getting started

```
git clone https://github.com/pivovarit/tdocker && cd tdocker && go run .
```

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
| `s` | Stop container |
| `S` | Start container |
| `d` | Delete container |
| `r` | Refresh |
| `q` / `Ctrl+C` | Quit |
