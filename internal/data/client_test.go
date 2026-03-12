package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/darin-patton-hpe/buckets/internal/data"
	"github.com/darin-patton-hpe/nbalive"
)

func TestInterfaceConformance(t *testing.T) {
	var _ data.NBAClient = (*data.MockClient)(nil)
	var _ data.NBAClient = (*data.LiveClient)(nil)
}

func TestMockClientScoreboard_DefaultAndCustom(t *testing.T) {
	t.Run("default returns empty response", func(t *testing.T) {
		m := &data.MockClient{}
		resp, err := m.Scoreboard(context.Background())
		if err != nil {
			t.Fatalf("Scoreboard() error = %v", err)
		}
		if resp == nil {
			t.Fatal("Scoreboard() response is nil")
		}
		if len(resp.Scoreboard.Games) != 0 {
			t.Fatalf("expected no games, got %d", len(resp.Scoreboard.Games))
		}
	})

	t.Run("custom function result is returned", func(t *testing.T) {
		want := &nbalive.ScoreboardResponse{
			Scoreboard: nbalive.Scoreboard{Games: []nbalive.Game{{GameID: "g1"}}},
		}
		m := &data.MockClient{
			ScoreboardFunc: func(context.Context) (*nbalive.ScoreboardResponse, error) {
				return want, nil
			},
		}

		got, err := m.Scoreboard(context.Background())
		if err != nil {
			t.Fatalf("Scoreboard() error = %v", err)
		}
		if got != want {
			t.Fatalf("Scoreboard() returned unexpected pointer")
		}
	})
}

func TestMockClientBoxScore_DefaultAndCustom(t *testing.T) {
	t.Run("default returns empty response", func(t *testing.T) {
		m := &data.MockClient{}
		resp, err := m.BoxScore(context.Background(), "0022400001")
		if err != nil {
			t.Fatalf("BoxScore() error = %v", err)
		}
		if resp == nil {
			t.Fatal("BoxScore() response is nil")
		}
		if resp.Game.GameID != "" {
			t.Fatalf("expected empty game id, got %q", resp.Game.GameID)
		}
	})

	t.Run("custom function result is returned", func(t *testing.T) {
		wantID := "0022400999"
		m := &data.MockClient{
			BoxScoreFunc: func(_ context.Context, gameID string) (*nbalive.BoxScoreResponse, error) {
				if gameID != wantID {
					t.Fatalf("BoxScore() gameID = %q, want %q", gameID, wantID)
				}
				return &nbalive.BoxScoreResponse{Game: nbalive.BoxScoreGame{GameID: gameID}}, nil
			},
		}

		got, err := m.BoxScore(context.Background(), wantID)
		if err != nil {
			t.Fatalf("BoxScore() error = %v", err)
		}
		if got.Game.GameID != wantID {
			t.Fatalf("BoxScore() game id = %q, want %q", got.Game.GameID, wantID)
		}
	})
}

func TestMockClientPlayByPlay_DefaultAndCustom(t *testing.T) {
	t.Run("default returns empty response", func(t *testing.T) {
		m := &data.MockClient{}
		resp, err := m.PlayByPlay(context.Background(), "0022400001")
		if err != nil {
			t.Fatalf("PlayByPlay() error = %v", err)
		}
		if resp == nil {
			t.Fatal("PlayByPlay() response is nil")
		}
		if len(resp.Game.Actions) != 0 {
			t.Fatalf("expected no actions, got %d", len(resp.Game.Actions))
		}
	})

	t.Run("custom function result is returned", func(t *testing.T) {
		wantID := "0022400002"
		m := &data.MockClient{
			PlayByPlayFunc: func(_ context.Context, gameID string) (*nbalive.PlayByPlayResponse, error) {
				if gameID != wantID {
					t.Fatalf("PlayByPlay() gameID = %q, want %q", gameID, wantID)
				}
				return &nbalive.PlayByPlayResponse{
					Game: nbalive.PlayByPlayGame{Actions: []nbalive.Action{{ActionNumber: 7}}},
				}, nil
			},
		}

		got, err := m.PlayByPlay(context.Background(), wantID)
		if err != nil {
			t.Fatalf("PlayByPlay() error = %v", err)
		}
		if len(got.Game.Actions) != 1 || got.Game.Actions[0].ActionNumber != 7 {
			t.Fatalf("PlayByPlay() unexpected actions: %+v", got.Game.Actions)
		}
	})
}

func TestMockClientWatch_DefaultAndCustom(t *testing.T) {
	t.Run("default returns already-closed channel", func(t *testing.T) {
		m := &data.MockClient{}
		ch := m.Watch(context.Background(), "0022400001", nbalive.WatchConfig{})

		select {
		case _, ok := <-ch:
			if ok {
				t.Fatal("expected closed channel")
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatal("timed out waiting for closed channel")
		}
	})

	t.Run("custom watch function is used", func(t *testing.T) {
		wantID := "0022401234"
		wantCfg := nbalive.WatchConfig{BoxScore: true}
		wantEvt := nbalive.Event{Kind: nbalive.EventAction, GameID: wantID}

		m := &data.MockClient{
			WatchFunc: func(_ context.Context, gameID string, cfg nbalive.WatchConfig) <-chan nbalive.Event {
				if gameID != wantID {
					t.Fatalf("Watch() gameID = %q, want %q", gameID, wantID)
				}
				if cfg.BoxScore != wantCfg.BoxScore {
					t.Fatalf("Watch() cfg.BoxScore = %v, want %v", cfg.BoxScore, wantCfg.BoxScore)
				}
				ch := make(chan nbalive.Event, 1)
				ch <- wantEvt
				close(ch)
				return ch
			},
		}

		ch := m.Watch(context.Background(), wantID, wantCfg)
		select {
		case got := <-ch:
			if got != wantEvt {
				t.Fatalf("Watch() event = %+v, want %+v", got, wantEvt)
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatal("timed out waiting for event")
		}
	})
}
