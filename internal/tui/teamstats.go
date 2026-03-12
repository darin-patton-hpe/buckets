package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/darin-patton-hpe/nbalive"
)

// teamStatRow defines a single row in the team stats comparison.
type teamStatRow struct {
	label string
	away  string
	home  string
	awayN float64 // numeric for bar rendering
	homeN float64
}

// renderTeamStats renders a side-by-side team stats comparison.
func renderTeamStats(game *nbalive.BoxScoreGame, width int, s styles) string {
	if game == nil {
		return s.dimText.Render("  Loading team stats...")
	}

	away := game.AwayTeam.Statistics
	home := game.HomeTeam.Statistics

	rows := []teamStatRow{
		statRow("Points", away.Points, home.Points),
		statRow("Field Goals", away.FieldGoalsMade, home.FieldGoalsMade),
		statRowPct("FG%", away.FieldGoalsPercentage, home.FieldGoalsPercentage),
		statRow("3-Pointers", away.ThreePointersMade, home.ThreePointersMade),
		statRowPct("3P%", away.ThreePointersPercentage, home.ThreePointersPercentage),
		statRow("Free Throws", away.FreeThrowsMade, home.FreeThrowsMade),
		statRowPct("FT%", away.FreeThrowsPercentage, home.FreeThrowsPercentage),
		statRow("Rebounds", away.ReboundsTotal, home.ReboundsTotal),
		statRow("Off. Rebounds", away.ReboundsOffensive, home.ReboundsOffensive),
		statRow("Def. Rebounds", away.ReboundsDefensive, home.ReboundsDefensive),
		statRow("Assists", away.Assists, home.Assists),
		statRow("Steals", away.Steals, home.Steals),
		statRow("Blocks", away.Blocks, home.Blocks),
		statRow("Turnovers", away.Turnovers, home.Turnovers),
		statRow("Fouls", away.FoulsPersonal, home.FoulsPersonal),
		statRow("Pts in Paint", away.PointsInThePaint, home.PointsInThePaint),
		statRow("Fast Break Pts", away.PointsFastBreak, home.PointsFastBreak),
		statRow("Bench Pts", away.BenchPoints, home.BenchPoints),
		statRow("Biggest Lead", away.BiggestLead, home.BiggestLead),
		statRow("Lead Changes", away.LeadChanges, home.LeadChanges),
		statRow("Times Tied", away.TimesTied, home.TimesTied),
		statRowPct("eFG%", away.FieldGoalsEffectiveAdjusted, home.FieldGoalsEffectiveAdjusted),
		statRowPct("TS%", away.TrueShootingPercentage, home.TrueShootingPercentage),
	}

	var b strings.Builder

	// Header: team tricodes.
	barWidth := 20
	headerAway := s.teamTricode.Width(8).Align(lipgloss.Right).Render(game.AwayTeam.TeamTricode)
	headerLabel := s.statLabel.Width(22).Align(lipgloss.Center).Render("")
	headerHome := s.teamTricode.Width(8).Render(game.HomeTeam.TeamTricode)
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center,
		headerAway, strings.Repeat(" ", barWidth+2), headerLabel, strings.Repeat(" ", barWidth+2), headerHome,
	))
	b.WriteString("\n\n")

	for _, row := range rows {
		b.WriteString(renderStatRow(row, barWidth, width, s))
		b.WriteString("\n")
	}

	return b.String()
}

// renderStatRow renders a single stat comparison row with visual bars.
func renderStatRow(row teamStatRow, barWidth int, _ int, s styles) string {
	maxVal := max(row.awayN, row.homeN)
	if maxVal == 0 {
		maxVal = 1
	}

	awayBarLen := int(row.awayN / maxVal * float64(barWidth))
	homeBarLen := int(row.homeN / maxVal * float64(barWidth))

	// Right-aligned away bar.
	awayBar := strings.Repeat(" ", barWidth-awayBarLen) + s.statBar.Render(strings.Repeat("█", awayBarLen))
	// Left-aligned home bar.
	homeBar := s.statBar.Render(strings.Repeat("█", homeBarLen)) + strings.Repeat(" ", barWidth-homeBarLen)

	awayVal := s.statValue.Render(row.away)
	homeVal := s.statValue.Width(8).Align(lipgloss.Left).Render(row.home)
	label := s.statLabel.Render(row.label)

	return lipgloss.JoinHorizontal(lipgloss.Center,
		awayVal, " ", awayBar, " ", label, " ", homeBar, " ", homeVal,
	)
}

// statRow creates a teamStatRow from integer values.
func statRow(label string, away, home int) teamStatRow {
	return teamStatRow{
		label: label,
		away:  fmt.Sprintf("%d", away),
		home:  fmt.Sprintf("%d", home),
		awayN: float64(away),
		homeN: float64(home),
	}
}

// statRowPct creates a teamStatRow from percentage values (0-1 scale).
func statRowPct(label string, away, home float64) teamStatRow {
	return teamStatRow{
		label: label,
		away:  fmt.Sprintf("%.1f", away*100),
		home:  fmt.Sprintf("%.1f", home*100),
		awayN: away,
		homeN: home,
	}
}
