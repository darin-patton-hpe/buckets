package tui

import (
	"context"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/darin-patton-hpe/buckets/internal/data"
	"github.com/darin-patton-hpe/nbalive"
)

// isNotAvailable reports whether err indicates an HTTP 403 (data not yet on CDN).
func isNotAvailable(err error) bool {
	return err != nil && strings.Contains(err.Error(), "status 403")
}

// gameModel holds all state for the game detail view.
type gameModel struct {
	client     data.NBAClient
	gameID     string
	gameStatus nbalive.GameStatus
	game       nbalive.Game
	s          styles

	// Current tab.
	activeTab int

	// Box score data.
	boxScore *nbalive.BoxScoreGame

	// Play-by-play data.
	actions []nbalive.Action

	// Viewport for scrollable content (play-by-play, box score, team stats).
	viewport viewport.Model

	// Watch state.
	watchCtx    context.Context
	watchCancel context.CancelFunc
	watchCh     <-chan nbalive.Event

	// Dimensions.
	width  int
	height int

	// Loading / error state.
	loading bool
	err     error
	spinner spinner.Model
}

// newGameModel creates a new game detail model and starts data fetches.
func newGameModel(client data.NBAClient, game nbalive.Game, width, height int, s styles) (*gameModel, tea.Cmd) {
	vpHeight := height - 5
	if vpHeight < 1 {
		vpHeight = 1
	}

	gm := &gameModel{
		client:     client,
		gameID:     game.GameID,
		gameStatus: game.GameStatus,
		game:       game,
		s:          s,
		activeTab:  tabBoxScore,
		viewport:   viewport.New(viewport.WithWidth(width), viewport.WithHeight(vpHeight)),
		width:      width,
		height:     height,
		loading:    game.GameStatus != nbalive.GameScheduled,
		spinner:    spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(s.spinner)),
	}

	if game.GameStatus == nbalive.GameScheduled {
		return gm, nil
	}

	cmds := []tea.Cmd{
		fetchBoxScoreCmd(client, game.GameID),
		fetchPlayByPlayCmd(client, game.GameID),
		gm.spinner.Tick,
	}

	return gm, tea.Batch(cmds...)
}

// startWatch starts watching a live game for real-time updates.
func (gm *gameModel) startWatch() tea.Cmd {
	if gm.watchCancel != nil {
		gm.watchCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	gm.watchCtx = ctx
	gm.watchCancel = cancel
	gm.watchCh = gm.client.Watch(ctx, gm.gameID, nbalive.WatchConfig{
		BoxScore: true,
	})
	return waitWatchCmd(gm.watchCh)
}

// stopWatch cancels any active watcher.
func (gm *gameModel) stopWatch() {
	if gm.watchCancel != nil {
		gm.watchCancel()
		gm.watchCancel = nil
		gm.watchCtx = nil
		gm.watchCh = nil
	}
}

// update handles messages for the game detail view.
func (gm *gameModel) update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case boxScoreMsg:
		gm.loading = false
		if msg.err != nil {
			if !isNotAvailable(msg.err) {
				gm.err = msg.err
			}
			return nil
		}
		gm.boxScore = msg.game
		gm.updateViewportContent()

		// Start watching if game is in progress.
		if gm.boxScore != nil && gm.boxScore.GameStatus == nbalive.GameInProgress {
			cmds = append(cmds, gm.startWatch())
		}

	case playByPlayMsg:
		if msg.err != nil {
			if !isNotAvailable(msg.err) {
				gm.err = msg.err
			}
			return nil
		}
		gm.actions = msg.actions
		gm.updateViewportContent()

	case watchEventMsg:
		evt := msg.event
		switch evt.Kind {
		case nbalive.EventBoxScore:
			gm.boxScore = evt.BoxScore
			gm.updateViewportContent()
		case nbalive.EventAction:
			if evt.Action != nil {
				gm.actions = append(gm.actions, *evt.Action)
				gm.updateViewportContent()
			}
		case nbalive.EventGameOver:
			gm.boxScore = evt.BoxScore
			gm.updateViewportContent()
			gm.stopWatch()
			return nil
		case nbalive.EventError:
			// Transient — keep watching.
		}
		// Re-listen on the channel.
		if gm.watchCh != nil {
			cmds = append(cmds, waitWatchCmd(gm.watchCh))
		}

	case watchClosedMsg:
		gm.stopWatch()

	case spinner.TickMsg:
		if gm.loading {
			var cmd tea.Cmd
			gm.spinner, cmd = gm.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tabSelectMsg:
		gm.activeTab = msg.tab
		gm.updateViewportContent()

	case tea.KeyPressMsg:
		switch msg.String() {
		case keyOne:
			gm.activeTab = tabBoxScore
			gm.updateViewportContent()
		case keyTwo:
			gm.activeTab = tabPlayByPlay
			gm.updateViewportContent()
		case keyThree:
			gm.activeTab = tabTeamStats
			gm.updateViewportContent()
		case keyTab:
			gm.activeTab = (gm.activeTab + 1) % tabCount
			gm.updateViewportContent()
		case keyShiftTab:
			gm.activeTab = (gm.activeTab - 1 + tabCount) % tabCount
			gm.updateViewportContent()
		default:
			// Forward to viewport for scrolling.
			var cmd tea.Cmd
			gm.viewport, cmd = gm.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.MouseWheelMsg:
		var cmd tea.Cmd
		gm.viewport, cmd = gm.viewport.Update(msg)
		cmds = append(cmds, cmd)

	case tea.WindowSizeMsg:
		gm.width = msg.Width
		gm.height = msg.Height
		vpHeight := msg.Height - 5
		if vpHeight < 1 {
			vpHeight = 1
		}
		gm.viewport.SetWidth(msg.Width)
		gm.viewport.SetHeight(vpHeight)
		gm.updateViewportContent()
	}

	return tea.Batch(cmds...)
}

// updateViewportContent sets the viewport content based on the active tab.
func (gm *gameModel) updateViewportContent() {
	if gm.gameStatus == nbalive.GameScheduled && gm.boxScore == nil {
		gm.viewport.SetContent("")
		return
	}

	var content string
	switch gm.activeTab {
	case tabBoxScore:
		content = renderBoxScore(gm.boxScore, gm.width, gm.s)
	case tabPlayByPlay:
		content = renderPlayByPlay(gm.actions, gm.s)
	case tabTeamStats:
		content = renderTeamStats(gm.boxScore, gm.width, gm.s)
	}
	gm.viewport.SetContent(content)
}

// view renders the game detail view.
func (gm *gameModel) view() string {
	var b strings.Builder

	// Game header.
	if gm.boxScore != nil {
		bs := gm.boxScore
		b.WriteString(gm.s.title.Render(
			bs.AwayTeam.TeamTricode + " " +
				formatInt(bs.AwayTeam.Score) + " - " +
				formatInt(bs.HomeTeam.Score) + " " +
				bs.HomeTeam.TeamTricode,
		))

		// Status.
		switch bs.GameStatus {
		case nbalive.GameInProgress:
			b.WriteString("  ")
			b.WriteString(gm.s.liveIndicator.Render())
			b.WriteString(" ")
			b.WriteString(gm.s.dimText.Render(
				"Q" + formatInt(bs.Period) + " " + formatClock(bs.GameClock),
			))
		case nbalive.GameFinal:
			b.WriteString("  ")
			b.WriteString(gm.s.finalIndicator.Render())
		}
		b.WriteString("\n\n")
	} else if gm.loading {
		b.WriteString("  " + gm.spinner.View() + " Loading...")
		b.WriteString("\n\n")
	} else if gm.gameStatus == nbalive.GameScheduled {
		b.WriteString("  " + renderGameRow(gm.game, gm.s))
		b.WriteString("\n\n")
	}

	// Tab bar.
	b.WriteString(renderTabBar(gm.activeTab, gm.width, gm.s))
	b.WriteString("\n")

	// Error.
	if gm.err != nil {
		b.WriteString(gm.s.errText.Render("Error: " + gm.err.Error()))
		b.WriteString("\n")
	}

	// Viewport content.
	b.WriteString(gm.viewport.View())

	return b.String()
}

// renderTabBar renders the tab navigation bar.
func renderTabBar(active int, _ int, s styles) string {
	var tabs []string
	for i, name := range tabNames {
		label := name
		if i == active {
			tabs = append(tabs, s.activeTab.Render(label))
		} else {
			tabs = append(tabs, s.inactiveTab.Render(label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

// tabFromX returns which tab a mouse X coordinate corresponds to, or -1.
func tabFromX(x int) int {
	cumX := 0
	for i, name := range tabNames {
		w := len(name) + 2 // padding
		if x >= cumX && x < cumX+w {
			return i
		}
		cumX += w
	}
	return -1
}

// formatInt formats an int for display.
func formatInt(n int) string {
	return strconv.Itoa(n)
}
