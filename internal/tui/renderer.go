package tui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"charm.land/glamour/v2"
	"github.com/bangweiz/alkaid/internal/llm"
	"github.com/charmbracelet/x/term"
)

type Renderer struct {
	oldOutput             string
	stopThinkingAnimation func()
	markdown              *glamour.TermRenderer
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
	fmt.Print("\r⠋ Thinking…")
	go func() {
		defer close(done)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		frame := 0
		for {
			select {
			case <-stop:
				fmt.Print("\r\x1b[2K")
				return
			case <-ticker.C:
				frame = (frame + 1) % len(frames)
				fmt.Printf("\r%s Thinking…", frames[frame])
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

// RenderResponseStream shows a thinking indicator while waiting for response
// text, then renders the streamed response.
func (r *Renderer) RenderResponseStream(responseStream <-chan llm.ChatResponse) {
	r.startThinking()
	defer r.stopThinking()

	var buffer strings.Builder
	dirty := false
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case chunk, ok := <-responseStream:
			if !ok {
				if dirty {
					r.renderDiff(buffer.String())
				}
				return
			}

			if chunk.Message.Content != "" {
				buffer.WriteString(chunk.Message.Content)
				dirty = true
			}

		case <-ticker.C:
			if dirty {
				r.renderDiff(buffer.String())
				dirty = false
			}
		}
	}
}

func (r *Renderer) renderDiff(output string) {
	glamouredOutput, err := r.markdown.Render(output)
	if err != nil {
		panic(err)
	}
	r.stopThinking()

	oldLines := strings.Split(r.oldOutput, "\n")
	newLines := strings.Split(glamouredOutput, "\n")

	firstChanged := 0
	for firstChanged < len(oldLines) && firstChanged < len(newLines) && oldLines[firstChanged] == newLines[firstChanged] {
		firstChanged++
	}

	linesUp := len(oldLines) - firstChanged - 1

	if linesUp > 0 {
		fmt.Printf("\x1b[%dA", linesUp)
	}

	fmt.Print("\r")
	fmt.Print("\x1b[J")

	if firstChanged < len(newLines) {
		fmt.Print(strings.Join(newLines[firstChanged:], "\n"))
	}
	r.oldOutput = glamouredOutput
}
