package tui

import (
	"time"

	"github.com/darin-patton-hpe/nbalive"
)

// scoreboardMsg carries the result of a scoreboard fetch.
type scoreboardMsg struct {
	games []nbalive.Game
	err   error
}

// boxScoreMsg carries the result of a box score fetch.
type boxScoreMsg struct {
	game *nbalive.BoxScoreGame
	err  error
}

// playByPlayMsg carries the result of a play-by-play fetch.
type playByPlayMsg struct {
	actions []nbalive.Action
	err     error
}

// watchEventMsg wraps a single event from the Watch channel.
type watchEventMsg struct {
	event nbalive.Event
}

// watchClosedMsg signals that the Watch channel has closed.
type watchClosedMsg struct{}

// errMsg is a generic error message.
type errMsg struct {
	err error
}

// tabSelectMsg requests switching to a specific tab.
type tabSelectMsg struct {
	tab int
}

// scoreboardTickMsg triggers a scoreboard refresh.
type scoreboardTickMsg time.Time
