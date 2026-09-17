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

	"github.com/BourgeoisBear/rasterm"
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

// Render writes the given PNG image data to out using the detected
// protocol. If no protocol is supported, it writes a fallback message
// pointing the user at filePath (the on-disk location, unpacked from the
// embedded binary) instead of attempting any ASCII/Sixel degradation.
func (d Detector) Render(out io.Writer, pngData []byte, filePath string) error {
	if !d.Supported() {
		fmt.Fprintf(out, "[this terminal doesn't support inline images -- open the figure directly: %s]\n", filePath)
		return nil
	}

	if d.InTmux {
		fmt.Fprintln(out, "[note: running inside tmux/screen -- image may not render without passthrough configured]")
	}

	switch d.Protocol {
	case ProtoKitty:
		img, err := png.Decode(bytes.NewReader(pngData))
		if err != nil {
			return err
		}
		return rasterm.KittyWriteImage(out, img, rasterm.KittyImgOpts{})
	case ProtoIterm:
		img, err := png.Decode(bytes.NewReader(pngData))
		if err != nil {
			return err
		}
		return rasterm.ItermWriteImage(out, img)
	case ProtoSixel:
		img, err := png.Decode(bytes.NewReader(pngData))
		if err != nil {
			return err
		}
		return rasterm.SixelWriteImage(out, toPaletted(img))
	}
	return nil
}

// toPaletted dithers an arbitrary decoded image down to a paletted image,
// since Sixel encoding (unlike the Kitty/iTerm paths) requires one.
func toPaletted(img image.Image) *image.Paletted {
	bounds := img.Bounds()
	pImg := image.NewPaletted(bounds, palette.Plan9)
	draw.FloydSteinberg.Draw(pImg, bounds, img, image.Point{})
	return pImg
}
