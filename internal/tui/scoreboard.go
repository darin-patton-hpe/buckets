package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/darin-patton-hpe/nbalive"
)

// renderScoreboard renders the full scoreboard view.
func renderScoreboard(games []nbalive.Game, cursor int, width int, s styles) string {
	var b strings.Builder

	b.WriteString(s.title.Render("🏀 NBA Scoreboard"))
	b.WriteString("\n")

	if len(games) == 0 {
		b.WriteString(s.dimText.Render("\n  No games today.\n"))
		return b.String()
	}

	// Date from first game.
	if len(games) > 0 {
		b.WriteString(s.subtitle.Render(games[0].GameET))
		b.WriteString("\n\n")
	}

	for i, g := range games {
		row := renderGameRow(g, s)
		if i == cursor {
			// Highlight selected row.
			b.WriteString(s.gameRowSelected.Width(width).Render("▸ " + row))
		} else {
			b.WriteString(s.gameRow.Render("  " + row))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderGameRow renders a single game line: "AWY 105 - 102 HME  Q3 5:42 ● LIVE"
func renderGameRow(g nbalive.Game, s styles) string {
	away := s.teamTricode.Render(g.AwayTeam.TeamTricode)
	home := s.teamTricode.Render(g.HomeTeam.TeamTricode)

	awayScore := s.score.Render(fmt.Sprintf("%d", g.AwayTeam.Score))
	homeScore := s.score.Render(fmt.Sprintf("%d", g.HomeTeam.Score))

	var status string
	switch g.GameStatus {
	case nbalive.GameInProgress:
		clock := formatClock(g.GameClock)
		status = s.liveIndicator.Render() + " " +
			s.dimText.Render(fmt.Sprintf("Q%d %s", g.Period, clock))
	case nbalive.GameFinal:
		status = s.finalIndicator.Render()
		if g.Period > 4 {
			status += s.dimText.Render(fmt.Sprintf(" (OT%d)", g.Period-4))
		}
	default: // Scheduled
		status = s.scheduledTime.Render(g.GameStatusText)
	}

	// Record
	awayRecord := s.dimText.Render(fmt.Sprintf("(%d-%d)", g.AwayTeam.Wins, g.AwayTeam.Losses))
	homeRecord := s.dimText.Render(fmt.Sprintf("(%d-%d)", g.HomeTeam.Wins, g.HomeTeam.Losses))

	return lipgloss.JoinHorizontal(lipgloss.Center,
		away, " ", awayRecord, " ",
		awayScore, s.dimText.Render(" - "), homeScore,
		" ", home, " ", homeRecord,
		"  ", status,
	)
}

// scoreboardRowY returns the Y offset where game rows start in the scoreboard.
// Used for mouse click hit-testing.
const scoreboardHeaderLines = 3 // title + date + blank line

// gameIndexFromY converts a mouse Y coordinate to a game index.
// Returns -1 if the click is outside the game list.
func gameIndexFromY(y int, numGames int) int {
	idx := y - scoreboardHeaderLines
	if idx < 0 || idx >= numGames {
		return -1
	}
	return idx
}

// formatClock formats a nbalive.Duration game clock as "M:SS".
func formatClock(d nbalive.Duration) string {
	total := int(d.Duration.Seconds())
	min := total / 60
	sec := total % 60
	return fmt.Sprintf("%d:%02d", min, sec)
}
