package tui

// Record is one graded question (or, for table questions, one graded row)
// kept for the end-of-run score and results table.
type Record struct {
	Chapter       int
	Prompt        string
	UserAnswer    string
	CorrectAnswer string
	Correct       bool
	TypoWarn      bool
}
