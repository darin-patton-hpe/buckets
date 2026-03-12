package tui

import (
	"context"
	"errors"
	"image/color"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/darin-patton-hpe/buckets/internal/data"
	"github.com/darin-patton-hpe/nbalive"
)

func testGame() nbalive.Game {
	return nbalive.Game{
		GameID:         "0022400001",
		GameStatus:     nbalive.GameFinal,
		GameStatusText: "Final",
		AwayTeam:       nbalive.GameTeam{TeamTricode: "BOS", Score: 108},
		HomeTeam:       nbalive.GameTeam{TeamTricode: "LAL", Score: 102},
	}
}

func testBoxScoreGame() *nbalive.BoxScoreGame {
	return &nbalive.BoxScoreGame{
		GameID:     "0022400001",
		GameStatus: nbalive.GameFinal,
		AwayTeam:   nbalive.BoxTeam{TeamTricode: "BOS", Score: 108},
		HomeTeam:   nbalive.BoxTeam{TeamTricode: "LAL", Score: 102},
	}
}

func testActions() []nbalive.Action {
	return []nbalive.Action{
		{Period: 1, Description: "Made jumper", IsFieldGoal: 1, ShotResult: "Made", TeamTricode: "BOS"},
		{Period: 1, Description: "Missed three", IsFieldGoal: 1, ShotResult: "Missed", TeamTricode: "LAL"},
	}
}

func testMockClient() *data.MockClient {
	game := testGame()
	box := testBoxScoreGame()
	actions := testActions()

	return &data.MockClient{
		ScoreboardFunc: func(_ context.Context) (*nbalive.ScoreboardResponse, error) {
			return &nbalive.ScoreboardResponse{Scoreboard: nbalive.Scoreboard{Games: []nbalive.Game{game}}}, nil
		},
		BoxScoreFunc: func(_ context.Context, _ string) (*nbalive.BoxScoreResponse, error) {
			return &nbalive.BoxScoreResponse{Game: *box}, nil
		},
		PlayByPlayFunc: func(_ context.Context, _ string) (*nbalive.PlayByPlayResponse, error) {
			return &nbalive.PlayByPlayResponse{Game: nbalive.PlayByPlayGame{Actions: actions}}, nil
		},
	}
}

func keyRune(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Text: string(r), Code: r}
}

func keyCode(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}

func keyCtrlRune(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Mod: tea.ModCtrl}
}

func asModel(t *testing.T, tm tea.Model) Model {
	t.Helper()
	m, ok := tm.(Model)
	if !ok {
		t.Fatalf("model type = %T, want tui.Model", tm)
	}
	return m
}

func TestNewModel(t *testing.T) {
	m := NewModel(testMockClient())

	if m.route != routeScoreboard {
		t.Fatalf("route = %v, want %v", m.route, routeScoreboard)
	}
	if !m.isDark {
		t.Fatalf("isDark = %v, want true", m.isDark)
	}
	if !m.loading {
		t.Fatalf("loading = %v, want true", m.loading)
	}
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", m.cursor)
	}
	if m.detail != nil {
		t.Fatalf("detail = %#v, want nil", m.detail)
	}
}

func TestModelInitReturnsBatchCmd(t *testing.T) {
	m := NewModel(testMockClient())
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init cmd is nil, want non-nil")
	}

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("Init cmd message = %T, want tea.BatchMsg", msg)
	}
	if len(batch) != 4 {
		t.Fatalf("batch size = %d, want 4", len(batch))
	}
}

func TestUpdateScoreboardMsg(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := NewModel(testMockClient())
		m.loading = true
		games := []nbalive.Game{testGame()}

		next, cmd := m.Update(scoreboardMsg{games: games})
		if cmd != nil {
			t.Fatalf("cmd = %v, want nil", cmd)
		}

		got := asModel(t, next)
		if got.loading {
			t.Fatal("loading = true, want false")
		}
		if got.sbErr != nil {
			t.Fatalf("sbErr = %v, want nil", got.sbErr)
		}
		if len(got.games) != 1 || got.games[0].GameID != games[0].GameID {
			t.Fatalf("games = %#v, want one game %q", got.games, games[0].GameID)
		}
	})

	t.Run("error", func(t *testing.T) {
		m := NewModel(testMockClient())
		m.loading = true
		m.games = []nbalive.Game{testGame()}
		origID := m.games[0].GameID

		next, cmd := m.Update(scoreboardMsg{err: errors.New("boom")})
		if cmd != nil {
			t.Fatalf("cmd = %v, want nil", cmd)
		}

		got := asModel(t, next)
		if got.loading {
			t.Fatal("loading = true, want false")
		}
		if got.sbErr == nil {
			t.Fatal("sbErr = nil, want error")
		}
		if len(got.games) != 1 || got.games[0].GameID != origID {
			t.Fatalf("games changed on error: %#v", got.games)
		}
	})
}

func TestUpdateBackgroundColorMsg(t *testing.T) {
	m := NewModel(testMockClient())
	next, _ := m.Update(tea.BackgroundColorMsg{Color: color.White})
	got := asModel(t, next)
	if got.isDark {
		t.Fatalf("isDark = %v, want false for white background", got.isDark)
	}
}

func TestUpdateWindowSizeMsg(t *testing.T) {
	t.Run("updates root dimensions", func(t *testing.T) {
		m := NewModel(testMockClient())
		next, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 42})
		if cmd != nil {
			t.Fatalf("cmd = %v, want nil", cmd)
		}

		got := asModel(t, next)
		if got.width != 120 || got.height != 42 {
			t.Fatalf("size = %dx%d, want 120x42", got.width, got.height)
		}
	})

	t.Run("passes through to detail model", func(t *testing.T) {
		m := NewModel(testMockClient())
		detail, _ := newGameModel(testMockClient(), testGame(), 80, 24, m.s)
		m.route = routeGameDetail
		m.detail = detail

		next, _ := m.Update(tea.WindowSizeMsg{Width: 101, Height: 33})
		got := asModel(t, next)
		if got.detail == nil {
			t.Fatal("detail = nil, want non-nil")
		}
		if got.detail.width != 101 || got.detail.height != 33 {
			t.Fatalf("detail size = %dx%d, want 101x33", got.detail.width, got.detail.height)
		}
	})
}

func TestUpdateScoreboardTickMsg(t *testing.T) {
	t.Run("scoreboard route returns fetch + tick batch", func(t *testing.T) {
		m := NewModel(testMockClient())
		m.route = routeScoreboard

		next, cmd := m.Update(scoreboardTickMsg(time.Now()))
		_ = asModel(t, next)
		if cmd == nil {
			t.Fatal("cmd = nil, want non-nil")
		}
		msg := cmd()
		batch, ok := msg.(tea.BatchMsg)
		if !ok {
			t.Fatalf("tick cmd msg = %T, want tea.BatchMsg", msg)
		}
		if len(batch) != 2 {
			t.Fatalf("batch size = %d, want 2", len(batch))
		}
	})

	t.Run("game detail route returns tick cmd", func(t *testing.T) {
		m := NewModel(testMockClient())
		m.route = routeGameDetail

		_, cmd := m.Update(scoreboardTickMsg(time.Now()))
		if cmd == nil {
			t.Fatal("cmd = nil, want non-nil tick command")
		}
	})
}

func TestHandleKeyScoreboardNavigation(t *testing.T) {
	m := NewModel(testMockClient())
	m.route = routeScoreboard
	m.games = []nbalive.Game{testGame(), {GameID: "0022400002"}}

	modelAfterJ, _ := m.handleKey(keyRune('j'))
	mj := asModel(t, modelAfterJ)
	if mj.cursor != 1 {
		t.Fatalf("cursor after j = %d, want 1", mj.cursor)
	}
	modelAfterDown, _ := mj.handleKey(keyCode(tea.KeyDown))
	md := asModel(t, modelAfterDown)
	if md.cursor != 1 {
		t.Fatalf("cursor after down at end = %d, want 1", md.cursor)
	}

	modelAfterK, _ := md.handleKey(keyRune('k'))
	mk := asModel(t, modelAfterK)
	if mk.cursor != 0 {
		t.Fatalf("cursor after k = %d, want 0", mk.cursor)
	}
	modelAfterUp, _ := mk.handleKey(keyCode(tea.KeyUp))
	mu := asModel(t, modelAfterUp)
	if mu.cursor != 0 {
		t.Fatalf("cursor after up at start = %d, want 0", mu.cursor)
	}

	_, qCmd := mu.handleKey(keyRune('q'))
	if qCmd == nil {
		t.Fatal("q cmd = nil, want tea.Quit")
	}
	if _, ok := qCmd().(tea.QuitMsg); !ok {
		t.Fatalf("q cmd msg = %T, want tea.QuitMsg", qCmd())
	}

	_, cCmd := mu.handleKey(keyCtrlRune('c'))
	if cCmd == nil {
		t.Fatal("ctrl+c cmd = nil, want tea.Quit")
	}
	if _, ok := cCmd().(tea.QuitMsg); !ok {
		t.Fatalf("ctrl+c cmd msg = %T, want tea.QuitMsg", cCmd())
	}

	mu.cursor = 0
	entered, enterCmd := mu.handleKey(keyCode(tea.KeyEnter))
	me := asModel(t, entered)
	if me.route != routeGameDetail {
		t.Fatalf("route after enter = %v, want %v", me.route, routeGameDetail)
	}
	if me.detail == nil {
		t.Fatal("detail = nil after enter, want non-nil")
	}
	if enterCmd == nil {
		t.Fatal("enter cmd = nil, want non-nil initial detail fetch batch")
	}
}

func TestHandleKeyGameDetail(t *testing.T) {
	m := NewModel(testMockClient())
	detail, _ := newGameModel(testMockClient(), testGame(), 100, 30, m.s)
	m.route = routeGameDetail
	m.detail = detail

	next1, _ := m.handleKey(keyRune('1'))
	m1 := asModel(t, next1)
	if m1.detail.activeTab != tabBoxScore {
		t.Fatalf("activeTab after 1 = %d, want %d", m1.detail.activeTab, tabBoxScore)
	}

	next2, _ := m1.handleKey(keyRune('2'))
	m2 := asModel(t, next2)
	if m2.detail.activeTab != tabPlayByPlay {
		t.Fatalf("activeTab after 2 = %d, want %d", m2.detail.activeTab, tabPlayByPlay)
	}

	next3, _ := m2.handleKey(keyRune('3'))
	m3 := asModel(t, next3)
	if m3.detail.activeTab != tabTeamStats {
		t.Fatalf("activeTab after 3 = %d, want %d", m3.detail.activeTab, tabTeamStats)
	}

	back, backCmd := m3.handleKey(tea.KeyPressMsg{Code: tea.KeyEscape})
	mb := asModel(t, back)
	if mb.route != routeScoreboard {
		t.Fatalf("route after escape = %v, want %v", mb.route, routeScoreboard)
	}
	if mb.detail != nil {
		t.Fatalf("detail after escape = %#v, want nil", mb.detail)
	}
	if backCmd == nil {
		t.Fatal("escape cmd = nil, want fetch+tick batch")
	}
}

func TestNavigateToGame(t *testing.T) {
	t.Run("cursor in range", func(t *testing.T) {
		m := NewModel(testMockClient())
		m.games = []nbalive.Game{testGame()}
		m.cursor = 0

		next, cmd := m.navigateToGame()
		got := asModel(t, next)
		if got.route != routeGameDetail {
			t.Fatalf("route = %v, want %v", got.route, routeGameDetail)
		}
		if got.detail == nil {
			t.Fatal("detail = nil, want non-nil")
		}
		if cmd == nil {
			t.Fatal("cmd = nil, want non-nil")
		}
	})

	t.Run("cursor out of range", func(t *testing.T) {
		m := NewModel(testMockClient())
		m.games = []nbalive.Game{testGame()}
		m.cursor = 10

		next, cmd := m.navigateToGame()
		got := asModel(t, next)
		if got.route != routeScoreboard {
			t.Fatalf("route = %v, want %v", got.route, routeScoreboard)
		}
		if got.detail != nil {
			t.Fatalf("detail = %#v, want nil", got.detail)
		}
		if cmd != nil {
			t.Fatalf("cmd = %v, want nil", cmd)
		}
	})
}

func TestNavigateToScoreboard(t *testing.T) {
	m := NewModel(testMockClient())
	detail, _ := newGameModel(testMockClient(), testGame(), 80, 24, m.s)
	m.route = routeGameDetail
	m.detail = detail

	next, cmd := m.navigateToScoreboard()
	got := asModel(t, next)

	if got.route != routeScoreboard {
		t.Fatalf("route = %v, want %v", got.route, routeScoreboard)
	}
	if got.detail != nil {
		t.Fatalf("detail = %#v, want nil", got.detail)
	}
	if cmd == nil {
		t.Fatal("cmd = nil, want non-nil")
	}

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("cmd msg = %T, want tea.BatchMsg", msg)
	}
	if len(batch) != 2 {
		t.Fatalf("batch size = %d, want 2", len(batch))
	}
}

func TestModelView(t *testing.T) {
	t.Run("scoreboard with games contains tricodes", func(t *testing.T) {
		m := NewModel(testMockClient())
		m.games = []nbalive.Game{testGame()}
		m.loading = false

		v := m.View()
		if !strings.Contains(v.Content, "BOS") || !strings.Contains(v.Content, "LAL") {
			t.Fatalf("scoreboard view missing tricodes: %q", v.Content)
		}
		if !v.AltScreen {
			t.Fatal("AltScreen = false, want true")
		}
		if v.MouseMode != tea.MouseModeCellMotion {
			t.Fatalf("MouseMode = %v, want %v", v.MouseMode, tea.MouseModeCellMotion)
		}
		if v.WindowTitle != "buckets" {
			t.Fatalf("WindowTitle = %q, want %q", v.WindowTitle, "buckets")
		}
	})

	t.Run("scoreboard loading contains loading text", func(t *testing.T) {
		m := NewModel(testMockClient())
		m.games = nil
		m.loading = true

		v := m.View()
		if !strings.Contains(v.Content, "Loading scoreboard...") {
			t.Fatalf("loading view text not found: %q", v.Content)
		}
	})

	t.Run("scoreboard error contains error text", func(t *testing.T) {
		m := NewModel(testMockClient())
		m.games = nil
		m.loading = false
		m.sbErr = errors.New("failed fetch")

		v := m.View()
		if !strings.Contains(v.Content, "Error:") {
			t.Fatalf("error view text not found: %q", v.Content)
		}
	})

	t.Run("game detail contains tabs and has OnMouse handler", func(t *testing.T) {
		m := NewModel(testMockClient())
		detail, _ := newGameModel(testMockClient(), testGame(), 100, 30, m.s)
		m.route = routeGameDetail
		m.detail = detail

		v := m.View()
		for _, name := range tabNames {
			if !strings.Contains(v.Content, name) {
				t.Fatalf("game detail view missing tab %q: %q", name, v.Content)
			}
		}
		if v.OnMouse == nil {
			t.Fatal("OnMouse = nil, want non-nil in game detail")
		}
	})
}

func TestGameModelUpdateBoxScoreMsg(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gm := &gameModel{
			activeTab: tabBoxScore,
			s:         newStyles(true),
			width:     100,
			height:    30,
			viewport:  viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
			loading:   true,
		}

		gm.update(boxScoreMsg{game: testBoxScoreGame()})

		if gm.loading {
			t.Fatal("loading = true, want false")
		}
		if gm.boxScore == nil {
			t.Fatal("boxScore = nil, want non-nil")
		}
		if gm.err != nil {
			t.Fatalf("err = %v, want nil", gm.err)
		}
	})

	t.Run("error", func(t *testing.T) {
		gm := &gameModel{loading: true}
		gm.update(boxScoreMsg{err: errors.New("box error")})

		if gm.loading {
			t.Fatal("loading = true, want false")
		}
		if gm.err == nil {
			t.Fatal("err = nil, want non-nil")
		}
	})
}

func TestGameModelUpdatePlayByPlayMsg(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gm := &gameModel{
			activeTab: tabPlayByPlay,
			s:         newStyles(true),
			width:     100,
			height:    30,
			viewport:  viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
		}

		actions := testActions()
		gm.update(playByPlayMsg{actions: actions})
		if len(gm.actions) != len(actions) {
			t.Fatalf("actions len = %d, want %d", len(gm.actions), len(actions))
		}
		if gm.err != nil {
			t.Fatalf("err = %v, want nil", gm.err)
		}
	})

	t.Run("error", func(t *testing.T) {
		gm := &gameModel{}
		gm.update(playByPlayMsg{err: errors.New("pbp error")})
		if gm.err == nil {
			t.Fatal("err = nil, want non-nil")
		}
	})
}

func TestGameModelUpdateTabSelectMsg(t *testing.T) {
	gm := &gameModel{
		activeTab: tabBoxScore,
		s:         newStyles(true),
		width:     100,
		height:    30,
		viewport:  viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
	}

	gm.update(tabSelectMsg{tab: tabTeamStats})
	if gm.activeTab != tabTeamStats {
		t.Fatalf("activeTab = %d, want %d", gm.activeTab, tabTeamStats)
	}
}

func TestGameModelUpdateKeyPressTabKeys(t *testing.T) {
	gm := &gameModel{
		activeTab: tabBoxScore,
		s:         newStyles(true),
		width:     100,
		height:    30,
		viewport:  viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
	}

	gm.update(keyRune('1'))
	if gm.activeTab != tabBoxScore {
		t.Fatalf("activeTab after 1 = %d, want %d", gm.activeTab, tabBoxScore)
	}

	gm.update(keyRune('2'))
	if gm.activeTab != tabPlayByPlay {
		t.Fatalf("activeTab after 2 = %d, want %d", gm.activeTab, tabPlayByPlay)
	}

	gm.update(keyRune('3'))
	if gm.activeTab != tabTeamStats {
		t.Fatalf("activeTab after 3 = %d, want %d", gm.activeTab, tabTeamStats)
	}

	gm.update(keyCode(tea.KeyTab))
	if gm.activeTab != tabBoxScore {
		t.Fatalf("activeTab after tab = %d, want %d", gm.activeTab, tabBoxScore)
	}

	gm.update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if gm.activeTab != tabTeamStats {
		t.Fatalf("activeTab after shift+tab = %d, want %d", gm.activeTab, tabTeamStats)
	}
}

func TestGameModelView(t *testing.T) {
	t.Run("with box score data shows tricodes and scores", func(t *testing.T) {
		gm := &gameModel{
			boxScore:  testBoxScoreGame(),
			activeTab: tabBoxScore,
			s:         newStyles(true),
			width:     100,
			height:    30,
			viewport:  viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
		}
		gm.updateViewportContent()

		out := gm.view()
		for _, want := range []string{"BOS", "LAL", "108", "102"} {
			if !strings.Contains(out, want) {
				t.Fatalf("view missing %q: %q", want, out)
			}
		}
	})

	t.Run("loading without data shows loading text", func(t *testing.T) {
		s := newStyles(true)
		gm := &gameModel{
			loading:   true,
			activeTab: tabBoxScore,
			s:         s,
			width:     100,
			height:    30,
			viewport:  viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
			spinner:   spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(s.spinner)),
		}

		out := gm.view()
		if !strings.Contains(out, "Loading...") {
			t.Fatalf("loading text not found: %q", out)
		}
	})
}

func TestUpdateSpinnerTickMsg(t *testing.T) {
	t.Run("scoreboard loading forwards to spinner", func(t *testing.T) {
		m := NewModel(testMockClient())
		m.loading = true
		m.route = routeScoreboard

		tick := m.spinner.Tick()
		next, cmd := m.Update(tick)
		got := asModel(t, next)
		if cmd == nil {
			t.Fatal("cmd = nil, want non-nil spinner tick continuation")
		}
		if got.spinner.ID() != m.spinner.ID() {
			t.Fatalf("spinner ID changed: %d -> %d", m.spinner.ID(), got.spinner.ID())
		}
	})

	t.Run("not loading returns no cmd", func(t *testing.T) {
		m := NewModel(testMockClient())
		m.loading = false
		m.route = routeScoreboard

		tick := m.spinner.Tick()
		_, cmd := m.Update(tick)
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(tea.BatchMsg); ok {
				return
			}
			if msg != nil {
				t.Fatalf("cmd = non-nil, want nil or empty batch")
			}
		}
	})
}

func TestNewGameModelScheduledSkipsFetch(t *testing.T) {
	gm, cmd := newGameModel(testMockClient(), nbalive.Game{GameID: "0022400099", GameStatus: nbalive.GameScheduled}, 100, 30, newStyles(true))
	if cmd != nil {
		t.Fatal("cmd = non-nil, want nil for scheduled game")
	}
	if gm.loading {
		t.Fatal("loading = true, want false for scheduled game")
	}
	if gm.gameStatus != nbalive.GameScheduled {
		t.Fatalf("gameStatus = %d, want %d", gm.gameStatus, nbalive.GameScheduled)
	}
}

func TestNewGameModelFinalFetches(t *testing.T) {
	_, cmd := newGameModel(testMockClient(), testGame(), 100, 30, newStyles(true))
	if cmd == nil {
		t.Fatal("cmd = nil, want non-nil fetch batch for final game")
	}
}

func TestGameModelViewScheduled(t *testing.T) {
	game := nbalive.Game{
		GameID:     "0022400099",
		GameStatus: nbalive.GameScheduled,
		AwayTeam:   nbalive.GameTeam{TeamTricode: "NYK"},
		HomeTeam:   nbalive.GameTeam{TeamTricode: "BKN"},
	}
	gm := &gameModel{
		game:       game,
		gameID:     game.GameID,
		gameStatus: nbalive.GameScheduled,
		loading:    false,
		activeTab:  tabBoxScore,
		s:          newStyles(true),
		width:      100,
		height:     30,
		viewport:   viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
	}

	out := gm.view()
	if !strings.Contains(out, "NYK") {
		t.Fatalf("scheduled game view missing away team 'NYK': %q", out)
	}
	if !strings.Contains(out, "BKN") {
		t.Fatalf("scheduled game view missing home team 'BKN': %q", out)
	}
}

func TestIsNotAvailable(t *testing.T) {
	t.Run("403 error", func(t *testing.T) {
		err := errors.New("nbalive: GET playbyplay/playbyplay_0022500953.json: status 403")
		if !isNotAvailable(err) {
			t.Fatal("isNotAvailable = false, want true for 403")
		}
	})

	t.Run("other error", func(t *testing.T) {
		err := errors.New("nbalive: connection refused")
		if isNotAvailable(err) {
			t.Fatal("isNotAvailable = true, want false for non-403")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		if isNotAvailable(nil) {
			t.Fatal("isNotAvailable = true, want false for nil")
		}
	})
}

func TestGameModelBoxScoreMsg403Suppressed(t *testing.T) {
	gm := &gameModel{
		loading:   true,
		activeTab: tabBoxScore,
		s:         newStyles(true),
		width:     100,
		height:    30,
		viewport:  viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
	}

	gm.update(boxScoreMsg{err: errors.New("nbalive: GET boxscore/boxscore_0022500953.json: status 403")})

	if gm.loading {
		t.Fatal("loading = true, want false")
	}
	if gm.err != nil {
		t.Fatalf("err = %v, want nil (403 should be suppressed)", gm.err)
	}
}

func TestGameModelPlayByPlayMsg403Suppressed(t *testing.T) {
	gm := &gameModel{
		activeTab: tabPlayByPlay,
		s:         newStyles(true),
		width:     100,
		height:    30,
		viewport:  viewport.New(viewport.WithWidth(100), viewport.WithHeight(10)),
	}

	gm.update(playByPlayMsg{err: errors.New("nbalive: GET playbyplay/playbyplay_0022500953.json: status 403")})

	if gm.err != nil {
		t.Fatalf("err = %v, want nil (403 should be suppressed)", gm.err)
	}
}

func TestGameModelBoxScoreNon403StillErrors(t *testing.T) {
	gm := &gameModel{loading: true}
	gm.update(boxScoreMsg{err: errors.New("nbalive: connection refused")})

	if gm.err == nil {
		t.Fatal("err = nil, want non-nil for non-403 error")
	}
}

func TestGameModelPlayByPlayNon403StillErrors(t *testing.T) {
	gm := &gameModel{}
	gm.update(playByPlayMsg{err: errors.New("nbalive: connection refused")})

	if gm.err == nil {
		t.Fatal("err = nil, want non-nil for non-403 error")
	}
}

func TestNavigateToScheduledGame(t *testing.T) {
	m := NewModel(testMockClient())
	m.games = []nbalive.Game{{
		GameID:     "0022500099",
		GameStatus: nbalive.GameScheduled,
	}}
	m.cursor = 0

	next, cmd := m.navigateToGame()
	got := asModel(t, next)

	if got.route != routeGameDetail {
		t.Fatalf("route = %v, want %v", got.route, routeGameDetail)
	}
	if got.detail == nil {
		t.Fatal("detail = nil, want non-nil")
	}
	if cmd != nil {
		t.Fatal("cmd = non-nil, want nil for scheduled game (no fetches)")
	}
	if got.detail.loading {
		t.Fatal("detail.loading = true, want false for scheduled game")
	}
}
