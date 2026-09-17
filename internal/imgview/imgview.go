// Package imgview renders quiz figure images inline in terminals that
// support it (iTerm2, Kitty, Sixel), falling back to pointing the user at
// the file on disk otherwise. Capability is detected once and cached, per
// quiztia-design.md.
package imgview

import (
	"bytes"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/png"
	"io"
	"math"

	"github.com/BourgeoisBear/rasterm"
	xdraw "golang.org/x/image/draw"
)

// Protocol is the inline-image protocol this terminal supports, in
// preference order when more than one is available.
type Protocol int

const (
	ProtoNone Protocol = iota
	ProtoKitty
	ProtoIterm
	ProtoSixel
)

// Detector holds the terminal's image capability, probed once at startup.
type Detector struct {
	Protocol Protocol
	InTmux   bool
}

// Detect probes the current terminal for inline-image support. Call this
// once at startup and reuse the result; re-probing per question is
// unnecessary and the Sixel check in particular round-trips an escape
// sequence to the terminal.
func Detect() Detector {
	d := Detector{InTmux: rasterm.IsTmuxScreen()}

	switch {
	case rasterm.IsKittyCapable():
		d.Protocol = ProtoKitty
	case rasterm.IsItermCapable():
		d.Protocol = ProtoIterm
	default:
		if ok, err := rasterm.IsSixelCapable(); err == nil && ok {
			d.Protocol = ProtoSixel
		}
	}
	return d
}

// Supported reports whether Detect found any usable inline-image protocol.
func (d Detector) Supported() bool {
	return d.Protocol != ProtoNone
}

// kittyDeleteAll tells a Kitty-protocol terminal to drop every placed image
// and its cached data (a=d, d=A). Kitty tracks image placements separately
// from the text grid, so a plain screen clear/redraw doesn't reliably clear
// them on its own -- kept as a defensive belt-and-suspenders measure
// alongside the row-accounting fix in Render below, which is the real fix
// for stale/overlapping images.
const kittyDeleteAll = "\x1b_Ga=d,d=A\x1b\\"

// ClearCmd returns the escape sequence (if any) that should be emitted at
// the start of every frame to make sure no previously placed image lingers
// on screen. It's a no-op string for protocols that don't need it.
func (d Detector) ClearCmd() string {
	if d.Protocol == ProtoKitty {
		return kittyDeleteAll
	}
	return ""
}

// cellAspect approximates a terminal cell's width:height ratio in pixels.
// Most monospace fonts land close to 1:2 (e.g. 8x16), so a "square" image
// needs roughly half as many text rows as columns to look square on
// screen.
const cellAspect = 0.5

// Render writes pngData to out sized to fit within maxCols x maxRows
// terminal cells (preserving the image's own aspect ratio), then pads with
// blank lines so the returned block is exactly as many lines tall as the
// image will visually occupy.
//
// That row accounting is what actually matters here, not just cosmetics:
// bubbletea's renderer tracks screen state by counted text lines, with no
// idea that a single "line" holding an image escape sequence can occupy
// many real terminal rows. Undercounting it is what let a big image bleed
// into (and never get cleared from underneath) whatever question came
// next -- the caller building the rest of the screen around this block can
// now trust it's exactly `rows` lines, so nothing else gets laid out on
// top of it.
func (d Detector) Render(out io.Writer, pngData []byte, filePath string, maxCols, maxRows int) error {
	if !d.Supported() {
		fmt.Fprintf(out, "[this terminal doesn't support inline images -- open the figure directly: %s]\n", filePath)
		return nil
	}

	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return err
	}
	cols, rows := fitCells(img.Bounds().Dx(), img.Bounds().Dy(), maxCols, maxRows)

	if d.InTmux {
		fmt.Fprintln(out, "[note: running inside tmux/screen -- image may not render without passthrough configured]")
	}

	switch d.Protocol {
	case ProtoKitty:
		opts := rasterm.KittyImgOpts{DstCols: uint32(cols), DstRows: uint32(rows)}
		if err := rasterm.KittyWriteImage(out, img, opts); err != nil {
			return err
		}
	case ProtoIterm:
		opts := rasterm.ItermImgOpts{
			DisplayInline: true,
			Width:         fmt.Sprintf("%d", cols),
			Height:        fmt.Sprintf("%d", rows),
		}
		if err := rasterm.ItermWriteImageWithOptions(out, img, opts); err != nil {
			return err
		}
	case ProtoSixel:
		// Sixel has no "display at N cells" instruction of its own -- the
		// terminal just paints the pixels it's given -- so cell-fitting has
		// to happen by resizing the source pixels before encoding.
		if err := rasterm.SixelWriteImage(out, toPaletted(resizeToCells(img, cols, rows))); err != nil {
			return err
		}
	}

	for i := 1; i < rows; i++ {
		fmt.Fprint(out, "\n")
	}
	return nil
}

// fitCells picks the largest (cols, rows) that fit an image of the given
// pixel dimensions inside a maxCols x maxRows box without distorting it,
// accounting for cellAspect.
func fitCells(imgW, imgH, maxCols, maxRows int) (cols, rows int) {
	if imgW <= 0 || imgH <= 0 || maxCols <= 0 || maxRows <= 0 {
		return 1, 1
	}
	imgAspect := float64(imgW) / float64(imgH)

	cols = maxCols
	rows = int(math.Round(float64(cols) / imgAspect * cellAspect))
	if rows < 1 {
		rows = 1
	}
	if rows > maxRows {
		rows = maxRows
		cols = int(math.Round(float64(rows) / cellAspect * imgAspect))
		if cols < 1 {
			cols = 1
		}
		if cols > maxCols {
			cols = maxCols
		}
	}
	return cols, rows
}

// approxCellPx is an assumed monospace cell size in pixels, used only to
// pick a target resolution for the pre-Sixel resize below. The exact value
// doesn't matter much since the terminal will still stretch the result to
// whatever its real cell size is -- what matters is landing close to the
// (cols, rows) aspect ratio fitCells already computed.
const approxCellPxW, approxCellPxH = 8, 16

// resizeToCells scales img to approximately cols x rows terminal cells
// worth of pixels using a high-quality filter.
func resizeToCells(img image.Image, cols, rows int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, cols*approxCellPxW, rows*approxCellPxH))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), xdraw.Over, nil)
	return dst
}

// toPaletted dithers an arbitrary decoded image down to a paletted image,
// since Sixel encoding requires one.
func toPaletted(img image.Image) *image.Paletted {
	bounds := img.Bounds()
	pImg := image.NewPaletted(bounds, palette.Plan9)
	draw.FloydSteinberg.Draw(pImg, bounds, img, image.Point{})
	return pImg
}
