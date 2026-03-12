package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/darin-patton-hpe/nbalive"
)

func TestRenderScoreboard(t *testing.T) {
	s := newStyles(true)

	t.Run("empty games", func(t *testing.T) {
		got := renderScoreboard(nil, 0, 100, s)
		if !strings.Contains(got, "No games today") {
			t.Fatalf("expected empty-state text, got: %q", got)
		}
	})

	t.Run("single final game", func(t *testing.T) {
		game := nbalive.Game{
			GameID:         "0022400001",
			GameStatus:     nbalive.GameFinal,
			GameStatusText: "Final",
			Period:         4,
			AwayTeam:       nbalive.GameTeam{TeamTricode: "BOS", Score: 108, Wins: 40, Losses: 12},
			HomeTeam:       nbalive.GameTeam{TeamTricode: "LAL", Score: 102, Wins: 30, Losses: 22},
		}

		got := renderScoreboard([]nbalive.Game{game}, 0, 120, s)
		for _, want := range []string{"BOS", "LAL", "108", "102", "FINAL"} {
			if !strings.Contains(got, want) {
				t.Fatalf("expected %q in output, got: %q", want, got)
			}
		}
	})

	t.Run("in progress game", func(t *testing.T) {
		game := nbalive.Game{
			GameStatus: nbalive.GameInProgress,
			Period:     2,
			GameClock:  nbalive.Duration{Duration: 5*time.Minute + 30*time.Second},
			AwayTeam:   nbalive.GameTeam{TeamTricode: "BOS", Score: 55},
			HomeTeam:   nbalive.GameTeam{TeamTricode: "LAL", Score: 60},
		}

		got := renderScoreboard([]nbalive.Game{game}, 0, 120, s)
		for _, want := range []string{"LIVE", "Q2 5:30"} {
			if !strings.Contains(got, want) {
				t.Fatalf("expected %q in output, got: %q", want, got)
			}
		}
	})

	t.Run("cursor highlighting includes pointer", func(t *testing.T) {
		games := []nbalive.Game{
			{AwayTeam: nbalive.GameTeam{TeamTricode: "BOS"}, HomeTeam: nbalive.GameTeam{TeamTricode: "LAL"}},
			{AwayTeam: nbalive.GameTeam{TeamTricode: "NYK"}, HomeTeam: nbalive.GameTeam{TeamTricode: "MIA"}},
		}
		got := renderScoreboard(games, 1, 120, s)
		if !strings.Contains(got, "▸") {
			t.Fatalf("expected selected-row pointer, got: %q", got)
		}
	})
}

func TestRenderGameRow(t *testing.T) {
	s := newStyles(true)

	t.Run("final overtime includes OT suffix", func(t *testing.T) {
		g := nbalive.Game{
			GameStatus: nbalive.GameFinal,
			Period:     5,
			AwayTeam:   nbalive.GameTeam{TeamTricode: "BOS", Score: 108, Wins: 40, Losses: 12},
			HomeTeam:   nbalive.GameTeam{TeamTricode: "LAL", Score: 102, Wins: 30, Losses: 22},
		}
		got := renderGameRow(g, s)
		if !strings.Contains(got, "OT1") {
			t.Fatalf("expected OT suffix in row, got: %q", got)
		}
	})

	t.Run("scheduled uses status text", func(t *testing.T) {
		g := nbalive.Game{
			GameStatus:     nbalive.GameScheduled,
			GameStatusText: "7:30 PM ET",
			AwayTeam:       nbalive.GameTeam{TeamTricode: "BOS"},
			HomeTeam:       nbalive.GameTeam{TeamTricode: "LAL"},
		}
		got := renderGameRow(g, s)
		if !strings.Contains(got, "7:30 PM ET") {
			t.Fatalf("expected status text in row, got: %q", got)
		}
	})

	t.Run("in progress shows period and clock", func(t *testing.T) {
		g := nbalive.Game{
			GameStatus: nbalive.GameInProgress,
			Period:     3,
			GameClock:  nbalive.Duration{Duration: 11*time.Minute + 58*time.Second},
			AwayTeam:   nbalive.GameTeam{TeamTricode: "BOS"},
			HomeTeam:   nbalive.GameTeam{TeamTricode: "LAL"},
		}
		got := renderGameRow(g, s)
		if !strings.Contains(got, "Q3 11:58") {
			t.Fatalf("expected period+clock in row, got: %q", got)
		}
	})
}

func TestRenderBoxScore(t *testing.T) {
	s := newStyles(true)

	t.Run("nil game", func(t *testing.T) {
		got := renderBoxScore(nil, 120, s)
		if !strings.Contains(got, "Loading box score") {
			t.Fatalf("expected loading text, got: %q", got)
		}
	})

	t.Run("valid game with players", func(t *testing.T) {
		boxGame := &nbalive.BoxScoreGame{
			GameID:     "0022400001",
			GameStatus: nbalive.GameFinal,
			AwayTeam: nbalive.BoxTeam{
				TeamTricode: "BOS",
				Score:       108,
				Players: []nbalive.BoxPlayer{
					{NameI: "J. Tatum", Position: "SF", Starter: nbalive.BoolString(true), Played: nbalive.BoolString(true), Statistics: nbalive.PlayerStats{Points: 30, Assists: 5, ReboundsTotal: 8}},
				},
				Statistics: nbalive.TeamStats{Points: 108, Assists: 25, ReboundsTotal: 44},
			},
			HomeTeam: nbalive.BoxTeam{
				TeamTricode: "LAL",
				Score:       102,
				Players: []nbalive.BoxPlayer{
					{NameI: "L. James", Position: "PF", Starter: nbalive.BoolString(true), Played: nbalive.BoolString(true), Statistics: nbalive.PlayerStats{Points: 28, Assists: 7, ReboundsTotal: 10}},
				},
				Statistics: nbalive.TeamStats{Points: 102, Assists: 22, ReboundsTotal: 40},
			},
		}

		got := renderBoxScore(boxGame, 140, s)
		for _, want := range []string{"BOS", "LAL", "J. Tatum", "L. James", "PLAYER", "PTS", "REB", "AST"} {
			if !strings.Contains(got, want) {
				t.Fatalf("expected %q in output, got: %q", want, got)
			}
		}
	})
}

func TestRenderPlayByPlay(t *testing.T) {
	s := newStyles(true)

	t.Run("empty actions", func(t *testing.T) {
		got := renderPlayByPlay(nil, s)
		if !strings.Contains(got, "No play-by-play data available") {
			t.Fatalf("expected empty-state text, got: %q", got)
		}
	})

	t.Run("multiple actions reverse order with period headers", func(t *testing.T) {
		actions := []nbalive.Action{
			{
				ActionNumber: 1,
				Period:       1,
				ActionType:   "2pt",
				Description:  "Q1 older action",
				TeamTricode:  "BOS",
				IsFieldGoal:  1,
				ShotResult:   "Made",
				ScoreHome:    "0",
				ScoreAway:    "2",
			},
			{
				ActionNumber: 2,
				Period:       2,
				ActionType:   "2pt",
				Description:  "Q2 newer action",
				TeamTricode:  "LAL",
				IsFieldGoal:  1,
				ShotResult:   "Missed",
				ScoreHome:    "2",
				ScoreAway:    "2",
			},
		}

		got := renderPlayByPlay(actions, s)
		idxQ2 := strings.Index(got, "Quarter 2")
		idxQ1 := strings.Index(got, "Quarter 1")
		if idxQ2 == -1 || idxQ1 == -1 {
			t.Fatalf("expected both period headers, got: %q", got)
		}
		if idxQ2 > idxQ1 {
			t.Fatalf("expected reverse period order (Q2 before Q1), got: %q", got)
		}

		idxNew := strings.Index(got, "Q2 newer action")
		idxOld := strings.Index(got, "Q1 older action")
		if idxNew == -1 || idxOld == -1 {
			t.Fatalf("expected both actions, got: %q", got)
		}
		if idxNew > idxOld {
			t.Fatalf("expected reverse action order, got: %q", got)
		}
	})

	t.Run("made and missed shots use corresponding styles", func(t *testing.T) {
		madeDesc := "Made layup"
		missedDesc := "Missed three"
		actions := []nbalive.Action{
			{Period: 1, ActionType: "2pt", Description: madeDesc, IsFieldGoal: 1, ShotResult: "Made"},
			{Period: 1, ActionType: "3pt", Description: missedDesc, IsFieldGoal: 1, ShotResult: "Missed"},
		}

		got := renderPlayByPlay(actions, s)
		if !strings.Contains(got, s.pbpMade.Render(madeDesc)) {
			t.Fatalf("expected made-shot style output, got: %q", got)
		}
		if !strings.Contains(got, s.pbpMissed.Render(missedDesc)) {
			t.Fatalf("expected missed-shot style output, got: %q", got)
		}
	})
}

func TestRenderTeamStats(t *testing.T) {
	s := newStyles(true)

	t.Run("nil game", func(t *testing.T) {
		got := renderTeamStats(nil, 120, s)
		if !strings.Contains(got, "Loading team stats") {
			t.Fatalf("expected loading text, got: %q", got)
		}
	})

	t.Run("valid game shows labels and tricodes", func(t *testing.T) {
		game := &nbalive.BoxScoreGame{
			AwayTeam: nbalive.BoxTeam{
				TeamTricode: "BOS",
				Statistics:  nbalive.TeamStats{Points: 108, Assists: 25, ReboundsTotal: 44, FieldGoalsPercentage: 0.5},
			},
			HomeTeam: nbalive.BoxTeam{
				TeamTricode: "LAL",
				Statistics:  nbalive.TeamStats{Points: 102, Assists: 22, ReboundsTotal: 40, FieldGoalsPercentage: 0.45},
			},
		}

		got := renderTeamStats(game, 120, s)
		for _, want := range []string{"BOS", "LAL", "Points", "Assists", "Rebounds", "FG%"} {
			if !strings.Contains(got, want) {
				t.Fatalf("expected %q in output, got: %q", want, got)
			}
		}
	})
}

func TestRenderTabBar(t *testing.T) {
	s := newStyles(true)

	bar0 := renderTabBar(0, 120, s)
	bar1 := renderTabBar(1, 120, s)
	bar2 := renderTabBar(2, 120, s)

	for _, bar := range []string{bar0, bar1, bar2} {
		for _, name := range tabNames {
			if !strings.Contains(bar, name) {
				t.Fatalf("expected %q in tab bar, got: %q", name, bar)
			}
		}
	}

	if bar0 == bar1 || bar1 == bar2 || bar0 == bar2 {
		t.Fatalf("expected different rendering for different active tabs")
	}
}
