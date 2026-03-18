package tui

// Key binding constants used throughout the application.
const (
	keyQuit     = "q"
	keyQuitAlt  = "ctrl+c"
	keyUp       = "up"
	keyDown     = "down"
	keyUpAlt    = "k"
	keyDownAlt  = "j"
	keyEnter    = "enter"
	keyEsc      = "esc"
	keyTab      = "tab"
	keyShiftTab = "shift+tab"
	keyOne      = "1"
	keyTwo      = "2"
	keyThree    = "3"
	keyQuestion = "?"
	keyLeft     = "left"
	keyRight    = "right"
	keyLeftAlt  = "h"
	keyRightAlt = "l"
	keyToday    = "t"
)

// Tab indices.
const (
	tabBoxScore = iota
	tabPlayByPlay
	tabTeamStats
	tabCount
)

// Tab names for display.
var tabNames = [tabCount]string{
	"Box Score",
	"Play-by-Play",
	"Team Stats",
}

// helpScoreboard returns help text for the scoreboard view.
func helpScoreboard() string {
	return " ↑/k up • ↓/j down • ←/h prev day • →/l next day • t today • enter select • q quit"
}

// helpGame returns help text for the game detail view.
func helpGame() string {
	return " 1/2/3 tabs • tab/shift+tab cycle • ↑/k ↓/j scroll • esc back • q quit"
}
