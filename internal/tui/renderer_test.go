package tui

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/glamour/v2"
	"github.com/bangweiz/alkaid/internal/llm"
	"github.com/charmbracelet/x/ansi"
)

func TestStreamedMarkdownPreservesTerminalHistory(t *testing.T) {
	for _, width := range []int{40, 80, 120} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			markdown, err := glamour.NewTermRenderer(glamour.WithStylePath("light"), glamour.WithWordWrap(width))
			if err != nil {
				t.Fatal(err)
			}
			r, err := NewRenderer()
			if err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			r.markdown, r.output = markdown, &output
			actual := &testTerminal{width: width, height: 8}
			want := &testTerminal{width: width, height: 8}
			code := "This is the most basic way to fetch data from a URL.\n\n```go\npackage main\n\nimport (\n    \"fmt\"\n    \"io\"\n    \"net/http\"\n)\n\nfunc main() {\n    resp, err := http.Get(\"https://jsonplaceholder.typicode.com/posts/1\")\n    fmt.Println(resp, err)\n}\n```\n"
			table := "| Name | Value |\n| --- | --- |\n" + strings.Repeat("| row | value |\n", 12) + "| a much longer name | a much longer value |\n"
			for _, message := range []string{"First response.\n", code, code, table, "Short reply.\n"} {
				actual.write("> question\r\n")
				want.write("> question\r\n")
				var pending bytes.Buffer
				for end := 1; end <= len(message); end++ {
					pending.WriteByte(message[end-1])
					r.renderPending(&pending, false)
				}
				r.renderPending(&pending, true)
				if regexp.MustCompile(`\x1b\[[0-9;]*[AJ]`).Match(output.Bytes()) {
					t.Fatal("response output must not move upward or erase history")
				}
				actual.write(output.String())
				output.Reset()
				final, err := markdown.Render(message)
				if err != nil {
					t.Fatal(err)
				}
				want.write(strings.ReplaceAll(final, "\n", "\r\n"))
				actual.write("\r\n")
				want.write("\r\n")
				if strings.Join(strings.Fields(actual.text()), " ") != strings.Join(strings.Fields(want.text()), " ") {
					t.Fatalf("streamed screen differs from rendering once:\ngot:\n%s\nwant:\n%s", actual.text(), want.text())
				}
			}
		})
	}
}

func TestCompletedBlocksStreamBeforeResponseEnds(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	r.output = &output
	var pending bytes.Buffer
	pending.WriteString("First paragraph.\n\nSecond paragraph")
	r.renderPending(&pending, false)
	if !strings.Contains(ansi.Strip(output.String()), "First paragraph.") {
		t.Fatal("completed paragraph was not printed before end of response")
	}
	if strings.Contains(output.String(), "Second") {
		t.Fatal("unfinished paragraph was printed")
	}
	output.Reset()
	r.renderPending(&pending, false)
	if output.Len() != 0 {
		t.Fatal("completed paragraph was printed again")
	}
	r.renderPending(&pending, true)
	if !strings.Contains(ansi.Strip(output.String()), "Second paragraph") || pending.Len() != 0 {
		t.Fatal("final paragraph was not flushed")
	}
}

func TestCodeSpacingWithTabs(t *testing.T) {
	code := "package main\n\nimport (\n\t\"fmt\"\n\t\"io\"\n\t\"net/http\"\n\t\"time\"\n)\n\nfunc main() {\n\tclient := &http.Client{\n\t\tTimeout: time.Second * 10,\n\t}\n\tfmt.Println(client)\n}"
	for _, width := range []int{80, 240} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			r, err := NewRenderer()
			if err != nil {
				t.Fatal(err)
			}
			r.markdown, err = glamour.NewTermRenderer(glamour.WithStylePath("light"), glamour.WithWordWrap(width))
			if err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			r.output = &output
			pending := bytes.NewBufferString("```go\n" + code + "\n```\n")
			r.renderPending(pending, true)
			plain := strings.ReplaceAll(ansi.Strip(output.String()), "\r\n", "\n")
			lines := strings.Split(strings.TrimSuffix(plain, "\n\n"), "\n")
			want := strings.Split(code, "\n")
			if len(lines) != len(want) {
				t.Fatalf("got %d code lines, want %d: %q", len(lines), len(want), plain)
			}
			for i, line := range lines {
				if strings.TrimRight(line, " \t") != line {
					t.Fatalf("line %d has trailing padding: %q", i, line)
				}
				if strings.TrimSpace(line) != strings.TrimSpace(want[i]) {
					t.Fatalf("line %d: got %q, want %q", i, line, want[i])
				}
			}
			screen := &testTerminal{width: width, height: 8}
			screen.write(output.String())
			if len(strings.Split(screen.text(), "\n")) != len(want) {
				t.Fatalf("code unexpectedly wrapped in terminal: %q", screen.text())
			}
			if !strings.Contains(output.String(), "\x1b[") {
				t.Fatal("Markdown syntax coloring was lost")
			}
		})
	}
}

func TestCompactMarkdownPreservesUnicodeAndBlankLines(t *testing.T) {
	rendered := "\n\x1b[31m  中文 👋\x1b[0m     \n    \n    next line    \n\n"
	got := compactMarkdown(rendered)
	want := "  中文 👋\r\n\r\n    next line\r\n\r\n"
	if ansi.Strip(got) != want {
		t.Fatalf("got %q, want %q", ansi.Strip(got), want)
	}
	if !strings.Contains(got, "\x1b[31m") || !strings.HasSuffix(got, ansi.ResetStyle+"\r\n\r\n") {
		t.Fatal("color must be preserved and reset before returning to the prompt")
	}
}

func TestResponseStreamFlushesEachTurnOnce(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	r.output = &output
	for _, reply := range []string{"Unique first reply.", "Unique second reply.", ""} {
		stream := make(chan llm.ChatResponse, len(reply))
		for _, c := range reply {
			stream <- llm.ChatResponse{Message: llm.Message{Content: string(c)}}
		}
		close(stream)
		r.RenderResponseStream(stream)
	}
	plain := ansi.Strip(output.String())
	for _, reply := range []string{"Unique first reply.", "Unique second reply."} {
		if strings.Count(plain, reply) != 1 {
			t.Fatalf("expected %q exactly once in %q", reply, plain)
		}
	}
}

// A small terminal model for the cursor, erase, color and auto-wrap
// sequences used by the renderer. Retains all rows to check scrollback too.
type testTerminal struct {
	width, row, col int
	height, top     int
	lines           [][]rune
}

func (s *testTerminal) write(output string) {
	for i := 0; i < len(output); i++ {
		c := output[i]
		if c == '\x1b' && i+1 < len(output) && output[i+1] == '[' {
			start := i + 2
			i = start
			for i < len(output) && (output[i] < '@' || output[i] > '~') {
				i++
			}
			if i == len(output) {
				panic("incomplete escape")
			}
			n, _ := strconv.Atoi(output[start:i])
			switch output[i] {
			case 'A':
				s.row = max(s.top, s.row-max(1, n))
			case 'J':
				if s.row < len(s.lines) {
					s.lines = s.lines[:s.row+1]
					s.lines[s.row] = s.lines[s.row][:min(s.col, len(s.lines[s.row]))]
				}
			case 'm': // Color does not change cursor position.
			default:
				panic("unsupported escape")
			}
			continue
		}
		switch c {
		case '\r':
			s.col = 0
		case '\t':
			s.col = min(s.width-1, (s.col/8+1)*8)
		case '\n':
			s.row++
			s.top = max(s.top, s.row-s.height+1)
		default:
			r, size := utf8.DecodeRuneInString(output[i:])
			i += size - 1
			if s.col >= s.width {
				s.col = 0
				s.row++
				s.top = max(s.top, s.row-s.height+1)
			}
			for len(s.lines) <= s.row {
				s.lines = append(s.lines, nil)
			}
			for len(s.lines[s.row]) <= s.col {
				s.lines[s.row] = append(s.lines[s.row], ' ')
			}
			s.lines[s.row][s.col] = r
			s.col++
		}
	}
}

func (s *testTerminal) text() string {
	lines := make([]string, len(s.lines))
	for i, line := range s.lines {
		lines[i] = strings.TrimRight(string(line), " ")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}
