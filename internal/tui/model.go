package tui

import (
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/darin-patton-hpe/buckets/internal/data"
	"github.com/darin-patton-hpe/nbalive"
)

// route identifies which view is active.
type route int

const (
	routeScoreboard route = iota
	routeGameDetail
)

// Model is the root Bubble Tea model for the buckets application.
type Model struct {
	client data.NBAClient
	route  route
	width  int
	height int
	isDark bool
	s      styles

	// Scoreboard state.
	games        []nbalive.Game
	cursor       int
	sbErr        error
	loading      bool
	spinner      spinner.Model
	selectedDate time.Time // zero value = live/today

	// Game detail sub-model (nil when on scoreboard).
	detail *gameModel

	// Status line text.
	status string
}

// NewModel creates a new root model.
func NewModel(client data.NBAClient) Model {
	s := newStyles(true)
	return Model{
		client:  client,
		route:   routeScoreboard,
		isDark:  true, // default until BackgroundColorMsg arrives
		s:       s,
		loading: true,
		spinner: spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(s.spinner)),
	}
}

func (m Model) isLive() bool {
	return m.selectedDate.IsZero()
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchScoreboardCmd(m.client),
		scoreboardTickCmd(),
		tea.RequestBackgroundColor,
		m.spinner.Tick,
	)
}

// Update handles all incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.BackgroundColorMsg:
		m.isDark = msg.IsDark()
		m.s = newStyles(m.isDark)
		m.spinner.Style = m.s.spinner
		if m.detail != nil {
			m.detail.s = m.s
			m.detail.spinner.Style = m.s.spinner
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.detail != nil {
			return m, m.detail.update(msg)
		}
		return m, nil

	case scoreboardMsg:
		m.loading = false
		if msg.err != nil {
			m.sbErr = msg.err
			return m, nil
		}
		m.games = msg.games
		m.sbErr = nil
		return m, nil

	case scoreboardTickMsg:
		if m.route == routeScoreboard && m.isLive() {
			return m, tea.Batch(
				fetchScoreboardCmd(m.client),
				scoreboardTickCmd(),
			)
		}
		return m, scoreboardTickCmd()

	case spinner.TickMsg:
		var cmds []tea.Cmd
		if m.loading && m.route == routeScoreboard {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.detail != nil && m.detail.loading {
			cmd := m.detail.update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case tea.MouseClickMsg:
		return m.handleMouseClick(msg)
	}

	// Delegate remaining messages to the detail sub-model when active.
	if m.route == routeGameDetail && m.detail != nil {
		cmd := m.detail.update(msg)
		return m, cmd
	}

	return m, nil
}

// handleKey processes keyboard input.
func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys.
	switch key {
	case keyQuit, keyQuitAlt:
		if m.detail != nil {
			m.detail.stopWatch()
		}
		return m, tea.Quit
	}

	switch m.route {
	case routeScoreboard:
		switch key {
		case keyUp, keyUpAlt:
			if m.cursor > 0 {
				m.cursor--
			}
		case keyDown, keyDownAlt:
			if m.cursor < len(m.games)-1 {
				m.cursor++
			}
		case keyLeft, keyLeftAlt:
			return m.navigateDate(-1)
		case keyRight, keyRightAlt:
			return m.navigateDate(1)
		case keyToday:
			if !m.isLive() {
				m.selectedDate = time.Time{}
				m.games = nil
				m.cursor = 0
				m.loading = true
				return m, tea.Batch(fetchScoreboardCmd(m.client), m.spinner.Tick)
			}
		case keyEnter:
			return m.navigateToGame()
		}

	case routeGameDetail:
		switch key {
		case keyEsc:
			return m.navigateToScoreboard()
		default:
			// Forward to game model.
			if m.detail != nil {
				cmd := m.detail.update(msg)
				return m, cmd
			}
		}
	}

	return m, nil
}

// handleMouseClick processes mouse click events.
func (m Model) handleMouseClick(msg tea.MouseClickMsg) (tea.Model, tea.Cmd) {
	if msg.Button != tea.MouseLeft {
		return m, nil
	}

	switch m.route {
	case routeScoreboard:
		idx := gameIndexFromY(msg.Y, len(m.games))
		if idx >= 0 {
			m.cursor = idx
			return m.navigateToGame()
		}

	case routeGameDetail:
		// Tab click detection is handled via OnMouse in View().
		// Forward other clicks to game model.
		if m.detail != nil {
			cmd := m.detail.update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// navigateToGame switches from scoreboard to game detail.
func (m Model) navigateToGame() (tea.Model, tea.Cmd) {
	if m.cursor < 0 || m.cursor >= len(m.games) {
		return m, nil
	}
	game := m.games[m.cursor]
	detail, cmd := newGameModel(m.client, game, m.width, m.height, m.s)
	m.detail = detail
	m.route = routeGameDetail
	return m, cmd
}

func (m Model) navigateToScoreboard() (tea.Model, tea.Cmd) {
	if m.detail != nil {
		m.detail.stopWatch()
	}
	m.detail = nil
	m.route = routeScoreboard

	var fetchCmd tea.Cmd
	if m.isLive() {
		fetchCmd = fetchScoreboardCmd(m.client)
	} else {
		fetchCmd = fetchScoreboardByDateCmd(m.client, m.selectedDate.Format("2006-01-02"))
	}
	return m, tea.Batch(fetchCmd, scoreboardTickCmd())
}

func (m Model) navigateDate(days int) (tea.Model, tea.Cmd) {
	var base time.Time
	if m.isLive() {
		base = time.Now()
	} else {
		base = m.selectedDate
	}

	target := base.AddDate(0, 0, days)
	today := time.Now()
	if target.After(today) {
		if m.isLive() {
			return m, nil
		}
		m.selectedDate = time.Time{}
		m.games = nil
		m.cursor = 0
		m.loading = true
		return m, tea.Batch(fetchScoreboardCmd(m.client), m.spinner.Tick)
	}

	m.selectedDate = target
	m.games = nil
	m.cursor = 0
	m.loading = true
	return m, tea.Batch(
		fetchScoreboardByDateCmd(m.client, target.Format("2006-01-02")),
		m.spinner.Tick,
	)
}

// View renders the current view.
func (m Model) View() tea.View {
	var content string

	switch m.route {
	case routeScoreboard:
		content = renderScoreboard(m.games, m.cursor, m.width, m.s, m.selectedDate, m.loading, m.spinner.View())
		if m.sbErr != nil && len(m.games) == 0 {
			content = m.s.errText.Render("Error: "+m.sbErr.Error()) + "\n"
		}
		content += "\n" + m.s.help.Render(helpScoreboard())

	case routeGameDetail:
		if m.detail != nil {
			content = m.detail.view()
			content += "\n" + m.s.help.Render(helpGame())
		}
	}

	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = "buckets"

	// OnMouse handler for tab click hit-testing in game detail view.
	if m.route == routeGameDetail {
		v.OnMouse = func(msg tea.MouseMsg) tea.Cmd {
			click, ok := msg.(tea.MouseClickMsg)
			if !ok || click.Button != tea.MouseLeft {
				return nil
			}
			// Tab bar is at Y=1 (below game header line).
			// Adjust if header is absent.
			tabY := 1
			if m.detail != nil && m.detail.boxScore == nil && !m.detail.loading {
				tabY = 0
			}
			if click.Y == tabY {
				tab := tabFromX(click.X)
				if tab >= 0 {
					return func() tea.Msg {
						return tabSelectMsg{tab: tab}
					}
				}
			}
			return nil
		}
	}

	return v
}
