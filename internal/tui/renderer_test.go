package tui

import (
	"bytes"
	"strings"
	"testing"

	"charm.land/glamour/v2"
	"github.com/bangweiz/alkaid/internal/llm"
	"github.com/charmbracelet/x/ansi"
)

func TestCompactMarkdownPreservesContent(t *testing.T) {
	got := ansi.Strip(compactMarkdown("\n\x1b[31m  中文 👋\x1b[0m     \n    \n"))
	if got != "  中文 👋\r\n\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestResponseRenderStateRedrawsOnlyActiveResponse(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	r.output = &out
	state := responseRenderState{height: 24}
	state.render(r, "value = \"partial")
	first := out.String()
	state.render(r, "value = \"complete\"")
	if !strings.Contains(out.String(), "\x1b[") {
		t.Fatal("expected redraw escape sequence")
	}
	if !strings.Contains(ansi.Strip(out.String()), "value = \"complete\"") {
		t.Fatal("final response missing")
	}
	if first == out.String() {
		t.Fatal("active response was not updated")
	}
}

func TestResponseStreamKeepsCompletedTurns(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	r.output = &out
	for _, reply := range []string{"first reply", "```toml\nversion = \"2.9\"\n```"} {
		stream := make(chan llm.ChatResponse, len(reply))
		for _, c := range reply {
			stream <- llm.ChatResponse{Message: llm.Message{Content: string(c)}}
		}
		close(stream)
		r.RenderResponseStream(stream)
	}
	plain := ansi.Strip(out.String())
	if strings.Count(plain, "first reply") != 1 || strings.Count(plain, `version = "2.9"`) != 1 {
		t.Fatalf("history changed: %q", plain)
	}
}

func TestResponseLines(t *testing.T) {
	rendered, err := glamour.NewTermRenderer(glamour.WithStylePath("light"), glamour.WithWordWrap(80))
	if err != nil {
		t.Fatal(err)
	}
	text, err := rendered.Render("one\n\ntwo")
	if err != nil {
		t.Fatal(err)
	}
	if len(responseLines(compactMarkdown(text))) != 4 {
		t.Fatal("unexpected line accounting")
	}
}
