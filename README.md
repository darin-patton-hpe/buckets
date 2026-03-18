# buckets

A terminal UI for visualizing and interacting with NBA live game data. Built with [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) and powered by the [nbalive](https://github.com/darin-patton-hpe/nbalive) client library.

## Features

- **Live scoreboard** — today's games with scores, records, and status
- **Game detail view** — box score, play-by-play, and team stats tabs
- **Live updates** — scores and play-by-play stream in real time via the NBA CDN
- **Keyboard and mouse navigation** — click games, click tabs, or use vim-style keys
- **Adaptive theming** — automatically detects dark/light terminal background

## Requirements

- Go 1.26+

## Install

```sh
go install github.com/darin-patton-hpe/buckets/cmd/buckets@latest
```

## Usage

```sh
buckets
```

### Scoreboard Controls

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Enter` | Select game |
| `q` | Quit |

### Game Detail Controls

| Key | Action |
|-----|--------|
| `1` / `2` / `3` | Switch to Box Score / Play-by-Play / Team Stats |
| `Tab` / `Shift+Tab` | Cycle tabs |
| `↑` / `k` / `↓` / `j` | Scroll content |
| `Esc` | Back to scoreboard |
| `q` | Quit |

Mouse clicks work on scoreboard rows and tab headers.

## Build

```sh
go build ./cmd/buckets
```

## Test

```sh
go test ./... -race
```

## Lint

```sh
go vet ./...
```

## Local Development

If you're working on both `buckets` and [nbalive](https://github.com/darin-patton-hpe/nbalive) simultaneously, add a replace directive to point at your local checkout:

```sh
go mod edit -replace github.com/darin-patton-hpe/nbalive=../nbalive
```

This tells Go to use the local `../nbalive` directory instead of fetching from the remote. Adjust the path to wherever your nbalive clone lives.

**Important:** Do not commit the replace directive. Remove it before pushing:

```sh
go mod edit -dropreplace github.com/darin-patton-hpe/nbalive
```

## Releasing

Releases are automated via GitHub Actions and [GoReleaser](https://goreleaser.com). To create a release:

1. Update the `VERSION` file to the new version (e.g., `v1.0.0`)
2. Open a PR titled `release v1.0.0`
3. Merge the PR

When merged, CI validates that the `VERSION` file matches the version in the merge commit message, then:

1. Runs tests
2. Creates a git tag matching the version
3. Builds binaries for darwin, linux, and windows (amd64 and arm64)
4. Publishes a GitHub Release with the artifacts

The release will fail if the `VERSION` file does not match the version in the PR title.

Pull requests targeting `main` automatically build snapshot binaries, available as downloadable artifacts on the workflow run.

## Project Structure

```
buckets/
  cmd/buckets/main.go         Entry point
  internal/
    data/
      client.go               NBAClient interface and LiveClient wrapper
      mock.go                  MockClient for testing
      client_test.go           Data layer tests
    tui/
      model.go                 Top-level Bubble Tea model, routing, Update loop
      game.go                  Game detail sub-model (tabs, viewport, live watcher)
      scoreboard.go            Scoreboard rendering
      boxscore.go              Box score rendering
      playbyplay.go            Play-by-play rendering
      teamstats.go             Team stats rendering
      styles.go                Adaptive dark/light theme styles
      keys.go                  Key bindings and tab constants
      messages.go              Custom tea.Msg types
      commands.go              Custom tea.Cmd functions
      model_test.go            Model and game detail tests
      render_test.go           Rendering tests
      helpers_test.go          Helper function tests
```

## Data Source

All data comes from the NBA's public CDN (`cdn.nba.com/static/json/liveData/`). No API keys or authentication required. The CDN updates every ~10-15 seconds during live games.

## License

MIT
