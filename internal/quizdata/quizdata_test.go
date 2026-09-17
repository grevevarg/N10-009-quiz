package quizdata

import (
	"math/rand"
	"os"
	"testing"

	"github.com/grevevarg/N10-009-quiz/internal/model"
)

func loadRealBank(t *testing.T) *Bank {
	t.Helper()
	b, err := Load(os.DirFS("../../"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return b
}

func TestLoadRealData(t *testing.T) {
	b := loadRealBank(t)
	if len(b.MultiChoice) != 420 {
		t.Errorf("MultiChoice count = %d, want 420", len(b.MultiChoice))
	}
	if len(b.WrittenLab) != 121 {
		t.Errorf("WrittenLab count = %d, want 121", len(b.WrittenLab))
	}
	for _, q := range b.MultiChoice {
		if q.Explanation == "" {
			t.Errorf("chapter %d question %q has no explanation", q.Chapter, q.Prompt)
		}
	}
}

func TestMultiChoiceForChapters_FiltersAndPreservesFixedBlocks(t *testing.T) {
	b := loadRealBank(t)
	rng := rand.New(rand.NewSource(1))

	got := b.MultiChoiceForChapters([]int{1, 5}, rng)
	if len(got) != 40 {
		t.Fatalf("got %d questions, want 40 (20 per chapter)", len(got))
	}
	for _, q := range got {
		if q.Chapter != 1 && q.Chapter != 5 {
			t.Errorf("unexpected chapter %d leaked into selection", q.Chapter)
		}
	}
}

func TestMultiChoiceForChapters_ChaptersNeverInterleave(t *testing.T) {
	b := loadRealBank(t)

	for seed := int64(0); seed < 30; seed++ {
		rng := rand.New(rand.NewSource(seed))
		got := b.MultiChoiceForChapters([]int{1, 5, 9}, rng)

		// Once we've moved off a chapter, we should never see it again --
		// otherwise chapters are interleaving instead of running as
		// separate contiguous blocks.
		seenChapters := map[int]bool{}
		prev := 0
		for i, q := range got {
			if q.Chapter != prev {
				if seenChapters[q.Chapter] {
					t.Fatalf("seed %d: chapter %d reappeared at index %d after another chapter started -- chapters are interleaving",
						seed, q.Chapter, i)
				}
				seenChapters[q.Chapter] = true
				prev = q.Chapter
			}
		}
	}
}

func TestMultiChoiceForChapters_ChapterOrderItselfShuffles(t *testing.T) {
	b := loadRealBank(t)

	firstChapterSeen := map[int]bool{}
	for seed := int64(0); seed < 30; seed++ {
		rng := rand.New(rand.NewSource(seed))
		got := b.MultiChoiceForChapters([]int{1, 5, 9}, rng)
		firstChapterSeen[got[0].Chapter] = true
	}
	if len(firstChapterSeen) < 2 {
		t.Error("chapter block order never varied across 30 seeds -- expected the chapters themselves to still shuffle relative to each other")
	}
}

func TestWrittenLabForChapters_FixedBlockChapterStaysContiguous(t *testing.T) {
	b := loadRealBank(t)

	// Chapter 9's 5 fill_blank statements are all random_order=false and
	// describe the same figure progressively; across many shuffles they
	// must never be interleaved with anything else or reordered among
	// themselves, since chapter 9 is the only chapter selected here.
	var want []string
	for _, q := range b.WrittenLab {
		if q.Chapter == 9 && q.Type == model.TypeFillBlank {
			want = append(want, q.Prompt)
		}
	}
	if len(want) != 5 {
		t.Fatalf("expected 5 chapter-9 fill_blank questions in source data, got %d", len(want))
	}

	for seed := int64(0); seed < 20; seed++ {
		rng := rand.New(rand.NewSource(seed))
		fb, _ := b.WrittenLabForChapters([]int{9}, rng)
		if len(fb) != len(want) {
			t.Fatalf("seed %d: got %d questions, want %d", seed, len(fb), len(want))
		}
		for i, q := range fb {
			if q.Prompt != want[i] {
				t.Fatalf("seed %d: chapter 9 order was reshuffled at index %d: got %q, want %q",
					seed, i, q.Prompt, want[i])
			}
		}
	}
}

func TestWrittenLabForChapters_TableRowShuffleRespectsFlag(t *testing.T) {
	b := loadRealBank(t)

	// Chapter 3 (pin-out tables) is random_order=false: row order must
	// never change.
	rng := rand.New(rand.NewSource(2))
	_, tables := b.WrittenLabForChapters([]int{3}, rng)
	for _, tbl := range tables {
		if len(tbl.Rows) == 0 {
			continue
		}
		if tbl.Rows[0].Cells[0] != "Pin 1" {
			t.Errorf("chapter 3 table row order was shuffled despite random_order=false: first row is %v", tbl.Rows[0].Cells)
		}
	}

	// Chapter 8 (CIDR table) is random_order=true: over enough seeds we
	// should see at least one shuffled arrangement.
	sawShuffled := false
	for seed := int64(0); seed < 30; seed++ {
		rng := rand.New(rand.NewSource(seed))
		_, tables := b.WrittenLabForChapters([]int{8}, rng)
		for _, tbl := range tables {
			if len(tbl.Rows) > 1 && tbl.Rows[0].Cells[0] != "/13" {
				sawShuffled = true
			}
		}
	}
	if !sawShuffled {
		t.Error("chapter 8 CIDR table rows never appeared shuffled across 30 seeds despite random_order=true")
	}
}
