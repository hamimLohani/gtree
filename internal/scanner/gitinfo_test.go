package scanner

import (
	"testing"
	"time"
)

func TestHumanDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "just now"},
		{90 * time.Second, "1m ago"},
		{45 * time.Minute, "45m ago"},
		{2 * time.Hour, "2h ago"},
		{25 * time.Hour, "1d ago"},
		{8 * 24 * time.Hour, "1w ago"},
		{35 * 24 * time.Hour, "1mo ago"},
		{400 * 24 * time.Hour, "1y ago"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := humanDuration(tt.d)
			if got != tt.want {
				t.Errorf("humanDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestDirtyStateParsing(t *testing.T) {
	// We test the line-counting logic indirectly via a helper that
	// accepts porcelain output directly.
	tests := []struct {
		name      string
		porcelain string
		wantDirty bool
		wantCount int
	}{
		{"clean", "", false, 0},
		{"one modified", " M foo.go", true, 1},
		{"two files", " M foo.go\n?? bar.go", true, 2},
		{"whitespace only", "   \n   ", false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirty, count := parsePorcelain(tt.porcelain)
			if dirty != tt.wantDirty {
				t.Errorf("dirty = %v, want %v", dirty, tt.wantDirty)
			}
			if count != tt.wantCount {
				t.Errorf("count = %d, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestAheadBehindParsing(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		wantAhead  int
		wantBehind int
	}{
		{"no upstream", "", 0, 0},
		{"ahead 2", "2\t0", 2, 0},
		{"behind 3", "0\t3", 0, 3},
		{"both", "1\t2", 1, 2},
		{"malformed", "xyz", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b := parseAheadBehind(tt.output)
			if a != tt.wantAhead {
				t.Errorf("ahead = %d, want %d", a, tt.wantAhead)
			}
			if b != tt.wantBehind {
				t.Errorf("behind = %d, want %d", b, tt.wantBehind)
			}
		})
	}
}
