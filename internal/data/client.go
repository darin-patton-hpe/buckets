// Package data provides an abstraction over the nbalive client for testability.
package data

import (
	"context"

	"github.com/darin-patton-hpe/nbalive"
	"github.com/darin-patton-hpe/nbalive/live"
)

// NBAClient is the interface used by the TUI to fetch NBA data.
// It mirrors the live.Client methods needed by the application.
type NBAClient interface {
	Scoreboard(ctx context.Context) (*nbalive.ScoreboardResponse, error)
	BoxScore(ctx context.Context, gameID string) (*nbalive.BoxScoreResponse, error)
	PlayByPlay(ctx context.Context, gameID string) (*nbalive.PlayByPlayResponse, error)
	Watch(ctx context.Context, gameID string, cfg live.WatchConfig) <-chan live.Event
}

// LiveClient wraps the real live.Client and implements NBAClient.
type LiveClient struct {
	c *live.Client
}

// NewLiveClient creates a LiveClient from a real live.Client.
func NewLiveClient(c *live.Client) *LiveClient {
	return &LiveClient{c: c}
}

func (l *LiveClient) Scoreboard(ctx context.Context) (*nbalive.ScoreboardResponse, error) {
	return l.c.Scoreboard(ctx)
}

func (l *LiveClient) BoxScore(ctx context.Context, gameID string) (*nbalive.BoxScoreResponse, error) {
	return l.c.BoxScore(ctx, gameID)
}

func (l *LiveClient) PlayByPlay(ctx context.Context, gameID string) (*nbalive.PlayByPlayResponse, error) {
	return l.c.PlayByPlay(ctx, gameID)
}

func (l *LiveClient) Watch(ctx context.Context, gameID string, cfg live.WatchConfig) <-chan live.Event {
	return l.c.Watch(ctx, gameID, cfg)
}
