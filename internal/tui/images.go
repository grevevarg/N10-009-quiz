package tui

import (
	"strings"

	"github.com/grevevarg/N10-009-quiz/internal/imgview"
)

// imageSet bundles each embedded figure's raw bytes (for inline terminal
// rendering) with its unpacked on-disk path (for the fallback message when
// the terminal can't render images inline).
type imageSet struct {
	bytes map[string][]byte
	paths map[string]string
}

func (s imageSet) render(det imgview.Detector, filename string) string {
	var b strings.Builder
	_ = det.Render(&b, s.bytes[filename], s.paths[filename])
	return b.String()
}
