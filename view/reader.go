package view

import (
	"strings"

	"github.com/woodywood117/bgate/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Reader struct {
	content   []model.Content
	wrap      bool
	padding   int
	lines     []string
	scroll    int
	maxscroll int
	vheight   int
	vwidth    int
}

func NewReader(content []model.Content, wrap bool, padding int) *Reader {
	return &Reader{
		content: content,
		wrap:    wrap,
		padding: padding,
		scroll:  0,
		vheight: 20,
		vwidth:  20,
	}
}

func (r *Reader) Init() tea.Cmd {
	return nil
}

func (r *Reader) chunks(s string, chunkSize int) []string {
	words := strings.Split(s, " ")
	var chunks []string
	var current []string
	var ccount int

	for _, word := range words {
		var wsize, _ = lipgloss.Size(word)
		var size = chunkSize
		if !r.wrap && len(chunks) > 0 {
			size -= 4
		}

		if ccount+wsize > size {
			if !r.wrap && len(chunks) > 0 {
				current[0] = "    " + current[0]
			}
			chunks = append(chunks, strings.Join(current, " "))

			ccount = 0
			current = nil
		}

		ccount += wsize + 1
		current = append(current, word)
	}

	if !r.wrap && len(chunks) > 0 {
		current[0] = "    " + current[0]
	}
	chunks = append(chunks, strings.Join(current, " "))

	return chunks
}

func (r *Reader) resize(width int) {
	lines := []string{}

	for i := 0; i < len(r.content); i++ {
		c := r.content[i]
		line := c.String()

		if !r.wrap && c.Type == model.VerseCont {
			line = "    " + line
		}

		if r.wrap && (c.Type == model.Verse || c.Type == model.VerseCont) {
			for {
				if i+1 >= len(r.content) {
					break
				}

				if r.content[i+1].Type != model.Verse && r.content[i+1].Type != model.VerseCont {
					break
				}

				line = strings.Join([]string{line, r.content[i+1].String()}, " ")
				i++
			}
		}

		chunked := r.chunks(line, width)
		lines = append(lines, chunked...)
	}

	r.lines = lines
}

func (r *Reader) SetWindowSize(width, height int) {
	r.vheight = height

	r.vwidth = width
	r.resize(width - 2*r.padding)

	r.maxscroll = max(0, (len(r.lines)-r.vheight)+1)
	r.scroll = min(r.scroll, r.maxscroll)
}

func (r *Reader) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
			return r, tea.Quit
		case "j", "down":
			if r.scroll < r.maxscroll {
				r.scroll++
			}
		case "k", "up":
			if r.scroll > 0 {
				r.scroll--
			}
		case "g":
			r.scroll = 0
		case "G":
			r.scroll = r.maxscroll
		}
	case tea.WindowSizeMsg:
		r.SetWindowSize(msg.Width, msg.Height)
	}
	return r, nil
}

func (r *Reader) View() string {
	var view strings.Builder

	lpad := strings.Repeat(" ", r.padding)
	for i := 0; i < r.vheight-1; i++ {
		if r.scroll+i >= len(r.lines) {
			break
		}
		view.WriteString(lpad + r.lines[r.scroll+i] + "\n")
	}

	return view.String()
}
