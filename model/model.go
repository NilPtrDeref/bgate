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
	Book    string  `db:"book"`
	Chapter int     `db:"chapter"`
	Number  int     `db:"number"`
	Part    int     `db:"part"`
	Text    string  `db:"text"`
	Title   *string `db:"title"`
}

func (v Verse) HasTitle() bool {
	return v.Title != nil
}

func (v Verse) TitleString() string {
	return TitleStyle.Render(*v.Title)
}

func (v Verse) ChapterString() string {
	text := fmt.Sprintf(" %d ", v.Chapter)
	return ChapterStyle.Render(text)
}

func (v Verse) NumberString() string {
	text := fmt.Sprintf("%d ", v.Number)
	return NumberStyle.Render(text)
}

var BookStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#6A7FDB"))

type Book struct {
	Name     string `db:"name"`
	Chapters int    `db:"chapters"`
}

func (b Book) String() string {
	return fmt.Sprintf("%s (%d)", BookStyle.Render(b.Name), b.Chapters)
}
