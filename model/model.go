package model

import (
	"github.com/charmbracelet/lipgloss"
)

var SectionStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#06D6A0"))

var ChapterStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#EF476F")).
	Background(lipgloss.Color("#FCFCFC"))

var VerseStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#6A7FDB"))

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
	switch c.Type {
	case Section:
		return SectionStyle.Render(c.Content)
	case Chapter:
		return ChapterStyle.Render(" " + c.Number)
	case Verse:
		return VerseStyle.Render(c.Number) + c.Content
	case VerseCont:
		return "    " + c.Content
	default:
		return ""
	}
}
