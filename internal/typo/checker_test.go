package typo

import "testing"

func TestCheckAnswer(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		target string
		want   Result
	}{
		{"exact", "LAN", "LAN", Exact},
		{"case insensitive", "lan", "LAN", Exact},
		{"trims whitespace", "  LAN  ", "LAN", Exact},
		{"transposition typo on long word", "recieve", "receive", TypoWarn},
		{"transposition typo, mixed case", "Recieve", "receive", TypoWarn},
		{"short word distance-1 is incorrect, not a typo", "LAN", "WAN", Incorrect},
		{"short word exactly 6 chars stays strict", "Access", "Accent", Incorrect},
		{"double space normalized before distance calc", "Spanning  Tree Protocol", "Spanning Tree Protocol", Exact},
		// Known false-positive risk: these are two different correct answers in
		// the written-lab bank, a single deletion apart, both >6 chars.
		{"distance-1 collision between two distinct real answers", "Symmetrical", "Asymmetrical", TypoWarn},
		{"totally different", "banana", "receive", Incorrect},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := CheckAnswer(c.input, c.target)
			if got != c.want {
				t.Errorf("CheckAnswer(%q, %q) = %v, want %v", c.input, c.target, got, c.want)
			}
		})
	}
}

func TestDamerauLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"ab", "ba", 1}, // transposition = 1 edit, not 2
		{"receive", "recieve", 1},
		{"kitten", "sitting", 3},
	}
	for _, c := range cases {
		got := damerauLevenshtein([]rune(c.a), []rune(c.b))
		if got != c.want {
			t.Errorf("damerauLevenshtein(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
