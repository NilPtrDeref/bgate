package view

import (
	"strings"

	"github.com/woodywood117/bgate/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	content   []model.Content
	wrap      bool
	padding   int
	lines     []string
	scroll    int
	maxscroll int
	vheight   int
	vwidth    int
}

func New(content []model.Content, wrap bool, padding int) *Model {
	return &Model{
		content: content,
		wrap:    wrap,
		padding: padding,
		scroll:  0,
		vheight: 20,
		vwidth:  20,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) chunks(s string, chunkSize int) []string {
	words := strings.Split(s, " ")
	var chunks []string
	var current []string
	var ccount int

	for _, word := range words {
		var wsize, _ = lipgloss.Size(word)
		var size = chunkSize
		if !m.wrap && len(chunks) > 0 {
			size -= 4
		}

		if ccount+wsize > size {
			if !m.wrap && len(chunks) > 0 {
				current[0] = "    " + current[0]
			}
			chunks = append(chunks, strings.Join(current, " "))

			ccount = 0
			current = nil
		}

		ccount += wsize + 1
		current = append(current, word)
	}

	if !m.wrap && len(chunks) > 0 {
		current[0] = "    " + current[0]
	}
	chunks = append(chunks, strings.Join(current, " "))

	return chunks
}

func (m *Model) resize(width int) {
	lines := []string{}

	for i := 0; i < len(m.content); i++ {
		c := m.content[i]
		line := c.String()

		if !m.wrap && c.Type == model.VerseCont {
			line = "    " + line
		}

		if m.wrap && (c.Type == model.Verse || c.Type == model.VerseCont) {
			for {
				if i+1 >= len(m.content) {
					break
				}

				if m.content[i+1].Type != model.Verse && m.content[i+1].Type != model.VerseCont {
					break
				}

				line = strings.Join([]string{line, m.content[i+1].String()}, " ")
				i++
			}
		}

		chunked := m.chunks(line, width)
		lines = append(lines, chunked...)
	}

	m.lines = lines
}

func (m *Model) SetWindowSize(width, height int) {
	m.vheight = height

	m.vwidth = width
	m.resize(width - 2*m.padding)

	m.maxscroll = max(0, (len(m.lines)-m.vheight)+1)
	m.scroll = min(m.scroll, m.maxscroll)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if m.scroll < m.maxscroll {
				m.scroll++
			}
		case "k", "up":
			if m.scroll > 0 {
				m.scroll--
			}
		case "g":
			m.scroll = 0
		case "G":
			m.scroll = m.maxscroll
		}
	case tea.WindowSizeMsg:
		m.SetWindowSize(msg.Width, msg.Height)
	}
	return m, nil
}

func (m *Model) View() string {
	var view strings.Builder

	lpad := strings.Repeat(" ", m.padding)
	for i := 0; i < m.vheight-1; i++ {
		if m.scroll+i >= len(m.lines) {
			break
		}
		view.WriteString(lpad + m.lines[m.scroll+i] + "\n")
	}

	return view.String()
}
