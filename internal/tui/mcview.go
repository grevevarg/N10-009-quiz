package tui

import (
	"fmt"
	"math/rand"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/grevevarg/N10-009-quiz/internal/imgview"
	"github.com/grevevarg/N10-009-quiz/internal/model"
)

const displayLetters = "ABCDEF"

// mcDoneMsg is emitted once the user has seen the revealed answer and
// pressed on to the next question.
type mcDoneMsg struct{ record Record }

type mcModel struct {
	q           model.MultiChoiceQuestion
	displayKeys []string // internal option keys, in shuffled display order
	cursor      int
	revealed    bool
	pickedKey   string
	renderedImg string
	imgDet      imgview.Detector
}

func newMCModel(q model.MultiChoiceQuestion, rng *rand.Rand, imgDet imgview.Detector, images imageSet, maxCols, maxRows int) mcModel {
	keys := q.OptionKeys()
	rng.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })

	m := mcModel{q: q, displayKeys: keys, imgDet: imgDet}
	if q.HasImages {
		m.renderedImg = images.render(imgDet, q.Image, maxCols, maxRows)
	}
	return m
}

func (m mcModel) Init() tea.Cmd { return nil }

func (m mcModel) Update(msg tea.Msg) (mcModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if m.revealed {
		switch keyMsg.String() {
		case "enter", " ":
			done := func() tea.Msg { return mcDoneMsg{record: m.toRecord()} }
			return m, tea.Batch(done, tea.ClearScreen)
		}
		return m, nil
	}

	switch keyMsg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.displayKeys)-1 {
			m.cursor++
		}
	case "enter", " ":
		m.pickedKey = m.displayKeys[m.cursor]
		m.revealed = true
		return m, tea.ClearScreen
	default:
		if letterIdx := strings.IndexByte(displayLetters, upperByte(keyMsg.String())); letterIdx >= 0 && letterIdx < len(m.displayKeys) {
			m.cursor = letterIdx
			m.pickedKey = m.displayKeys[letterIdx]
			m.revealed = true
			return m, tea.ClearScreen
		}
	}
	return m, nil
}

func upperByte(s string) byte {
	if len(s) != 1 {
		return 0
	}
	b := s[0]
	if b >= 'a' && b <= 'z' {
		b -= 'a' - 'A'
	}
	return b
}

func (m mcModel) toRecord() Record {
	correct := m.q.IsCorrect(m.pickedKey)
	return Record{
		Chapter:       m.q.Chapter,
		Prompt:        m.q.Prompt,
		UserAnswer:    m.q.Options[m.pickedKey],
		CorrectAnswer: m.correctAnswerText(),
		Correct:       correct,
	}
}

func (m mcModel) correctAnswerText() string {
	var parts []string
	for _, k := range m.q.Correct {
		parts = append(parts, m.q.Options[k])
	}
	return strings.Join(parts, " / ")
}

func (m mcModel) View() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Chapter %d\n\n", m.q.Chapter))
	if m.q.Correct != nil && len(m.q.Correct) > 1 {
		b.WriteString(helpStyle.Render("(more than one answer is accepted for this question)"))
		b.WriteString("\n")
	}
	b.WriteString(promptStyle.Render(m.q.Prompt))
	b.WriteString("\n")
	if m.renderedImg != "" {
		b.WriteString(m.renderedImg)
		b.WriteString("\n")
	}

	for i, key := range m.displayKeys {
		letter := string(displayLetters[i])
		text := m.q.Options[key]
		line := fmt.Sprintf("%s) %s", letter, text)

		switch {
		case m.revealed && m.q.IsCorrect(key):
			line = correctStyle.Render("  " + line + "  <- correct")
		case m.revealed && key == m.pickedKey:
			line = incorrectStyle.Render("  " + line + "  <- your answer")
		case i == m.cursor:
			line = selectedStyle.Render("> " + line)
		default:
			line = "  " + line
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if !m.revealed {
		b.WriteString(helpStyle.Render("up/down or A-F to pick, enter or space to confirm"))
	} else {
		verdict := incorrectStyle.Render("Incorrect")
		if m.q.IsCorrect(m.pickedKey) {
			verdict = correctStyle.Render("Correct!")
		}
		b.WriteString(verdict)
		b.WriteString(explanationStyle.Render(m.q.Explanation))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("press enter or space to continue"))
	}
	return b.String()
}
