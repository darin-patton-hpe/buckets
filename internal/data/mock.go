package data

import (
	"context"

	"github.com/darin-patton-hpe/nbalive"
	"github.com/darin-patton-hpe/nbalive/live"
)

// MockClient implements NBAClient for testing.
type MockClient struct {
	ScoreboardFunc       func(ctx context.Context) (*nbalive.ScoreboardResponse, error)
	ScoreboardByDateFunc func(ctx context.Context, date string) (*nbalive.ScoreboardResponse, error)
	BoxScoreFunc         func(ctx context.Context, gameID string) (*nbalive.BoxScoreResponse, error)
	PlayByPlayFunc       func(ctx context.Context, gameID string) (*nbalive.PlayByPlayResponse, error)
	WatchFunc            func(ctx context.Context, gameID string, cfg live.WatchConfig) <-chan live.Event
}

func (m *MockClient) Scoreboard(ctx context.Context) (*nbalive.ScoreboardResponse, error) {
	if m.ScoreboardFunc != nil {
		return m.ScoreboardFunc(ctx)
	}
	return &nbalive.ScoreboardResponse{}, nil
}

func (m *MockClient) ScoreboardByDate(ctx context.Context, date string) (*nbalive.ScoreboardResponse, error) {
	if m.ScoreboardByDateFunc != nil {
		return m.ScoreboardByDateFunc(ctx, date)
	}
	return &nbalive.ScoreboardResponse{}, nil
}

func (m *MockClient) BoxScore(ctx context.Context, gameID string) (*nbalive.BoxScoreResponse, error) {
	if m.BoxScoreFunc != nil {
		return m.BoxScoreFunc(ctx, gameID)
	}
	return &nbalive.BoxScoreResponse{}, nil
}

func (m *MockClient) PlayByPlay(ctx context.Context, gameID string) (*nbalive.PlayByPlayResponse, error) {
	if m.PlayByPlayFunc != nil {
		return m.PlayByPlayFunc(ctx, gameID)
	}
	return &nbalive.PlayByPlayResponse{}, nil
}

func (m *MockClient) Watch(ctx context.Context, gameID string, cfg live.WatchConfig) <-chan live.Event {
	if m.WatchFunc != nil {
		return m.WatchFunc(ctx, gameID, cfg)
	}
	ch := make(chan live.Event)
	close(ch)
	return ch
}
