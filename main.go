package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/ui"
)

var version = "dev"

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println("tdocker", version)
		return
	}

	p := tea.NewProgram(ui.New())
	if _, err := p.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
