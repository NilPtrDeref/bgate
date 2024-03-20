package model

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var TitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#06D6A0"))

var ChapterStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#EF476F")).
	Background(lipgloss.Color("#FCFCFC"))

var NumberStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#6A7FDB"))

type Verse struct {
	Book    string
	Chapter string
	Number  string
	Text    string
	Title   *string
}

func (v Verse) HasTitle() bool {
	return v.Title != nil
}

func (v Verse) TitleString() string {
	return TitleStyle.Render(*v.Title)
}

func (v Verse) ChapterString() string {
	return ChapterStyle.Render(" " + v.Chapter + " ")
}

func (v Verse) NumberString() string {
	return NumberStyle.Render(v.Number + " ")
}

var BookStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#6A7FDB"))

type Book struct {
	Name     string
	Chapters int
}

func (b Book) String() string {
	return fmt.Sprintf("%s (%d)", BookStyle.Render(b.Name), b.Chapters)
}
