package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/grevevarg/N10-009-quiz/internal/imgview"
	"github.com/grevevarg/N10-009-quiz/internal/model"
	"github.com/grevevarg/N10-009-quiz/internal/typo"
)

type fbDoneMsg struct{ record Record }

type fbModel struct {
	q           model.WrittenLabQuestion
	inputs      []textinput.Model
	focus       int
	revealed    bool
	results     []typo.Result
	renderedImg string
}

func newFBModel(q model.WrittenLabQuestion, imgDet imgview.Detector, images imageSet) fbModel {
	inputs := make([]textinput.Model, len(q.Blanks))
	for i := range inputs {
		ti := textinput.New()
		ti.Placeholder = fmt.Sprintf("blank %d", i+1)
		ti.CharLimit = 200
		ti.Width = 40
		inputs[i] = ti
	}
	if len(inputs) > 0 {
		inputs[0].Focus()
	}

	m := fbModel{q: q, inputs: inputs}
	if q.HasImages {
		m.renderedImg = images.render(imgDet, q.Image)
	}
	return m
}

func (m fbModel) Init() tea.Cmd { return nil }

func (m fbModel) Update(msg tea.Msg) (fbModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if m.revealed {
		if keyMsg.String() == "enter" {
			return m, func() tea.Msg { return fbDoneMsg{record: m.toRecord()} }
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

func (m *fbModel) setFocus(i int) {
	m.inputs[m.focus].Blur()
	m.focus = i
	m.inputs[m.focus].Focus()
}

func (m *fbModel) grade() {
	m.results = make([]typo.Result, len(m.q.Blanks))
	for i, want := range m.q.Blanks {
		m.results[i] = typo.CheckAnswer(m.inputs[i].Value(), want)
	}
	m.revealed = true
}

func (m fbModel) allCorrect() bool {
	for _, r := range m.results {
		if r == typo.Incorrect {
			return false
		}
	}
	return true
}

func (m fbModel) anyTypoWarn() bool {
	for _, r := range m.results {
		if r == typo.TypoWarn {
			return true
		}
	}
	return false
}

func (m fbModel) toRecord() Record {
	userVals := make([]string, len(m.inputs))
	for i, ti := range m.inputs {
		userVals[i] = ti.Value()
	}
	return Record{
		Chapter:       m.q.Chapter,
		Prompt:        m.q.Prompt,
		UserAnswer:    strings.Join(userVals, "; "),
		CorrectAnswer: strings.Join(m.q.Blanks, "; "),
		Correct:       m.allCorrect(),
		TypoWarn:      m.anyTypoWarn(),
	}
}

func (m fbModel) View() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Chapter %d\n\n", m.q.Chapter))
	b.WriteString(promptStyle.Render(m.q.Prompt))
	b.WriteString("\n")
	if m.renderedImg != "" {
		b.WriteString(m.renderedImg)
		b.WriteString("\n")
	}

	for i, ti := range m.inputs {
		label := fmt.Sprintf("Blank %d: ", i+1)
		if len(m.inputs) == 1 {
			label = "Answer: "
		}
		b.WriteString(label)
		b.WriteString(ti.View())
		if m.revealed {
			b.WriteString("  ")
			b.WriteString(verdictBadge(m.results[i]))
			b.WriteString(fmt.Sprintf("  (textbook answer: %s)", m.q.Blanks[i]))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if !m.revealed {
		b.WriteString(helpStyle.Render("type your answer(s), tab/enter to move between blanks, enter on the last one to submit"))
	} else {
		b.WriteString(helpStyle.Render("press enter to continue"))
	}
	return b.String()
}

func verdictBadge(r typo.Result) string {
	switch r {
	case typo.Exact:
		return correctStyle.Render("correct")
	case typo.TypoWarn:
		return typoWarnStyle.Render("close enough (typo)")
	default:
		return incorrectStyle.Render("incorrect")
	}
}
