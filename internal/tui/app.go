// Package tui implements the Quiztia Bubble Tea application: a chapter
// picker, three per-question-type quiz views, and a final score/summary
// screen, per quiztia-design.md.
package tui

import (
	"math/rand"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/grevevarg/N10-009-quiz/internal/imgview"
	"github.com/grevevarg/N10-009-quiz/internal/model"
	"github.com/grevevarg/N10-009-quiz/internal/quizdata"
)

type phase int

const (
	phasePicker phase = iota
	phaseMultiChoice
	phaseFillBlank
	phaseTable
	phaseSummary
)

// App is the root Bubble Tea model.
type App struct {
	bank   *quizdata.Bank
	images imageSet
	imgDet imgview.Detector
	rng    *rand.Rand

	phase         phase
	width, height int

	picker pickerModel

	mcList []model.MultiChoiceQuestion
	mcIdx  int
	mc     mcModel

	fbList []model.WrittenLabQuestion
	fbIdx  int
	fb     fbModel

	tblList []model.WrittenLabQuestion
	tblIdx  int
	tbl     tblModel

	records []Record
	summary summaryModel
}

// NewApp builds the initial application model. imageBytes maps figure
// filenames (e.g. "6.png") to their raw embedded bytes, and imagePaths maps
// the same filenames to where they were unpacked on disk (used only for the
// no-inline-image-support fallback message).
func NewApp(bank *quizdata.Bank, imageBytes map[string][]byte, imagePaths map[string]string, imgDet imgview.Detector) App {
	return App{
		bank:   bank,
		images: imageSet{bytes: imageBytes, paths: imagePaths},
		imgDet: imgDet,
		rng:    rand.New(rand.NewSource(rand.Int63())),
		phase:  phasePicker,
		picker: newPickerModel(),
	}
}

func (a App) Init() tea.Cmd { return nil }

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	case chaptersConfirmedMsg:
		a.startQuiz(msg.chapters)
		return a, tea.ClearScreen
	case mcDoneMsg:
		a.records = append(a.records, msg.record)
		a.mcIdx++
		a.advance()
		return a, tea.ClearScreen
	case fbDoneMsg:
		a.records = append(a.records, msg.record)
		a.fbIdx++
		a.advance()
		return a, tea.ClearScreen
	case tblQuestionDoneMsg:
		a.records = append(a.records, msg.records...)
		a.tblIdx++
		a.advance()
		return a, tea.ClearScreen
	}

	var cmd tea.Cmd
	switch a.phase {
	case phasePicker:
		a.picker, cmd = a.picker.Update(msg)
	case phaseMultiChoice:
		a.mc, cmd = a.mc.Update(msg)
	case phaseFillBlank:
		a.fb, cmd = a.fb.Update(msg)
	case phaseTable:
		a.tbl, cmd = a.tbl.Update(msg)
	case phaseSummary:
		a.summary, cmd = a.summary.Update(msg)
	}
	return a, cmd
}

func (a *App) startQuiz(chapters []int) {
	a.mcList = a.bank.MultiChoiceForChapters(chapters, a.rng)
	a.fbList, a.tblList = a.bank.WrittenLabForChapters(chapters, a.rng)
	a.mcIdx, a.fbIdx, a.tblIdx = 0, 0, 0
	a.records = nil
	a.phase = phaseMultiChoice
	a.enterPhase()
}

// advance moves to the next question within the current phase, or to the
// next non-empty phase if the current one is exhausted.
func (a *App) advance() {
	a.enterPhase()
}

// enterPhase sets up the active sub-model for a.phase, skipping forward
// through phases/questions that have nothing left.
func (a *App) enterPhase() {
	cols, rows := a.imageBudget()
	for {
		switch a.phase {
		case phaseMultiChoice:
			if a.mcIdx < len(a.mcList) {
				a.mc = newMCModel(a.mcList[a.mcIdx], a.rng, a.imgDet, a.images, cols, rows)
				return
			}
			a.phase = phaseFillBlank
		case phaseFillBlank:
			if a.fbIdx < len(a.fbList) {
				a.fb = newFBModel(a.fbList[a.fbIdx], a.imgDet, a.images, cols, rows)
				return
			}
			a.phase = phaseTable
		case phaseTable:
			if a.tblIdx < len(a.tblList) {
				a.tbl = newTblModel(a.tblList[a.tblIdx], a.imgDet, a.images, cols, rows)
				return
			}
			a.phase = phaseSummary
		case phaseSummary:
			a.summary = newSummaryModel(a.records, a.width, a.height)
			return
		default:
			return
		}
	}
}

// imageBudget sizes the box a figure is allowed to occupy, based on the
// current terminal size: capped at the same column width the rest of the
// content wraps to, and at half the terminal's height so a tall image
// never crowds the prompt/options/help text out of view below it.
func (a App) imageBudget() (cols, rows int) {
	width, height := a.width, a.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	cols = contentWidth
	if width-4 < cols {
		cols = width - 4
	}
	if cols < 20 {
		cols = 20
	}

	rows = height / 2
	if rows < 6 {
		rows = 6
	}
	if rows > 24 {
		rows = 24
	}
	return cols, rows
}

func (a App) View() string {
	var content string
	switch a.phase {
	case phasePicker:
		content = a.picker.View()
	case phaseMultiChoice:
		content = a.mc.View()
	case phaseFillBlank:
		content = a.fb.View()
	case phaseTable:
		content = a.tbl.View()
	case phaseSummary:
		content = a.summary.View()
	}

	// lipgloss.Place centers each line of a multi-line string independently
	// based on that line's own width, which makes a left-aligned list (menu
	// options, table rows) look staggered instead of centered as one block.
	// Padding every line out to a uniform rectangle first keeps the internal
	// layout left-aligned and lets Place center the block as a whole.
	blockWidth := lipgloss.Width(content)
	if blockWidth < contentWidth {
		blockWidth = contentWidth
	}
	block := lipgloss.NewStyle().Width(blockWidth).Render(content)

	frame := block
	if a.width > 0 && a.height > 0 {
		frame = lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, block)
	}

	// Every frame starts by dropping any previously placed inline image
	// (Kitty tracks placements independent of the text grid, so a screen
	// clear alone doesn't reliably clear a stale one) -- see imgview.ClearCmd.
	// Prepended after centering so the escape sequence itself never affects
	// the width math above.
	return a.imgDet.ClearCmd() + frame
}
