package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"nemu/internal/dns"
	"nemu/internal/storage"
	"nemu/internal/tui"
)

func main() {
	store, err := storage.NewUserConfigStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "nemu: %v\n", err)
		os.Exit(1)
	}

	program := tea.NewProgram(tui.NewModel(dns.SystemResolver{}, store))
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "nemu: %v\n", err)
		os.Exit(1)
	}
}
