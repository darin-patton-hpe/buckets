package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/darin-patton-hpe/nbalive"
)

// boxScoreColumns defines the stat columns shown in the box score table.
// Each entry is (header, width, extractor).
type boxScoreCol struct {
	header string
	width  int
	format func(nbalive.PlayerStats) string
}

func boxScoreCols() []boxScoreCol {
	return []boxScoreCol{
		{"MIN", 6, func(s nbalive.PlayerStats) string { return formatMinutes(s.Minutes) }},
		{"PTS", 5, func(s nbalive.PlayerStats) string { return fmt.Sprintf("%d", s.Points) }},
		{"REB", 5, func(s nbalive.PlayerStats) string { return fmt.Sprintf("%d", s.ReboundsTotal) }},
		{"AST", 5, func(s nbalive.PlayerStats) string { return fmt.Sprintf("%d", s.Assists) }},
		{"STL", 5, func(s nbalive.PlayerStats) string { return fmt.Sprintf("%d", s.Steals) }},
		{"BLK", 5, func(s nbalive.PlayerStats) string { return fmt.Sprintf("%d", s.Blocks) }},
		{"TO", 5, func(s nbalive.PlayerStats) string { return fmt.Sprintf("%d", s.Turnovers) }},
		{"FG", 8, func(s nbalive.PlayerStats) string {
			return fmt.Sprintf("%d-%d", s.FieldGoalsMade, s.FieldGoalsAttempted)
		}},
		{"3P", 8, func(s nbalive.PlayerStats) string {
			return fmt.Sprintf("%d-%d", s.ThreePointersMade, s.ThreePointersAttempted)
		}},
		{"FT", 8, func(s nbalive.PlayerStats) string {
			return fmt.Sprintf("%d-%d", s.FreeThrowsMade, s.FreeThrowsAttempted)
		}},
		{"+/-", 6, func(s nbalive.PlayerStats) string {
			return fmt.Sprintf("%+.0f", s.PlusMinusPoints)
		}},
	}
}

// renderBoxScore renders the box score tab for a game.
func renderBoxScore(game *nbalive.BoxScoreGame, width int, s styles) string {
	if game == nil {
		return s.dimText.Render("  Loading box score...")
	}

	var b strings.Builder
	cols := boxScoreCols()
	isLive := game.GameStatus == nbalive.GameInProgress

	nameWidth := 20
	posWidth := 4
	usedCols := fitColumns(cols, width, nameWidth+posWidth)

	renderTeamTable := func(team *nbalive.BoxTeam) {
		header := s.teamTricode.Render(team.TeamTricode) + " " +
			s.score.Render(fmt.Sprintf("%d", team.Score))
		b.WriteString(header)
		b.WriteString("\n")

		headerLine := s.headerCell.Width(nameWidth).Align(lipgloss.Left).PaddingRight(0).Render("PLAYER")
		headerLine += s.headerCell.Width(posWidth).Align(lipgloss.Left).PaddingRight(0).Render("")
		for _, col := range usedCols {
			headerLine += s.headerCell.Width(col.width).Render(col.header)
		}
		b.WriteString(headerLine)
		b.WriteString("\n")
		b.WriteString(strings.Repeat("─", min(width, len(headerLine)+10)))
		b.WriteString("\n")

		for _, p := range team.Players {
			if !p.Starter.Bool() {
				continue
			}
			b.WriteString(renderPlayerRow(p, usedCols, nameWidth, posWidth, s.starterRow, s, isLive))
			b.WriteString("\n")
		}

		b.WriteString(s.dimText.Render(strings.Repeat("·", min(width, 60))))
		b.WriteString("\n")

		for _, p := range team.Players {
			if p.Starter.Bool() || !p.Played.Bool() {
				continue
			}
			b.WriteString(renderPlayerRow(p, usedCols, nameWidth, posWidth, s.benchRow, s, isLive))
			b.WriteString("\n")
		}
	}

	renderTeamTable(&game.AwayTeam)
	b.WriteString("\n")
	renderTeamTable(&game.HomeTeam)

	return b.String()
}

// renderPlayerRow renders a single player's stats row.
func renderPlayerRow(p nbalive.BoxPlayer, cols []boxScoreCol, nameWidth int, posWidth int, rowStyle lipgloss.Style, s styles, isLive bool) string {
	name := p.NameI
	if len(name) > nameWidth-1 {
		name = name[:nameWidth-1]
	}

	row := rowStyle.Width(nameWidth).Render(name)
	row += s.dimText.Width(posWidth).Render(p.Position)
	for _, col := range cols {
		val := col.format(p.Statistics)
		row += s.dataCell.Width(col.width).Render(val)
	}

	// On-court indicator.
	if isLive && p.OnCourt.Bool() {
		row += " ●"
	}

	return row
}

// fitColumns returns as many columns as fit within the available width.
func fitColumns(cols []boxScoreCol, avail int, nameWidth int) []boxScoreCol {
	remaining := avail - nameWidth
	var result []boxScoreCol
	for _, col := range cols {
		if remaining < col.width {
			break
		}
		result = append(result, col)
		remaining -= col.width
	}
	return result
}

// formatMinutes formats a Duration as "MM:SS" for minutes played.
func formatMinutes(d nbalive.Duration) string {
	total := int(d.Duration.Seconds())
	m := total / 60
	sec := total % 60
	return fmt.Sprintf("%d:%02d", m, sec)
}
