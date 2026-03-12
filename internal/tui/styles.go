package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// styles holds all UI styles, adaptive to dark/light backgrounds.
type styles struct {
	// Header / title
	title     lipgloss.Style
	subtitle  lipgloss.Style
	statusBar lipgloss.Style

	// Scoreboard
	gameRow         lipgloss.Style
	gameRowSelected lipgloss.Style
	teamTricode     lipgloss.Style
	score           lipgloss.Style
	liveIndicator   lipgloss.Style
	finalIndicator  lipgloss.Style
	scheduledTime   lipgloss.Style

	// Tabs
	activeTab   lipgloss.Style
	inactiveTab lipgloss.Style
	tabGap      lipgloss.Style

	// Box score table
	headerCell lipgloss.Style
	dataCell   lipgloss.Style
	starterRow lipgloss.Style
	benchRow   lipgloss.Style

	// Play-by-play
	pbpPeriod      lipgloss.Style
	pbpClock       lipgloss.Style
	pbpAction      lipgloss.Style
	pbpMade        lipgloss.Style
	pbpMissed      lipgloss.Style
	pbpScore       lipgloss.Style
	pbpDescription lipgloss.Style

	// Team stats
	statLabel lipgloss.Style
	statValue lipgloss.Style
	statBar   lipgloss.Style

	// General
	help    lipgloss.Style
	spinner lipgloss.Style
	errText lipgloss.Style
	dimText lipgloss.Style
}

func newStyles(isDark bool) styles {
	var (
		primary   color.Color
		secondary color.Color
		accent    color.Color
		subtle    color.Color
		warn      color.Color
		success   color.Color
		danger    color.Color
		bg        color.Color
		fg        color.Color
		dimFg     color.Color
	)

	if isDark {
		primary = lipgloss.Color("#7D56F4")
		secondary = lipgloss.Color("#6C71C4")
		accent = lipgloss.Color("#F25D94")
		subtle = lipgloss.Color("#383838")
		warn = lipgloss.Color("#FDBE5A")
		success = lipgloss.Color("#73D98A")
		danger = lipgloss.Color("#FF6B6B")
		bg = lipgloss.Color("#1A1A2E")
		fg = lipgloss.Color("#FAFAFA")
		dimFg = lipgloss.Color("#626262")
	} else {
		primary = lipgloss.Color("#5A3ECF")
		secondary = lipgloss.Color("#5B61B0")
		accent = lipgloss.Color("#D94478")
		subtle = lipgloss.Color("#E0E0E0")
		warn = lipgloss.Color("#C89A2E")
		success = lipgloss.Color("#3DA854")
		danger = lipgloss.Color("#CC4444")
		bg = lipgloss.Color("#FFFFFF")
		fg = lipgloss.Color("#1A1A1A")
		dimFg = lipgloss.Color("#999999")
	}

	_ = bg // reserved for future explicit background use

	return styles{
		// Header
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(primary).
			PaddingLeft(1),
		subtitle: lipgloss.NewStyle().
			Foreground(dimFg).
			PaddingLeft(1),
		statusBar: lipgloss.NewStyle().
			Foreground(dimFg).
			PaddingLeft(1),

		// Scoreboard
		gameRow: lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2),
		gameRowSelected: lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2).
			Bold(true).
			Background(subtle).
			Foreground(fg),
		teamTricode: lipgloss.NewStyle().
			Bold(true).
			Foreground(fg).
			Width(4),
		score: lipgloss.NewStyle().
			Bold(true).
			Foreground(fg).
			Width(4).
			Align(lipgloss.Right),
		liveIndicator: lipgloss.NewStyle().
			Bold(true).
			Foreground(danger).
			SetString("● LIVE"),
		finalIndicator: lipgloss.NewStyle().
			Foreground(dimFg).
			SetString("FINAL"),
		scheduledTime: lipgloss.NewStyle().
			Foreground(dimFg),

		// Tabs
		activeTab: lipgloss.NewStyle().
			Bold(true).
			Foreground(primary).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(primary).
			Padding(0, 1),
		inactiveTab: lipgloss.NewStyle().
			Foreground(dimFg).
			BorderStyle(lipgloss.HiddenBorder()).
			BorderBottom(true).
			Padding(0, 1),
		tabGap: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(subtle),

		// Box score
		headerCell: lipgloss.NewStyle().
			Bold(true).
			Foreground(secondary).
			Align(lipgloss.Right).
			PaddingRight(1),
		dataCell: lipgloss.NewStyle().
			Foreground(fg).
			Align(lipgloss.Right).
			PaddingRight(1),
		starterRow: lipgloss.NewStyle().
			Foreground(fg),
		benchRow: lipgloss.NewStyle().
			Foreground(dimFg),

		// Play-by-play
		pbpPeriod: lipgloss.NewStyle().
			Bold(true).
			Foreground(primary).
			PaddingTop(1),
		pbpClock: lipgloss.NewStyle().
			Foreground(secondary).
			Width(7),
		pbpAction: lipgloss.NewStyle().
			Foreground(fg),
		pbpMade: lipgloss.NewStyle().
			Foreground(success),
		pbpMissed: lipgloss.NewStyle().
			Foreground(danger),
		pbpScore: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),
		pbpDescription: lipgloss.NewStyle().
			Foreground(fg),

		// Team stats
		statLabel: lipgloss.NewStyle().
			Foreground(dimFg).
			Width(22).
			Align(lipgloss.Right).
			PaddingRight(2),
		statValue: lipgloss.NewStyle().
			Bold(true).
			Foreground(fg).
			Width(8).
			Align(lipgloss.Right),
		statBar: lipgloss.NewStyle().
			Foreground(warn),

		// General
		help: lipgloss.NewStyle().
			Foreground(dimFg).
			PaddingLeft(1),
		spinner: lipgloss.NewStyle().
			Foreground(accent),
		errText: lipgloss.NewStyle().
			Foreground(danger).
			Bold(true).
			PaddingLeft(1),
		dimText: lipgloss.NewStyle().
			Foreground(dimFg),
	}
}
