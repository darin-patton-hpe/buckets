package tui

import (
	"testing"
	"time"

	"github.com/darin-patton-hpe/nbalive"
)

func TestFormatInt(t *testing.T) {
	if got := formatInt(0); got != "0" {
		t.Fatalf("formatInt(0) = %q, want %q", got, "0")
	}
	if got := formatInt(42); got != "42" {
		t.Fatalf("formatInt(42) = %q, want %q", got, "42")
	}
	if got := formatInt(-7); got != "-7" {
		t.Fatalf("formatInt(-7) = %q, want %q", got, "-7")
	}
}

func TestTabFromX(t *testing.T) {
	if got := tabFromX(0); got != 0 {
		t.Fatalf("tabFromX(0) = %d, want 0", got)
	}
	if got := tabFromX(14); got != 1 {
		t.Fatalf("tabFromX(14) = %d, want 1", got)
	}
	if got := tabFromX(100); got != -1 {
		t.Fatalf("tabFromX(100) = %d, want -1", got)
	}
	if got := tabFromX(-1); got != -1 {
		t.Fatalf("tabFromX(-1) = %d, want -1", got)
	}
}

func TestGameIndexFromY(t *testing.T) {
	numGames := 3

	if got := gameIndexFromY(scoreboardHeaderLines-1, numGames); got != -1 {
		t.Fatalf("gameIndexFromY(y<header, %d) = %d, want -1", numGames, got)
	}
	if got := gameIndexFromY(scoreboardHeaderLines, numGames); got != 0 {
		t.Fatalf("gameIndexFromY(y==header, %d) = %d, want 0", numGames, got)
	}
	if got := gameIndexFromY(scoreboardHeaderLines+numGames, numGames); got != -1 {
		t.Fatalf("gameIndexFromY(y==header+numGames, %d) = %d, want -1", numGames, got)
	}
	if got := gameIndexFromY(scoreboardHeaderLines+numGames-1, numGames); got != numGames-1 {
		t.Fatalf("gameIndexFromY(last row, %d) = %d, want %d", numGames, got, numGames-1)
	}
}

func TestFormatClock(t *testing.T) {
	tests := []struct {
		name string
		d    nbalive.Duration
		want string
	}{
		{name: "five minutes", d: nbalive.Duration{Duration: 5 * time.Minute}, want: "5:00"},
		{name: "thirty seconds", d: nbalive.Duration{Duration: 30 * time.Second}, want: "0:30"},
		{name: "zero", d: nbalive.Duration{}, want: "0:00"},
		{name: "eleven fifty eight", d: nbalive.Duration{Duration: 11*time.Minute + 58*time.Second}, want: "11:58"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatClock(tt.d); got != tt.want {
				t.Fatalf("formatClock(%v) = %q, want %q", tt.d.Duration, got, tt.want)
			}
		})
	}
}

func TestPeriodName(t *testing.T) {
	for i := 1; i <= 4; i++ {
		want := "Quarter " + formatInt(i)
		if got := periodName(i); got != want {
			t.Fatalf("periodName(%d) = %q, want %q", i, got, want)
		}
	}

	if got := periodName(5); got != "Overtime" {
		t.Fatalf("periodName(5) = %q, want %q", got, "Overtime")
	}
	if got := periodName(6); got != "OT2" {
		t.Fatalf("periodName(6) = %q, want %q", got, "OT2")
	}
	if got := periodName(7); got != "OT3" {
		t.Fatalf("periodName(7) = %q, want %q", got, "OT3")
	}
}

func TestFitColumns(t *testing.T) {
	cols := boxScoreCols()
	nameWidth := 18

	t.Run("full width fits all", func(t *testing.T) {
		avail := nameWidth
		for _, c := range cols {
			avail += c.width
		}
		got := fitColumns(cols, avail, nameWidth)
		if len(got) != len(cols) {
			t.Fatalf("fitColumns(full) count = %d, want %d", len(got), len(cols))
		}
	})

	t.Run("narrow width fits subset", func(t *testing.T) {
		got := fitColumns(cols, nameWidth+11, nameWidth)
		if len(got) != 2 {
			t.Fatalf("fitColumns(narrow) count = %d, want 2", len(got))
		}
		if got[0].header != "MIN" || got[1].header != "PTS" {
			t.Fatalf("fitColumns(narrow) headers = %q, %q; want MIN, PTS", got[0].header, got[1].header)
		}
	})

	t.Run("zero available fits none", func(t *testing.T) {
		got := fitColumns(cols, 0, nameWidth)
		if len(got) != 0 {
			t.Fatalf("fitColumns(zero) count = %d, want 0", len(got))
		}
	})
}

func TestStatRow(t *testing.T) {
	got := statRow("Points", 108, 102)
	if got.label != "Points" {
		t.Fatalf("label = %q, want %q", got.label, "Points")
	}
	if got.away != "108" || got.home != "102" {
		t.Fatalf("values = (%q,%q), want (%q,%q)", got.away, got.home, "108", "102")
	}
	if got.awayN != 108 || got.homeN != 102 {
		t.Fatalf("numeric values = (%v,%v), want (108,102)", got.awayN, got.homeN)
	}
}

func TestStatRowPct(t *testing.T) {
	got := statRowPct("FG%", 0.475, 0.521)
	if got.label != "FG%" {
		t.Fatalf("label = %q, want %q", got.label, "FG%")
	}
	if got.away != "47.5" || got.home != "52.1" {
		t.Fatalf("values = (%q,%q), want (%q,%q)", got.away, got.home, "47.5", "52.1")
	}
	if got.awayN != 0.475 || got.homeN != 0.521 {
		t.Fatalf("numeric values = (%v,%v), want (0.475,0.521)", got.awayN, got.homeN)
	}
}
