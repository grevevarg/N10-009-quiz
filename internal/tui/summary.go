package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type summaryModel struct {
	table   table.Model
	score   int
	total   int
	quitted bool
}

func newSummaryModel(records []Record, width, height int) summaryModel {
	score := 0
	for _, r := range records {
		if r.Correct {
			score++
		}
	}

	cols := []table.Column{
		{Title: "Ch", Width: 3},
		{Title: "Question", Width: clampWidth(width-70, 20, 60)},
		{Title: "Your answer", Width: 20},
		{Title: "Textbook answer", Width: 20},
		{Title: "Result", Width: 8},
	}
	rows := make([]table.Row, len(records))
	for i, r := range records {
		result := "wrong"
		if r.Correct {
			result = "correct"
		} else if r.TypoWarn {
			result = "typo?"
		}
		rows[i] = table.Row{
			fmt.Sprintf("%d", r.Chapter),
			r.Prompt,
			r.UserAnswer,
			r.CorrectAnswer,
			result,
		}
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(clampWidth(height-8, 5, 20)),
	)
	return summaryModel{table: t, score: score, total: len(records)}
}

func clampWidth(w, min, max int) int {
	if w < min {
		return min
	}
	if w > max {
		return max
	}
	return w
}

func (m summaryModel) Init() tea.Cmd { return nil }

func (m summaryModel) Update(msg tea.Msg) (summaryModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if keyMsg.String() == "q" || keyMsg.String() == "esc" {
		m.quitted = true
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m summaryModel) View() string {
	var b strings.Builder
	pct := 0
	if m.total > 0 {
		pct = m.score * 100 / m.total
	}
	b.WriteString(titleStyle.Render(fmt.Sprintf("Score: %d / %d (%d%%)", m.score, m.total, pct)))
	b.WriteString("\n\n")
	b.WriteString(m.table.View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("up/down/pgup/pgdn to scroll, q to quit"))
	return b.String()
}
