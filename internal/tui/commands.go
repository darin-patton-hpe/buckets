package tui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/darin-patton-hpe/buckets/internal/data"
	"github.com/darin-patton-hpe/nbalive/live"
)

const scoreboardInterval = 30 * time.Second

// fetchScoreboardCmd fetches today's scoreboard.
func fetchScoreboardCmd(client data.NBAClient) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.Scoreboard(context.Background())
		if err != nil {
			return scoreboardMsg{err: err}
		}
		return scoreboardMsg{games: resp.Scoreboard.Games}
	}
}

// fetchBoxScoreCmd fetches the box score for a game.
func fetchBoxScoreCmd(client data.NBAClient, gameID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.BoxScore(context.Background(), gameID)
		if err != nil {
			return boxScoreMsg{err: err}
		}
		return boxScoreMsg{game: &resp.Game}
	}
}

// fetchPlayByPlayCmd fetches the play-by-play for a game.
func fetchPlayByPlayCmd(client data.NBAClient, gameID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.PlayByPlay(context.Background(), gameID)
		if err != nil {
			return playByPlayMsg{err: err}
		}
		return playByPlayMsg{actions: resp.Game.Actions}
	}
}

// waitWatchCmd blocks on the watch channel and returns events one at a time.
// Returns watchClosedMsg when the channel closes.
func waitWatchCmd(ch <-chan live.Event) tea.Cmd {
	return func() tea.Msg {
		evt, ok := <-ch
		if !ok {
			return watchClosedMsg{}
		}
		return watchEventMsg{event: evt}
	}
}

// scoreboardTickCmd returns a tick command for periodic scoreboard refresh.
func scoreboardTickCmd() tea.Cmd {
	return tea.Tick(scoreboardInterval, func(t time.Time) tea.Msg {
		return scoreboardTickMsg(t)
	})
}
