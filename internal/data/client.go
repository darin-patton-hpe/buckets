// Package data provides an abstraction over the nbalive client for testability.
package data

import (
	"context"

	"github.com/darin-patton-hpe/nbalive"
	"github.com/darin-patton-hpe/nbalive/live"
	"github.com/darin-patton-hpe/nbalive/stats"
)

// NBAClient is the interface used by the TUI to fetch NBA data.
type NBAClient interface {
	Scoreboard(ctx context.Context) (*nbalive.ScoreboardResponse, error)
	ScoreboardByDate(ctx context.Context, date string) (*nbalive.ScoreboardResponse, error)
	BoxScore(ctx context.Context, gameID string) (*nbalive.BoxScoreResponse, error)
	PlayByPlay(ctx context.Context, gameID string) (*nbalive.PlayByPlayResponse, error)
	Watch(ctx context.Context, gameID string, cfg live.WatchConfig) <-chan live.Event
}

// Client wraps both live.Client and stats.Client, implementing NBAClient.
type Client struct {
	lc *live.Client
	sc *stats.Client
}

// NewClient creates a Client from a live.Client and a stats.Client.
func NewClient(lc *live.Client, sc *stats.Client) *Client {
	return &Client{lc: lc, sc: sc}
}

func (c *Client) Scoreboard(ctx context.Context) (*nbalive.ScoreboardResponse, error) {
	return c.lc.Scoreboard(ctx)
}

func (c *Client) ScoreboardByDate(ctx context.Context, date string) (*nbalive.ScoreboardResponse, error) {
	return c.sc.ScoreboardByDate(ctx, date)
}

func (c *Client) BoxScore(ctx context.Context, gameID string) (*nbalive.BoxScoreResponse, error) {
	return c.lc.BoxScore(ctx, gameID)
}

func (c *Client) PlayByPlay(ctx context.Context, gameID string) (*nbalive.PlayByPlayResponse, error) {
	return c.lc.PlayByPlay(ctx, gameID)
}

func (c *Client) Watch(ctx context.Context, gameID string, cfg live.WatchConfig) <-chan live.Event {
	return c.lc.Watch(ctx, gameID, cfg)
}
