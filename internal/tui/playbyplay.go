package tui

import (
	"fmt"
	"strings"

	"github.com/darin-patton-hpe/nbalive"
)

// renderPlayByPlay renders the play-by-play tab content.
// Returns the full string to be set as viewport content.
func renderPlayByPlay(actions []nbalive.Action, s styles) string {
	if len(actions) == 0 {
		return s.dimText.Render("  No play-by-play data available.")
	}

	var b strings.Builder
	currentPeriod := 0

	// Render in reverse (most recent first).
	for i := len(actions) - 1; i >= 0; i-- {
		a := actions[i]

		// Period header when period changes.
		if a.Period != currentPeriod {
			currentPeriod = a.Period
			header := periodName(currentPeriod)
			b.WriteString("\n")
			b.WriteString(s.pbpPeriod.Render(header))
			b.WriteString("\n")
		}

		b.WriteString(renderAction(a, s))
		b.WriteString("\n")
	}

	return b.String()
}

// renderAction formats a single play-by-play action.
func renderAction(a nbalive.Action, s styles) string {
	clock := s.pbpClock.Render(formatActionClock(a.Clock))
	team := ""
	if a.TeamTricode != "" {
		team = s.teamTricode.Render(a.TeamTricode) + " "
	}

	desc := formatActionDescription(a, s)

	// Score display.
	score := ""
	if a.ScoreHome != "" && a.ScoreAway != "" {
		score = s.pbpScore.Render(fmt.Sprintf(" %s-%s", a.ScoreAway, a.ScoreHome))
	}

	return fmt.Sprintf(" %s %s%s%s", clock, team, desc, score)
}

// formatActionDescription styles the action description based on type.
func formatActionDescription(a nbalive.Action, s styles) string {
	desc := a.Description
	if desc == "" {
		desc = a.ActionType
		if a.SubType != "" {
			desc += " - " + a.SubType
		}
	}

	// Color based on shot result.
	if a.IsFieldGoal == 1 || a.ActionType == "freethrow" {
		if a.IsMade() {
			return s.pbpMade.Render(desc)
		}
		return s.pbpMissed.Render(desc)
	}

	return s.pbpDescription.Render(desc)
}

// formatActionClock formats the play clock as "M:SS".
func formatActionClock(d nbalive.Duration) string {
	total := int(d.Duration.Seconds())
	m := total / 60
	sec := total % 60
	return fmt.Sprintf("%d:%02d", m, sec)
}

// periodName returns a display name for a period number.
func periodName(period int) string {
	if period <= 4 {
		return fmt.Sprintf("Quarter %d", period)
	}
	ot := period - 4
	if ot == 1 {
		return "Overtime"
	}
	return fmt.Sprintf("OT%d", ot)
}
