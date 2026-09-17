package imgview

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRender_FallsBackWithoutCapability(t *testing.T) {
	d := Detector{Protocol: ProtoNone}
	var buf bytes.Buffer
	if err := d.Render(&buf, nil, "/tmp/images/6.png", 40, 20); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(buf.String(), "/tmp/images/6.png") {
		t.Errorf("fallback message missing file path, got %q", buf.String())
	}
	if strings.ContainsAny(buf.String(), "\x1b") {
		t.Errorf("fallback message should not contain escape sequences, got %q", buf.String())
	}
}

func TestRender_EachProtocolProducesEscapeOutput(t *testing.T) {
	png, err := os.ReadFile("testdata/1x1.png")
	if err != nil {
		t.Fatalf("reading test fixture: %v", err)
	}

	cases := []struct {
		name   string
		proto  Protocol
		prefix string
	}{
		{"kitty", ProtoKitty, "\x1b_G"},
		{"iterm", ProtoIterm, "\x1b]1337;File="},
		{"sixel", ProtoSixel, "\x1bP"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := Detector{Protocol: c.proto}
			var buf bytes.Buffer
			// A 1x1 (square) source at a 40x20 cell budget should fit to
			// exactly 40x20, per fitCells' cellAspect assumption.
			if err := d.Render(&buf, png, "unused.png", 40, 20); err != nil {
				t.Fatalf("Render: %v", err)
			}
			out := buf.String()
			if !strings.HasPrefix(out, c.prefix) {
				t.Errorf("Render() output doesn't start with %q escape prefix, got %q", c.prefix, out[:min(40, len(out))])
			}
			// The whole point of the row accounting: the returned block
			// must be exactly as many lines as the image will occupy on
			// screen, so callers can lay out the rest of the question
			// without overlapping it.
			if got := strings.Count(out, "\n"); got != 19 {
				t.Errorf("Render() output has %d newlines, want 19 (20 rows total)", got)
			}
		})
	}
}

func TestFitCells(t *testing.T) {
	cases := []struct {
		name                         string
		imgW, imgH, maxCols, maxRows int
		wantCols, wantRows           int
	}{
		{"square image, width-constrained", 1, 1, 40, 20, 40, 20},
		{"square image, height-constrained", 1, 1, 40, 5, 10, 5},
		{"wide image", 200, 50, 40, 20, 40, 5},
		{"tall image", 50, 200, 40, 40, 20, 40},
		{"degenerate dimensions fall back to 1x1", 0, 0, 40, 20, 1, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cols, rows := fitCells(c.imgW, c.imgH, c.maxCols, c.maxRows)
			if cols != c.wantCols || rows != c.wantRows {
				t.Errorf("fitCells(%d,%d,%d,%d) = (%d,%d), want (%d,%d)",
					c.imgW, c.imgH, c.maxCols, c.maxRows, cols, rows, c.wantCols, c.wantRows)
			}
			if cols > c.maxCols || rows > c.maxRows {
				t.Errorf("fitCells result (%d,%d) exceeds budget (%d,%d)", cols, rows, c.maxCols, c.maxRows)
			}
		})
	}
}

func TestClearCmd(t *testing.T) {
	if got := (Detector{Protocol: ProtoKitty}).ClearCmd(); got != kittyDeleteAll {
		t.Errorf("ClearCmd() for kitty = %q, want %q", got, kittyDeleteAll)
	}
	for _, p := range []Protocol{ProtoNone, ProtoIterm, ProtoSixel} {
		if got := (Detector{Protocol: p}).ClearCmd(); got != "" {
			t.Errorf("ClearCmd() for protocol %v = %q, want empty (only Kitty tracks placements outside the text grid)", p, got)
		}
	}
}

func TestSupported(t *testing.T) {
	if (Detector{Protocol: ProtoNone}).Supported() {
		t.Error("ProtoNone should not report Supported")
	}
	if !(Detector{Protocol: ProtoKitty}).Supported() {
		t.Error("ProtoKitty should report Supported")
	}
}
