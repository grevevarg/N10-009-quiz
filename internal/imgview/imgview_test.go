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
	if err := d.Render(&buf, nil, "/tmp/images/6.png"); err != nil {
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
			if err := d.Render(&buf, png, "unused.png"); err != nil {
				t.Fatalf("Render: %v", err)
			}
			if !strings.HasPrefix(buf.String(), c.prefix) {
				t.Errorf("Render() output doesn't start with %q escape prefix, got %q", c.prefix, buf.String()[:min(40, buf.Len())])
			}
		})
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
