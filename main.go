// Command quiztia is an interactive terminal quiz for the CompTIA N10-009
// Network+ exam, built from the two textbook question banks embedded in
// this binary.
package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/grevevarg/N10-009-quiz/internal/imgview"
	"github.com/grevevarg/N10-009-quiz/internal/quizdata"
	"github.com/grevevarg/N10-009-quiz/internal/tui"
)

//go:embed data/multichoice.json data/writtenlab.json images
var dataFS embed.FS

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "quiztia:", err)
		os.Exit(1)
	}
}

func run() error {
	bank, err := quizdata.Load(dataFS)
	if err != nil {
		return err
	}

	imageBytes, imagePaths, err := unpackImages(dataFS)
	if err != nil {
		return err
	}

	imgDet := imgview.Detect()

	app := tui.NewApp(bank, imageBytes, imagePaths, imgDet)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// unpackImages copies the embedded figure PNGs out to a per-user cache
// directory so the no-inline-image-support fallback message can point at a
// real, openable file -- the embedded bytes alone aren't reachable as a
// path once compiled into the binary.
func unpackImages(f embed.FS) (bytesByName map[string][]byte, pathsByName map[string]string, err error) {
	entries, err := fs.ReadDir(f, "images")
	if err != nil {
		return nil, nil, fmt.Errorf("reading embedded images dir: %w", err)
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	outDir := filepath.Join(cacheDir, "quiztia", "images")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("creating image cache dir: %w", err)
	}

	bytesByName = make(map[string][]byte, len(entries))
	pathsByName = make(map[string]string, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := fs.ReadFile(f, "images/"+e.Name())
		if err != nil {
			return nil, nil, fmt.Errorf("reading embedded image %s: %w", e.Name(), err)
		}
		outPath := filepath.Join(outDir, e.Name())
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			return nil, nil, fmt.Errorf("unpacking image %s: %w", e.Name(), err)
		}
		bytesByName[e.Name()] = data
		pathsByName[e.Name()] = outPath
	}
	return bytesByName, pathsByName, nil
}
