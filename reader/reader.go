package reader

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/woodywood117/bgate/reader/model"
	"github.com/woodywood117/bgate/reader/style"
	"github.com/woodywood117/bgate/search"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mode int

const (
	read mode = iota
	searching
	help
)

type Reader struct {
	searcher search.Searcher
	query    string
	viewport viewport.Model
	ready    bool
	mode     mode
	wrap     bool
	padding  int

	first model.Verse
	last  model.Verse
	books []model.Book

	searchbuffer string
}

func NewReader(searcher search.Searcher, query string) *Reader {
	return &Reader{
		searcher: searcher,
		query:    query,
	}
}

func (r *Reader) Init() tea.Cmd {
	return nil
}

func (r *Reader) SetPadding(p int) {
	r.padding = p
	if r.ready {
		r.viewport.Style = r.viewport.Style.Padding(0, r.padding)
	}
}

func (r *Reader) SetWrap(w bool) {
	r.wrap = w
}

func (r *Reader) Query(query string) (string, error) {
	r.query = query

	verses, err := r.searcher.Query(query)
	if err != nil {
		return "", err
	}

	var writer strings.Builder
	for index, verse := range verses {
		if index == 0 {
			r.first = verse
		}
		if index == len(verses)-1 {
			r.last = verse
		}

		title := verse.HasTitle()
		chapter := verse.Number == 1 && verse.Part == 1

		if index > 0 && r.wrap && (title || chapter) {
			writer.WriteString("\n")
		}

		if title {
			writer.WriteString(verse.TitleString() + "\n")
		}

		if chapter {
			writer.WriteString(verse.ChapterString() + "\n")
		}

		if verse.Part == 1 {
			writer.WriteString(verse.NumberString())
		}

		writer.WriteString(verse.Text + " ")

		if !r.wrap {
			writer.WriteString("\n")
		}
	}

	indentation := "    "
	if r.wrap {
		indentation = ""
	}
	return ResizeString(writer.String(), r.viewport.Width-(2*r.padding), indentation), nil
}

func (r *Reader) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if r.mode == read {
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				return r, tea.Quit
			case "g":
				r.viewport.GotoTop()
			case "G":
				r.viewport.GotoBottom()
			case "+":
				r.padding++
				r.viewport.Style = r.viewport.Style.Padding(0, r.padding)
			case "-":
				r.padding = max(0, r.padding-1)
				r.viewport.Style = r.viewport.Style.Padding(0, r.padding)
			case "w":
				r.viewport.YOffset = 0
				r.wrap = !r.wrap
				content, err := r.Query(r.query)
				if err != nil {
					e := err.Error()
					r.viewport.SetContent(style.ErrorStyle.Render(e))
					return r, nil
				}
				r.viewport.SetContent(content)
			case "p":
				r.viewport.YOffset = 0

				// Previous chapter
				chapter := r.first.Chapter

				if r.books == nil {
					var err error
					r.books, err = r.searcher.Booklist()
					if err != nil {
						e := err.Error()
						r.viewport.SetContent(style.ErrorStyle.Render(e))
						return r, nil
					}
				}

				// Handle being beginning of book
				book := r.first.Book
				if chapter == 1 {
					index := slices.IndexFunc(r.books, func(b model.Book) bool {
						return b.Name == r.first.Book
					})
					if index == -1 {
						e := "error finding current book in booklist: not found"
						r.viewport.SetContent(style.ErrorStyle.Render(e))
						return r, nil
					} else if index == 0 {
						book = r.books[len(r.books)-1].Name
						chapter = r.books[len(r.books)-1].Chapters + 1
					} else {
						book = r.books[index-1].Name
						chapter = r.books[index-1].Chapters + 1
					}
				}
				content, err := r.Query(book + " " + strconv.Itoa(chapter-1))
				if err != nil {
					e := err.Error()
					r.viewport.SetContent(style.ErrorStyle.Render(e))
					return r, nil
				}

				r.viewport.SetContent(content)
				return r, tea.SetWindowTitle(r.query)
			case "n":
				r.viewport.YOffset = 0

				// Next chapter
				chapter := r.last.Chapter

				if r.books == nil {
					var err error
					r.books, err = r.searcher.Booklist()
					if err != nil {
						e := err.Error()
						r.viewport.SetContent(style.ErrorStyle.Render(e))
						return r, nil
					}
				}

				index := slices.IndexFunc(r.books, func(b model.Book) bool {
					return b.Name == r.last.Book
				})
				if index == -1 {
					e := "error finding current book in booklist: not found"
					r.viewport.SetContent(style.ErrorStyle.Render(e))
					return r, nil
				}

				// Handle being end of book
				book := r.last.Book
				if chapter == r.books[index].Chapters {
					if index == len(r.books)-1 {
						book = r.books[0].Name
						chapter = 0
					} else {
						book = r.books[index+1].Name
						chapter = 0
					}
				}
				content, err := r.Query(book + " " + strconv.Itoa(chapter+1))
				if err != nil {
					e := err.Error()
					r.viewport.SetContent(style.ErrorStyle.Render(e))
					return r, nil
				}

				r.viewport.SetContent(content)
				return r, tea.SetWindowTitle(r.query)
			case "/":
				r.mode = searching
			case "?":
				r.mode = help
			}
		} else if r.mode == searching {
			switch msg.String() {
			case "esc":
				r.mode = read
				r.searchbuffer = ""
			case "ctrl+c":
				return r, tea.Quit
			case "enter":
				r.viewport.YOffset = 0

				content, err := r.Query(r.searchbuffer)
				if err != nil {
					e := err.Error()
					r.viewport.SetContent(style.ErrorStyle.Render(e))
					r.searchbuffer = ""
					r.mode = read
					return r, nil
				}
				r.viewport.SetContent(content)

				title := r.searchbuffer
				r.searchbuffer = ""
				r.mode = read
				return r, tea.SetWindowTitle(title)
			case "backspace":
				if len(r.searchbuffer) > 0 {
					r.searchbuffer = r.searchbuffer[:len(r.searchbuffer)-1]
				}
			default:
				runes := []rune(msg.String())
				if len(runes) == 1 && utf8.ValidRune(runes[0]) {
					r.searchbuffer += string(runes[0])
				}
			}
		} else if r.mode == help {
			switch msg.String() {
			case "esc", "q":
				r.mode = read
			case "ctrl+c":
				return r, tea.Quit
			}
		} else {
			panic("Invalid mode")
		}
	case tea.WindowSizeMsg:
		if !r.ready {
			r.viewport = viewport.New(msg.Width, msg.Height-2)

			content, err := r.Query(r.query)
			if err != nil {
				e := err.Error()
				r.viewport.SetContent(style.ErrorStyle.Render(e))
				return r, nil
			}

			r.viewport.Style = r.viewport.Style.Padding(0, r.padding)
			r.viewport.SetContent(content)

			r.ready = true
		} else {
			r.viewport.Width = msg.Width
			r.viewport.Height = msg.Height - 2
		}
	}

	var cmd tea.Cmd
	if r.mode == read {
		r.viewport, cmd = r.viewport.Update(msg)
	}
	return r, cmd
}

const helptext = "q/esc: quit\n\ng/G: top/bottom\n\np/n: prev/next chapter\n\n+/-: increase/decrease padding\n\nw: toggle wrap\n\n/: search\n\n?: help\n\n"

func (r *Reader) Header() string {
	return style.HeaderStyle.Width(r.viewport.Width-(2*r.padding)).Margin(0, r.padding).Render(r.searcher.Translation())
}

func (r *Reader) Footer() string {
	if r.mode == searching {
		return style.SearchStyle.Padding(0, r.padding).Render("/" + r.searchbuffer)
	}
	return ""
}

func (r *Reader) View() string {
	if !r.ready {
		return "\n  Initializing..."
	}

	if r.mode == help {
		var writer strings.Builder
		writer.WriteString(helptext)
		writer.WriteString(r.viewport.KeyMap.Up.Help().Key + ": " + r.viewport.KeyMap.Up.Help().Desc + "\n\n")
		writer.WriteString(r.viewport.KeyMap.Down.Help().Key + ": " + r.viewport.KeyMap.Down.Help().Desc + "\n\n")
		writer.WriteString(r.viewport.KeyMap.PageUp.Help().Key + ": " + r.viewport.KeyMap.PageUp.Help().Desc + "\n\n")
		writer.WriteString(r.viewport.KeyMap.PageDown.Help().Key + ": " + r.viewport.KeyMap.PageDown.Help().Desc + "\n\n")
		writer.WriteString(r.viewport.KeyMap.HalfPageUp.Help().Key + ": " + r.viewport.KeyMap.HalfPageUp.Help().Desc + "\n\n")
		writer.WriteString(r.viewport.KeyMap.HalfPageDown.Help().Key + ": " + r.viewport.KeyMap.HalfPageDown.Help().Desc + "\n\n")

		hpad := (r.viewport.Width - lipgloss.Width(writer.String())) / 2
		vpad := (r.viewport.Height - lipgloss.Height(writer.String())) / 2

		return fmt.Sprintf(
			"%s\n%s\n%s",
			r.Header(),
			style.HelpStyle.Padding(vpad, hpad).Render(writer.String()),
			r.Footer(),
		)
	}
	return fmt.Sprintf(
		"%s\n%s\n%s",
		r.Header(),
		r.viewport.View(),
		r.Footer(),
	)
}
