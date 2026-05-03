package lookup

import "testing"

// ParseSeries should extract series name and volume number from titles
// when a recognizable separator + number suffix is present, and return
// empty values when not. False positives are worse than false negatives
// here — the user can always type series in by hand, but auto-grouping
// the wrong things together is annoying.
func TestParseSeries(t *testing.T) {
	cases := []struct {
		in      string
		series  string
		volume  int
	}{
		// keyword separators
		{"Berserk Volume 1", "Berserk", 1},
		{"Berserk Vol. 1", "Berserk", 1},
		{"Berserk Vol 1", "Berserk", 1},
		{"Berserk Book 1", "Berserk", 1},
		{"Berserk #1", "Berserk", 1},
		{"Naruto, Vol. 12", "Naruto", 12},
		{"One Piece, Volume 100", "One Piece", 100},

		// drops a trailing subtitle when the volume keyword comes first
		{"Berserk Vol. 1: The Black Swordsman", "Berserk", 1},

		// trims trailing punctuation/whitespace from series name
		{"  Berserk    Vol. 3   ", "Berserk", 3},

		// no series indicator → no match
		{"Dune", "", 0},
		{"1984", "", 0},
		{"The Lord of the Rings", "", 0},

		// ambiguous: trailing number alone shouldn't trigger (too many false
		// positives like "Star Wars Episode I" or "iOS 17")
		{"Star Wars 1", "", 0},

		// Roman numerals not supported (yet) — user can override
		{"Star Wars: Episode I", "", 0},

		// empty / whitespace
		{"", "", 0},
		{"   ", "", 0},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			gotSeries, gotVol := ParseSeries(tc.in)
			if gotSeries != tc.series || gotVol != tc.volume {
				t.Errorf("ParseSeries(%q) = (%q, %d), want (%q, %d)",
					tc.in, gotSeries, gotVol, tc.series, tc.volume)
			}
		})
	}
}
