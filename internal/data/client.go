// Package data provides an abstraction over the nbalive client for testability.
package data

import (
	"context"

	"github.com/darin-patton-hpe/nbalive"
)

// NBAClient is the interface used by the TUI to fetch NBA data.
// It mirrors the nbalive.Client methods needed by the application.
type NBAClient interface {
	Scoreboard(ctx context.Context) (*nbalive.ScoreboardResponse, error)
	BoxScore(ctx context.Context, gameID string) (*nbalive.BoxScoreResponse, error)
	PlayByPlay(ctx context.Context, gameID string) (*nbalive.PlayByPlayResponse, error)
	Watch(ctx context.Context, gameID string, cfg nbalive.WatchConfig) <-chan nbalive.Event
}

// LiveClient wraps the real nbalive.Client and implements NBAClient.
type LiveClient struct {
	c *nbalive.Client
}

// NewLiveClient creates a LiveClient from a real nbalive.Client.
func NewLiveClient(c *nbalive.Client) *LiveClient {
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

func (l *LiveClient) Watch(ctx context.Context, gameID string, cfg nbalive.WatchConfig) <-chan nbalive.Event {
	return l.c.Watch(ctx, gameID, cfg)
}
