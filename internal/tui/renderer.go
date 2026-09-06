package tui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"charm.land/glamour/v2"
	"github.com/bangweiz/alkaid/internal/llm"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
)

type Renderer struct {
	stopThinkingAnimation func()
	markdown              *glamour.TermRenderer
	output                io.Writer
}

func NewRenderer() (*Renderer, error) {
	markdown, err := glamour.NewTermRenderer(glamour.WithStylePath("light"), glamour.WithWordWrap(terminalWidth()))
	if err != nil {
		return nil, err
	}
	return &Renderer{markdown: markdown, output: os.Stdout}, nil
}

func terminalWidth() int {
	width, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil || width <= 0 {
		return 80
	}
	return width
}
func terminalHeight() int {
	_, height, err := term.GetSize(os.Stdout.Fd())
	if err != nil || height <= 0 {
		return 24
	}
	return height
}

func (r *Renderer) startThinking() {
	r.stopThinking()
	stop, done := make(chan struct{}), make(chan struct{})
	var once sync.Once
	r.stopThinkingAnimation = func() { once.Do(func() { close(stop) }); <-done }
	fmt.Fprint(r.output, "\r⠋ Thinking…")
	go func() {
		defer close(done)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		for frame := 0; ; frame = (frame + 1) % len(frames) {
			select {
			case <-stop:
				fmt.Fprint(r.output, "\r\x1b[2K")
				return
			case <-ticker.C:
				fmt.Fprintf(r.output, "\r%s Thinking…", frames[frame])
			}
		}
	}()
}

func (r *Renderer) stopThinking() {
	if r.stopThinkingAnimation != nil {
		r.stopThinkingAnimation()
		r.stopThinkingAnimation = nil
	}
}

// RenderResponseStream redraws only the active response. Earlier responses are
// never revisited, so they remain intact in normal terminal scrollback.
func (r *Renderer) RenderResponseStream(stream <-chan llm.ChatResponse) {
	r.startThinking()
	defer r.stopThinking()
	var source strings.Builder
	state := responseRenderState{height: terminalHeight()}
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case chunk, ok := <-stream:
			if !ok {
				state.render(r, source.String())
				return
			}
			if chunk.Message.Content != "" {
				source.WriteString(chunk.Message.Content)
			}
		case <-ticker.C:
			if source.Len() > 0 {
				state.render(r, source.String())
			}
		}
	}
}

type responseRenderState struct {
	previous          string
	lines             []string
	committed, height int
}

func (s *responseRenderState) render(r *Renderer, source string) {
	rendered, err := r.markdown.Render(source)
	if err != nil {
		panic(err)
	}
	rendered = compactMarkdown(rendered)
	if rendered == s.previous {
		return
	}
	newLines := responseLines(rendered)
	if s.height < 1 {
		s.height = 24
	}
	if len(newLines)-s.committed > s.height {
		s.committed = len(newLines) - s.height
	}
	oldStart := min(s.committed, len(s.lines))
	newStart := min(s.committed, len(newLines))
	oldVisible, newVisible := s.lines[oldStart:], newLines[newStart:]
	first := commonPrefix(oldVisible, newVisible)
	if first == len(oldVisible) && strings.HasPrefix(rendered, s.previous) {
		r.stopThinking()
		fmt.Fprint(r.output, rendered[len(s.previous):])
	} else {
		if up := len(oldVisible) - first; up > 0 {
			fmt.Fprintf(r.output, "\x1b[%dA", up)
		}
		fmt.Fprint(r.output, "\r\x1b[J")
		r.stopThinking()
		fmt.Fprint(r.output, strings.Join(newVisible[first:], ""))
	}
	s.previous, s.lines = rendered, newLines
}

func responseLines(rendered string) []string {
	if rendered == "" {
		return nil
	}
	rendered = strings.TrimSuffix(strings.ReplaceAll(rendered, "\r\n", "\n"), "\n")
	parts := strings.Split(rendered, "\n")
	lines := make([]string, len(parts))
	for i, line := range parts {
		lines[i] = line + "\r\n"
	}
	return lines
}

func commonPrefix(a, b []string) int {
	n := min(len(a), len(b))
	for i := range n {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

func compactMarkdown(rendered string) string {
	lines := strings.Split(strings.ReplaceAll(rendered, "\r\n", "\n"), "\n")
	for i, line := range lines {
		content := strings.TrimRight(ansi.Strip(line), " \t\r")
		if content == "" {
			lines[i] = ""
		} else {
			lines[i] = ansi.Truncate(line, ansi.StringWidth(content), "")
		}
	}
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\r\n") + ansi.ResetStyle + "\r\n\r\n"
}
