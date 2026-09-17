// Package quizdata loads the embedded question banks and builds the
// randomized per-session question lists the quiz engine and TUI consume.
package quizdata

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"

	"github.com/grevevarg/N10-009-quiz/internal/model"
)

// Bank holds every question from both data files, parsed once at startup.
type Bank struct {
	MultiChoice []model.MultiChoiceQuestion
	WrittenLab  []model.WrittenLabQuestion
}

// Load parses multichoice.json and writtenlab.json out of the given
// filesystem (the embed.FS built in main).
func Load(f fs.FS) (*Bank, error) {
	mcRaw, err := fs.ReadFile(f, "multichoice.json")
	if err != nil {
		return nil, fmt.Errorf("reading multichoice.json: %w", err)
	}
	var mc []model.MultiChoiceQuestion
	if err := json.Unmarshal(mcRaw, &mc); err != nil {
		return nil, fmt.Errorf("parsing multichoice.json: %w", err)
	}

	wlRaw, err := fs.ReadFile(f, "writtenlab.json")
	if err != nil {
		return nil, fmt.Errorf("reading writtenlab.json: %w", err)
	}
	var wl []model.WrittenLabQuestion
	if err := json.Unmarshal(wlRaw, &wl); err != nil {
		return nil, fmt.Errorf("parsing writtenlab.json: %w", err)
	}

	return &Bank{MultiChoice: mc, WrittenLab: wl}, nil
}

// chapterSet is a small membership helper over the chosen chapter numbers.
type chapterSet map[int]bool

func newChapterSet(chapters []int) chapterSet {
	s := make(chapterSet, len(chapters))
	for _, c := range chapters {
		s[c] = true
	}
	return s
}

// MultiChoiceForChapters returns a presentation-ordered copy of the
// multi-choice questions belonging to the selected chapters.
//
// Ordering rule: consecutive questions from the same chapter that share
// random_order=false are kept together, in their original relative order
// (some chapters use several JSON entries to represent one narrative, e.g.
// exhibit-based questions that build on each other); everything else is
// shuffled freely across chapters. This mirrors the row-order rule applied
// to table questions below.
func (b *Bank) MultiChoiceForChapters(chapters []int, rng *rand.Rand) []model.MultiChoiceQuestion {
	set := newChapterSet(chapters)
	var selected []model.MultiChoiceQuestion
	for _, q := range b.MultiChoice {
		if set[q.Chapter] {
			selected = append(selected, q)
		}
	}
	blocks := groupFixedBlocks(len(selected), func(i int) (int, bool) {
		return selected[i].Chapter, selected[i].RandomOrder
	})
	out := make([]model.MultiChoiceQuestion, 0, len(selected))
	for _, idx := range shuffleBlocks(blocks, rng) {
		out = append(out, selected[idx])
	}
	return out
}

// WrittenLabForChapters returns presentation-ordered copies of the
// fill-in-the-blank and table questions belonging to the selected chapters,
// as two separate lists (fillBlank, table). Table rows are shuffled
// per-question according to that question's own random_order flag.
func (b *Bank) WrittenLabForChapters(chapters []int, rng *rand.Rand) (fillBlank, table []model.WrittenLabQuestion) {
	set := newChapterSet(chapters)
	var fb, tb []model.WrittenLabQuestion
	for _, q := range b.WrittenLab {
		if !set[q.Chapter] {
			continue
		}
		switch q.Type {
		case model.TypeFillBlank:
			fb = append(fb, q)
		case model.TypeTable:
			tb = append(tb, cloneWithShuffledRows(q, rng))
		}
	}
	fillBlank = reorder(fb, rng)
	table = reorder(tb, rng)
	return fillBlank, table
}

func cloneWithShuffledRows(q model.WrittenLabQuestion, rng *rand.Rand) model.WrittenLabQuestion {
	if !q.RandomOrder || len(q.Rows) < 2 {
		return q
	}
	rows := make([]model.TableRow, len(q.Rows))
	copy(rows, q.Rows)
	rng.Shuffle(len(rows), func(i, j int) { rows[i], rows[j] = rows[j], rows[i] })
	q.Rows = rows
	return q
}

func reorder(qs []model.WrittenLabQuestion, rng *rand.Rand) []model.WrittenLabQuestion {
	blocks := groupFixedBlocks(len(qs), func(i int) (int, bool) {
		return qs[i].Chapter, qs[i].RandomOrder
	})
	out := make([]model.WrittenLabQuestion, 0, len(qs))
	for _, idx := range shuffleBlocks(blocks, rng) {
		out = append(out, qs[idx])
	}
	return out
}

// groupFixedBlocks partitions [0, n) into blocks: a run of consecutive
// indices sharing the same key and randomOrder=false becomes one fixed
// block; every randomOrder=true index is its own singleton block.
func groupFixedBlocks(n int, at func(i int) (chapter int, randomOrder bool)) [][]int {
	var blocks [][]int
	i := 0
	for i < n {
		chapter, ro := at(i)
		if ro {
			blocks = append(blocks, []int{i})
			i++
			continue
		}
		j := i + 1
		for j < n {
			c2, ro2 := at(j)
			if c2 != chapter || ro2 {
				break
			}
			j++
		}
		block := make([]int, j-i)
		for k := range block {
			block[k] = i + k
		}
		blocks = append(blocks, block)
		i = j
	}
	return blocks
}

// shuffleBlocks shuffles the block order, then flattens back to indices.
// A single-index "randomOrder=false" run of length 1 is included as-is: it
// only becomes a fixed block relative to its 2+ siblings.
func shuffleBlocks(blocks [][]int, rng *rand.Rand) []int {
	order := make([]int, len(blocks))
	for i := range order {
		order[i] = i
	}
	rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })

	out := make([]int, 0)
	for _, bi := range order {
		out = append(out, blocks[bi]...)
	}
	return out
}
