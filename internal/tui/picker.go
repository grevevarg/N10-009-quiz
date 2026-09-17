package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/grevevarg/N10-009-quiz/internal/model"
)

const pickerCols = 3

// chaptersConfirmedMsg is emitted once the user has picked at least one
// chapter and pressed enter to start the quiz.
type chaptersConfirmedMsg struct {
	chapters []int
}

type pickerModel struct {
	selected map[int]bool
	sortDesc bool

	// cursor: -1 means the "select all" row; 0..len(order)-1 indexes into
	// order (the chapter numbers in current sort direction).
	cursor int
	order  []int
}

func newPickerModel() pickerModel {
	m := pickerModel{selected: make(map[int]bool), cursor: -1}
	m.rebuildOrder()
	return m
}

func (m *pickerModel) rebuildOrder() {
	m.order = make([]int, len(model.Chapters))
	for i, c := range model.Chapters {
		if m.sortDesc {
			m.order[i] = model.Chapters[len(model.Chapters)-1-i].Number
		} else {
			m.order[i] = c.Number
		}
	}
}

func (m pickerModel) allSelected() bool {
	for _, c := range model.Chapters {
		if !m.selected[c.Number] {
			return false
		}
	}
	return true
}

func (m *pickerModel) toggleSelectAll() {
	all := m.allSelected()
	for _, c := range model.Chapters {
		m.selected[c.Number] = !all
	}
}

func (m *pickerModel) toggleCursor() {
	if m.cursor == -1 {
		m.toggleSelectAll()
		return
	}
	n := m.order[m.cursor]
	m.selected[n] = !m.selected[n]
}

func (m pickerModel) chosenChapters() []int {
	var out []int
	for _, c := range model.Chapters {
		if m.selected[c.Number] {
			out = append(out, c.Number)
		}
	}
	return out
}

func (m pickerModel) Init() tea.Cmd { return nil }

func (m pickerModel) Update(msg tea.Msg) (pickerModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "up", "k":
		if m.cursor == -1 {
			break
		}
		if m.cursor-pickerCols >= 0 {
			m.cursor -= pickerCols
		} else {
			m.cursor = -1
		}
	case "down", "j":
		if m.cursor == -1 {
			m.cursor = 0
			break
		}
		if next := m.cursor + pickerCols; next < len(m.order) {
			m.cursor = next
		}
	case "left", "h":
		if m.cursor > 0 && m.cursor%pickerCols != 0 {
			m.cursor--
		}
	case "right", "l":
		if m.cursor >= 0 && m.cursor%pickerCols != pickerCols-1 && m.cursor+1 < len(m.order) {
			m.cursor++
		}
	case " ":
		m.toggleCursor()
	case "s":
		m.sortDesc = !m.sortDesc
		m.rebuildOrder()
	case "enter":
		if chosen := m.chosenChapters(); len(chosen) > 0 {
			return m, func() tea.Msg { return chaptersConfirmedMsg{chapters: chosen} }
		}
	}
	return m, nil
}

func (m pickerModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Quiztia -- pick chapters to review"))
	b.WriteString("\n\n")

	selectAllLine := renderCheckbox(m.allSelected()) + " Select all chapters"
	if m.cursor == -1 {
		selectAllLine = selectedStyle.Render("> " + selectAllLine)
	} else {
		selectAllLine = "  " + selectAllLine
	}
	b.WriteString(selectAllLine)
	b.WriteString("\n\n")

	rows := (len(m.order) + pickerCols - 1) / pickerCols
	colWidth := 34
	for r := 0; r < rows; r++ {
		var cells []string
		for c := 0; c < pickerCols; c++ {
			idx := r*pickerCols + c
			if idx >= len(m.order) {
				break
			}
			num := m.order[idx]
			title := chapterTitle(num)
			line := fmt.Sprintf("%s Ch %2d: %s", renderCheckbox(m.selected[num]), num, truncate(title, colWidth-14))
			if idx == m.cursor {
				line = selectedStyle.Render("> " + line)
			} else {
				line = "  " + line
			}
			cells = append(cells, lipgloss.NewStyle().Width(colWidth).Render(line))
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cells...))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	sortLabel := "ascending"
	if m.sortDesc {
		sortLabel = "descending"
	}
	b.WriteString(helpStyle.Render(fmt.Sprintf(
		"sort: %s  --  arrows/hjkl move, space toggles, s swaps sort, enter starts quiz, ctrl+c quits",
		sortLabel,
	)))
	return b.String()
}

func renderCheckbox(on bool) string {
	if on {
		return checkedBoxStyle.Render("[x]")
	}
	return "[ ]"
}

func chapterTitle(num int) string {
	for _, c := range model.Chapters {
		if c.Number == num {
			return c.Title
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}
