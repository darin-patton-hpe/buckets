package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/darin-patton-hpe/nbalive"
)

const dateDisplayFormat = "Mon, Jan 2, 2006"

// skeletonCount is the number of placeholder rows shown while loading.
const skeletonCount = 6

func renderScoreboard(games []nbalive.Game, cursor int, width int, s styles, selectedDate time.Time, loading bool, spinnerView string) string {
	var b strings.Builder
	isLive := selectedDate.IsZero()

	title := "🏀 NBA Scoreboard"
	if loading {
		title += "  " + spinnerView
	}
	b.WriteString(s.title.Render(title))
	b.WriteString("\n")

	if isLive {
		b.WriteString(s.subtitle.Render("Today — " + time.Now().Format(dateDisplayFormat)))
	} else {
		b.WriteString(s.subtitle.Render("◀ " + selectedDate.Format(dateDisplayFormat) + " ▶"))
	}
	b.WriteString("\n\n")

	if loading && len(games) == 0 {
		for range skeletonCount {
			b.WriteString(renderSkeletonRow(s))
			b.WriteString("\n")
		}
		return b.String()
	}

	if len(games) == 0 {
		if isLive {
			b.WriteString(s.dimText.Render("\n  No games today.\n"))
		} else {
			b.WriteString(s.dimText.Render("\n  No games on " + selectedDate.Format(dateDisplayFormat) + ".\n"))
		}
		return b.String()
	}

	for i, g := range games {
		row := renderGameRow(g, s)
		if i == cursor {
			b.WriteString(s.gameRowSelected.Width(width).Render("▸ " + row))
		} else {
			b.WriteString(s.gameRow.Render("  " + row))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderSkeletonRow renders a dim placeholder row that mimics the shape of a game row.
func renderSkeletonRow(s styles) string {
	return s.gameRow.Render("  " + s.dimText.Render("███  (██-██)   ███ - ███   ███  (██-██)    ████████"))
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
