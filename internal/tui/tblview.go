package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/grevevarg/N10-009-quiz/internal/imgview"
	"github.com/grevevarg/N10-009-quiz/internal/model"
	"github.com/grevevarg/N10-009-quiz/internal/typo"
)

// tblQuestionDoneMsg carries one Record per row of the table question.
type tblQuestionDoneMsg struct{ records []Record }

type tblModel struct {
	q      model.WrittenLabQuestion
	widths []int

	rowIdx    int
	blankCols []int
	inputs    []textinput.Model
	focus     int
	revealed  bool
	results   []bool

	records []Record

	renderedImg string
}

func newTblModel(q model.WrittenLabQuestion, imgDet imgview.Detector, images imageSet) tblModel {
	m := tblModel{q: q, widths: columnWidths(q)}
	if q.HasImages {
		m.renderedImg = images.render(imgDet, q.Image)
	}
	m.startRow(0)
	return m
}

func columnWidths(q model.WrittenLabQuestion) []int {
	widths := make([]int, len(q.Header))
	for i, h := range q.Header {
		widths[i] = len(h)
	}
	for _, row := range q.Rows {
		for i, cell := range row.Cells {
			if cell != "_" && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
		for colStr, ans := range row.Answers {
			if col, err := strconv.Atoi(colStr); err == nil && col < len(widths) && len(ans) > widths[col] {
				widths[col] = len(ans)
			}
		}
	}
	return widths
}

func (m *tblModel) startRow(idx int) {
	m.rowIdx = idx
	m.revealed = false
	m.results = nil
	if idx >= len(m.q.Rows) {
		m.inputs = nil
		m.blankCols = nil
		return
	}
	row := m.q.Rows[idx]
	m.blankCols = nil
	for i, cell := range row.Cells {
		if cell == "_" {
			m.blankCols = append(m.blankCols, i)
		}
	}
	m.inputs = make([]textinput.Model, len(m.blankCols))
	for i := range m.inputs {
		ti := textinput.New()
		ti.CharLimit = 100
		ti.Width = 24
		m.inputs[i] = ti
	}
	m.focus = 0
	if len(m.inputs) > 0 {
		m.inputs[0].Focus()
	}
}

func (m tblModel) Init() tea.Cmd { return nil }

func (m tblModel) Update(msg tea.Msg) (tblModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if m.revealed {
		if keyMsg.String() == "enter" {
			next := m.rowIdx + 1
			if next >= len(m.q.Rows) {
				return m, func() tea.Msg { return tblQuestionDoneMsg{records: m.records} }
			}
			m.startRow(next)
		}
		return m, nil
	}

	// A row with no blanks (shouldn't happen in the data, but stay safe)
	// reveals immediately on enter.
	if len(m.inputs) == 0 {
		if keyMsg.String() == "enter" {
			m.grade()
		}
		return m, nil
	}

	switch keyMsg.String() {
	case "tab", "down":
		m.setFocus((m.focus + 1) % len(m.inputs))
		return m, nil
	case "shift+tab", "up":
		m.setFocus((m.focus - 1 + len(m.inputs)) % len(m.inputs))
		return m, nil
	case "enter":
		if m.focus < len(m.inputs)-1 {
			m.setFocus(m.focus + 1)
			return m, nil
		}
		m.grade()
		return m, nil
	}

	var cmd tea.Cmd
	m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
	return m, cmd
}

func (m *tblModel) setFocus(i int) {
	if len(m.inputs) == 0 {
		return
	}
	m.inputs[m.focus].Blur()
	m.focus = i
	m.inputs[m.focus].Focus()
}

func (m *tblModel) grade() {
	row := m.q.Rows[m.rowIdx]
	m.results = make([]bool, len(m.blankCols))
	var userVals, wantVals []string
	allCorrect := true
	for i, col := range m.blankCols {
		want := row.Answers[strconv.Itoa(col)]
		ok := typo.ExactMatch(m.inputs[i].Value(), want)
		m.results[i] = ok
		allCorrect = allCorrect && ok
		userVals = append(userVals, m.inputs[i].Value())
		wantVals = append(wantVals, want)
	}
	m.revealed = true
	m.records = append(m.records, Record{
		Chapter:       m.q.Chapter,
		Prompt:        rowPrompt(m.q, row),
		UserAnswer:    strings.Join(userVals, "; "),
		CorrectAnswer: strings.Join(wantVals, "; "),
		Correct:       allCorrect,
	})
}

// rowPrompt builds a human-readable description of a row for the results
// table, e.g. "SFTP, ?, ? -> Transport protocol / Port number".
func rowPrompt(q model.WrittenLabQuestion, row model.TableRow) string {
	var parts []string
	for i, cell := range row.Cells {
		if cell == "_" {
			cell = "?"
		}
		if i < len(q.Header) {
			parts = append(parts, fmt.Sprintf("%s: %s", q.Header[i], cell))
		} else {
			parts = append(parts, cell)
		}
	}
	return q.Prompt + " (" + strings.Join(parts, ", ") + ")"
}

func (m tblModel) View() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Chapter %d\n\n", m.q.Chapter))
	b.WriteString(promptStyle.Render(m.q.Prompt))
	b.WriteString("\n")
	if m.renderedImg != "" {
		b.WriteString(m.renderedImg)
		b.WriteString("\n")
	}

	b.WriteString(m.renderHeader())
	for i, row := range m.q.Rows {
		switch {
		case i < m.rowIdx:
			b.WriteString(m.renderRevealedRow(row, i))
		case i == m.rowIdx:
			b.WriteString(m.renderActiveRow(row))
		default:
			b.WriteString(m.renderPendingRow(row))
		}
	}

	b.WriteString("\n")
	switch {
	case m.rowIdx >= len(m.q.Rows):
		b.WriteString(helpStyle.Render("press enter to continue"))
	case m.revealed:
		b.WriteString(helpStyle.Render("press enter to continue"))
	default:
		b.WriteString(helpStyle.Render("fill in the blank(s), tab/enter to move, enter on the last one to submit"))
	}
	return b.String()
}

func (m tblModel) renderHeader() string {
	cells := make([]string, len(m.q.Header))
	for i, h := range m.q.Header {
		cells[i] = pad(h, m.widths[i])
	}
	return strings.Join(cells, " | ") + "\n" + strings.Repeat("-", lineWidth(m.widths)) + "\n"
}

func (m tblModel) renderPendingRow(row model.TableRow) string {
	cells := make([]string, len(row.Cells))
	for i, c := range row.Cells {
		cells[i] = pad(c, m.widths[i])
	}
	return strings.Join(cells, " | ") + "\n"
}

// renderRevealedRow is used for a past row this session's own m.records
// already scored; we don't retain per-cell correctness after the fact, so
// it's shown plainly with the answers filled in.
func (m tblModel) renderRevealedRow(row model.TableRow, idx int) string {
	cells := make([]string, len(row.Cells))
	for i, c := range row.Cells {
		if c == "_" {
			c = row.Answers[strconv.Itoa(i)]
		}
		cells[i] = pad(c, m.widths[i])
	}
	return strings.Join(cells, " | ") + "\n"
}

func (m tblModel) renderActiveRow(row model.TableRow) string {
	cells := make([]string, len(row.Cells))
	inputIdx := 0
	for i, c := range row.Cells {
		if c != "_" {
			cells[i] = pad(c, m.widths[i])
			continue
		}
		if m.revealed {
			ok := m.results[inputIdx]
			ans := row.Answers[strconv.Itoa(i)]
			badge := incorrectStyle.Render("x")
			if ok {
				badge = correctStyle.Render("v")
			}
			cells[i] = pad(fmt.Sprintf("%s [%s]", ans, badge), m.widths[i])
		} else {
			cells[i] = m.inputs[inputIdx].View()
		}
		inputIdx++
	}
	return selectedStyle.Render("> ") + strings.Join(cells, " | ") + "\n"
}

func pad(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

func lineWidth(widths []int) int {
	total := 0
	for _, w := range widths {
		total += w
	}
	return total + 3*(len(widths)-1)
}
