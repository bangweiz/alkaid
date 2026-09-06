package tui

import (
	"bytes"
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
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type Renderer struct {
	stopThinkingAnimation func()
	markdown              *glamour.TermRenderer
	parser                parser.Parser
	output                io.Writer
}

// NewRenderer initializes a Markdown renderer for the current terminal width.
func NewRenderer() (*Renderer, error) {
	width := terminalWidth()
	markdown, err := glamour.NewTermRenderer(
		glamour.WithStylePath("light"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}
	return &Renderer{
		markdown: markdown,
		parser:   goldmark.New(goldmark.WithExtensions(extension.GFM, extension.DefinitionList)).Parser(),
		output:   os.Stdout,
	}, nil
}

func terminalWidth() int {
	width, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

// startThinking animates a waiting indicator until response text is rendered.
func (r *Renderer) startThinking() {
	r.stopThinking()
	stop := make(chan struct{})
	done := make(chan struct{})
	var once sync.Once
	r.stopThinkingAnimation = func() {
		once.Do(func() { close(stop) })
		<-done
	}
	fmt.Fprint(r.output, "\r⠋ Thinking…")
	go func() {
		defer close(done)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		frame := 0
		for {
			select {
			case <-stop:
				fmt.Fprint(r.output, "\r\x1b[2K")
				return
			case <-ticker.C:
				frame = (frame + 1) % len(frames)
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

// RenderResponseStream appends completed Markdown blocks once. The unfinished
// block stays buffered because new tokens can change its layout and highlighting.
func (r *Renderer) RenderResponseStream(responseStream <-chan llm.ChatResponse) {
	r.startThinking()
	defer r.stopThinking()

	var buffer bytes.Buffer
	dirty := false
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case chunk, ok := <-responseStream:
			if !ok {
				r.renderPending(&buffer, true)
				return
			}

			if chunk.Message.Content != "" {
				buffer.WriteString(chunk.Message.Content)
				dirty = true
			}

		case <-ticker.C:
			if dirty {
				r.renderPending(&buffer, false)
				dirty = false
			}
		}
	}
}

func (r *Renderer) renderPending(buffer *bytes.Buffer, final bool) {
	end := buffer.Len()
	if !final {
		end = r.completedPrefix(buffer.Bytes())
	}
	if end == 0 {
		return
	}
	glamouredOutput, err := r.markdown.Render(string(buffer.Next(end)))
	if err != nil {
		panic(err)
	}
	r.stopThinking()
	fmt.Fprint(r.output, compactMarkdown(glamouredOutput))
}

// Glamour pads lines to the wrap width, but tabs occupy additional terminal
// columns. Remove that padding without stripping color or code indentation.
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

// A new top-level paragraph or heading gives an unambiguous source boundary.
// Keep other trailing blocks buffered: fences, lists and tables can still grow.
func (r *Renderer) completedPrefix(source []byte) int {
	last := r.parser.Parse(text.NewReader(source)).LastChild()
	if last == nil || last.PreviousSibling() == nil {
		return 0
	}
	switch last.(type) {
	case *ast.Paragraph, *ast.Heading:
		if last.Lines().Len() > 0 {
			start := last.Lines().At(0).Start
			return bytes.LastIndexByte(source[:start], '\n') + 1
		}
	}
	return 0
}
