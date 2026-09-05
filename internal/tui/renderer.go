package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/glamour/v2"
	"github.com/bangweiz/alkaid/internal/llm"
)

type Renderer struct {
	oldOutput string
}

func (r *Renderer) RenderResponseStream(responseStream <-chan llm.ChatResponse) {
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

			buffer.WriteString(chunk.Message.Content)
			dirty = true

		case <-ticker.C:
			if dirty {
				r.renderDiff(buffer.String())
				dirty = false
			}
		}
	}
}

func (r *Renderer) renderDiff(output string) {
	glamouredOutput, err := glamour.Render(output, "light")
	if err != nil {
		panic(err)
	}

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
