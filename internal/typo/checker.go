// Package typo implements the fill-in-the-blank answer checker described in
// quiztia-design.md: exact match, then Damerau-Levenshtein distance-1 typo
// tolerance (warn, don't dock score) for answers longer than 6 characters.
package typo

import "strings"

// Result is the verdict CheckAnswer returns for a single free-text answer.
type Result string

const (
	Exact     Result = "EXACT"
	TypoWarn  Result = "TYPO_WARN"
	Incorrect Result = "INCORRECT"
)

// CheckAnswer grades a single free-text answer against a single target
// string. It is case-insensitive and normalizes internal whitespace (so a
// stray double space doesn't burn the distance-1 budget meant for a real
// character typo) before comparing.
func CheckAnswer(userInput, target string) Result {
	u := normalize(userInput)
	t := normalize(target)

	if u == t {
		return Exact
	}

	ur, tr := []rune(u), []rune(t)
	dist := damerauLevenshtein(ur, tr)
	if len(tr) > 6 && dist <= 1 {
		return TypoWarn
	}
	return Incorrect
}

// ExactMatch is used for table cells, which the design calls out as
// exact-match only (e.g. CIDR notation "/8" must not accept "8" or typo
// tolerance in general). It's still case-insensitive and whitespace
// normalized, matching every other answer check in the app.
func ExactMatch(userInput, target string) bool {
	return normalize(userInput) == normalize(target)
}

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.Join(strings.Fields(s), " ")
}

// damerauLevenshtein computes the optimal string alignment distance between
// two rune slices, where a transposition of two adjacent runes counts as a
// single edit (unlike plain Levenshtein, which would count it as two).
func damerauLevenshtein(a, b []rune) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// d[i][j] = distance between a[:i] and b[:j]
	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}

	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := d[i-1][j] + 1
			ins := d[i][j-1] + 1
			sub := d[i-1][j-1] + cost
			best := min3(del, ins, sub)

			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				trans := d[i-2][j-2] + 1
				if trans < best {
					best = trans
				}
			}
			d[i][j] = best
		}
	}
	return d[la][lb]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}
