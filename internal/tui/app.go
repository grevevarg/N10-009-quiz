// Package tui implements the Quiztia Bubble Tea application: a chapter
// picker, three per-question-type quiz views, and a final score/summary
// screen, per quiztia-design.md.
package tui

import (
	"math/rand"

	tea "github.com/charmbracelet/bubbletea"

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
		return a, nil
	case mcDoneMsg:
		a.records = append(a.records, msg.record)
		a.mcIdx++
		a.advance()
		return a, nil
	case fbDoneMsg:
		a.records = append(a.records, msg.record)
		a.fbIdx++
		a.advance()
		return a, nil
	case tblQuestionDoneMsg:
		a.records = append(a.records, msg.records...)
		a.tblIdx++
		a.advance()
		return a, nil
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
	for {
		switch a.phase {
		case phaseMultiChoice:
			if a.mcIdx < len(a.mcList) {
				a.mc = newMCModel(a.mcList[a.mcIdx], a.rng, a.imgDet, a.images)
				return
			}
			a.phase = phaseFillBlank
		case phaseFillBlank:
			if a.fbIdx < len(a.fbList) {
				a.fb = newFBModel(a.fbList[a.fbIdx], a.imgDet, a.images)
				return
			}
			a.phase = phaseTable
		case phaseTable:
			if a.tblIdx < len(a.tblList) {
				a.tbl = newTblModel(a.tblList[a.tblIdx], a.imgDet, a.images)
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

func (a App) View() string {
	switch a.phase {
	case phasePicker:
		return a.picker.View()
	case phaseMultiChoice:
		return a.mc.View()
	case phaseFillBlank:
		return a.fb.View()
	case phaseTable:
		return a.tbl.View()
	case phaseSummary:
		return a.summary.View()
	}
	return ""
}
