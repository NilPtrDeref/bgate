package model

import (
	"fmt"
r/style"
	"github.com/charmbracelet/lipgloss"t/liple"
	"github.com/charmbracegloss
)

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
	return style.TitleStyle.Render(*v.Title)
}

func (v Verse) ChapterString() string {
	text := fmt.Sprintf(" %s: %d ", v.Book, v.Chapter)
	return style.ChapterStyle.Render(text)
}

func (v Verse) NumberString() string {
	text := fmt.Sprintf("%d ", v.Number)
	return style.NumberStyle.Render(text)
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
