package main

import (
	"context"
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

	ctx, cancel := context.WithCancel(context.Background())

	p := tea.NewProgram(ui.New(ctx, version))
	_, err := p.Run()
	cancel()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
