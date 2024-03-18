package view

import (
	"bgate/model"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	content []model.Content
	lines   []string
	scroll  int
	vheight int
	vwidth  int
}

func New(content []model.Content) *Model {
	return &Model{
		content: content,
		scroll:  0,
		vheight: 20,
		vwidth:  20,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func chunks(s string, chunkSize int) []string {
	if len(s) == 0 {
		return nil
	}
	if chunkSize >= len(s) {
		return []string{s}
	}
	var chunks []string = make([]string, 0, (len(s)-1)/chunkSize+1)
	currentLen := 0
	currentStart := 0
	for i := range s {
		if currentLen == chunkSize {
			chunks = append(chunks, s[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, s[currentStart:])
	return chunks
}

func (m *Model) resize(width int) {
	lines := []string{}

	for _, c := range m.content {
		chunked := chunks(c.String(), width)
		for i, chunk := range chunked {
			switch c.Type {
			case model.Section:
				chunked[i] = model.SectionStyle.Render(chunk)
			case model.Chapter:
				chunked[i] = model.ChapterStyle.Render(chunk)
			}
		}

		lines = append(lines, chunked...)
	}

	m.lines = lines
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if m.scroll <= len(m.lines)-m.vheight {
				m.scroll++
			}
		case "k", "up":
			if m.scroll > 0 {
				m.scroll--
			}
		}
	case tea.WindowSizeMsg:
		m.vheight = msg.Height
		m.vwidth = msg.Width
		m.resize(msg.Width)

		m.scroll = min(
			m.scroll,
			max(0, len(m.lines)-m.vheight),
		)
	}
	return m, nil
}

func (m *Model) View() string {
	var view strings.Builder

	for i := 0; i < m.vheight-1; i++ {
		if m.scroll+i >= len(m.lines) {
			break
		}
		view.WriteString(m.lines[m.scroll+i] + "\n")
	}

	return view.String()
}
