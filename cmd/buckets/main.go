package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/darin-patton-hpe/buckets/internal/data"
	"github.com/darin-patton-hpe/buckets/internal/tui"
	"github.com/darin-patton-hpe/nbalive/live"
)

func main() {
	client := data.NewLiveClient(live.NewClient())
	m := tui.NewModel(client)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "buckets: %v\n", err)
		os.Exit(1)
	}
}
