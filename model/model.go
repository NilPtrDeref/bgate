package model

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var SectionStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#06D6A0"))

var ChapterStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#EF476F")).
	Background(lipgloss.Color("#FCFCFC"))

type ContentType uint8

var (
	Section   ContentType = 1
	Chapter   ContentType = 2
	Verse     ContentType = 3
	VerseCont ContentType = 4
)

type Content struct {
	Type    ContentType
	Number  string
	Content string
}

func (c Content) String() string {
	view := strings.Builder{}
	switch c.Type {
	case Section:
		{
			view.WriteString(" " + c.Content)
		}
	case Chapter:
		{
			view.WriteString(" " + c.Number)
		}
	case Verse:
		{
			view.WriteString(" " + c.Number + c.Content)
		}
	case VerseCont:
		{
			view.WriteString("     " + c.Content)
		}
	}
	return view.String()
}
