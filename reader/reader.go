package reader

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/woodywood117/bgate/reader/model"
	"github.com/woodywood117/bgate/reader/style"
	"github.com/woodywood117/bgate/search"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type reader_state int

const (
	reading reader_state = iota
	searching
	help
)

type Reader struct {
	vwidth    int
	vheight   int
	scroll    int
	maxscroll int
	wrap      bool
	padding   int

	searcher     search.Searcher
	query        string
	verses       []model.Verse
	lines        []string
	books        []model.Book
	searchbuffer string

	state reader_state
	error error
}

func NewReader(searcher search.Searcher, width, height int) *Reader {
	return &Reader{
		searcher: searcher,
		vwidth:   width,
		vheight:  height,
	}
}

func (r *Reader) SetWindowSize(width, height int) {
	r.vheight = height
	r.vwidth = width

	r.ResizeText()

	r.maxscroll = max(0, len(r.lines)-1)
	r.scroll = min(r.scroll, r.maxscroll)
}

func (r *Reader) SetWrap(wrap bool) {
	r.wrap = wrap
	r.ResizeText()
}

func (r *Reader) SetPadding(padding int) {
	r.padding = padding
	r.ResizeText()
}

func (r *Reader) SetQuery(query string) error {
	r.query = query

	if r.query != "" {
		r.verses, r.error = r.searcher.Query(query)
		if r.error != nil {
			r.lines = nil
			return r.error
		}

		r.ResizeText()
	}

	return nil
}

func (r *Reader) GetError() error {
	return r.error
}

func (r *Reader) Init() tea.Cmd {
	err := r.SetQuery(r.query)
	r.error = err
	return nil
}

func (r *Reader) ResizeText() {
	width := r.vwidth - 2*r.padding
	lines := []string{}
	if len(r.verses) == 0 {
		r.error = fmt.Errorf("no results found for %q", r.query)
		r.lines = nil
		return
	}

	for i := 0; i < len(r.verses); i++ {
		current := r.verses[i]
		if current.HasTitle() {
			lines = append(lines, current.TitleString())
		}

		if current.Number == 1 && current.Part == 1 {
			lines = append(lines, current.ChapterString())
		}

		var line string
		if current.Part > 1 {
			line = "    " + current.Text
		} else {
			line = current.NumberString() + current.Text
		}

		if r.wrap && current.Part == 1 {
			for {
				if i+1 >= len(r.verses) || r.verses[i+1].HasTitle() || r.verses[i+1].Part > 1 {
					break
				}

				current = r.verses[i+1]
				line = strings.Join([]string{line, current.NumberString() + current.Text}, " ")
				i++
			}
		}

		indentation := ""
		if !r.wrap {
			indentation = "    "
		}

		chunked := ResizeString(line, width, indentation)
		lines = append(lines, chunked...)
	}

	r.lines = lines
}

func (r *Reader) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if r.state == reading {
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
				r.scroll = max(0, (r.maxscroll-r.vheight)+2)
			case "+":
				r.padding++
				r.ResizeText()
			case "-":
				r.padding = max(0, r.padding-1)
				r.ResizeText()
			case "p":
				// Previous chapter
				first := r.verses[0]
				chapter := first.Chapter

				if r.books == nil {
					r.books, r.error = r.searcher.Booklist()
					if r.error != nil {
						return r, nil
					}
				}

				// Handle being beginning of book
				book := first.Book
				if chapter == 1 {
					index := slices.IndexFunc(r.books, func(b model.Book) bool {
						return b.Name == first.Book
					})
					if index == -1 {
						r.error = errors.New("error finding current book in booklist: not found")
						return r, nil
					} else if index == 0 {
						book = r.books[len(r.books)-1].Name
						chapter = r.books[len(r.books)-1].Chapters + 1
					} else {
						book = r.books[index-1].Name
						chapter = r.books[index-1].Chapters + 1
					}
				}
				r.error = r.SetQuery(book + " " + strconv.Itoa(chapter-1))
				if r.error != nil {
					return r, nil
				}
				return r, tea.SetWindowTitle(r.query)
			case "n":
				// Next chapter
				last := r.verses[len(r.verses)-1]
				chapter := last.Chapter

				if r.books == nil {
					r.books, r.error = r.searcher.Booklist()
					if r.error != nil {
						return r, nil
					}
				}

				index := slices.IndexFunc(r.books, func(b model.Book) bool {
					return b.Name == last.Book
				})
				if index == -1 {
					r.error = errors.New("error when finding current book in booklist: not found")
					return r, nil
				}

				// Handle being end of book
				book := last.Book
				if chapter == r.books[index].Chapters {
					if index == len(r.books)-1 {
						book = r.books[0].Name
						chapter = 0
					} else {
						book = r.books[index+1].Name
						chapter = 0
					}
				}
				r.error = r.SetQuery(book + " " + strconv.Itoa(chapter+1))
				if r.error != nil {
					return r, nil
				}
				return r, tea.SetWindowTitle(r.query)
			case "/":
				r.state = searching
			case "?":
				r.state = help
			}
		} else if r.state == searching {
			switch msg.String() {
			case "esc":
				r.state = reading
				r.searchbuffer = ""
			case "ctrl+c":
				return r, tea.Quit
			case "enter":
				r.SetQuery(r.searchbuffer)
				title := r.searchbuffer
				r.searchbuffer = ""
				r.state = reading
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
		} else if r.state == help {
			switch msg.String() {
			case "esc", "q":
				r.state = reading
			case "ctrl+c":
				return r, tea.Quit
			}
		} else {
			panic("Invalid state")
		}
	case tea.MouseMsg:
		switch msg.String() {
		case "wheel down":
			if r.scroll < r.maxscroll {
				r.scroll += min(3, r.maxscroll-r.scroll)
			}
		case "wheel up":
			if r.scroll > 0 {
				r.scroll -= min(3, r.scroll)
			}
		}
	case tea.WindowSizeMsg:
		r.SetWindowSize(msg.Width, msg.Height)
	}
	return r, nil
}

const helptext = "q/esc: quit\n\nj/k or up/down: scroll\n\ng/G: top/bottom\n\np/n: prev/next chapter\n\n/: search\n\n?: help"

func (r *Reader) View() string {
	var view strings.Builder
	lpad := strings.Repeat(" ", r.padding)

	if r.state == help {
		hpad := (r.vwidth - lipgloss.Width(helptext)) / 2
		vpad := (r.vheight - lipgloss.Height(helptext)) / 2
		return style.SearchStyle.PaddingTop(vpad).PaddingLeft(hpad).Render(helptext)
	}

	if r.error != nil {
		view.WriteString(lpad + style.ErrorStyle.Render(r.error.Error()) + "\n")
	} else {
		for i := 0; i < r.vheight-1; i++ {
			if r.scroll+i >= len(r.lines) {
				break
			}
			view.WriteString(lpad + r.lines[r.scroll+i] + "\n")
		}
	}

	output := view.String()
	if r.state == searching {
		split := strings.Split(output, "\n")
		for len(split) < r.vheight-1 {
			split = append(split, "")
		}
		split[len(split)-1] = lpad + style.SearchStyle.Render("/"+r.searchbuffer)
		output = strings.Join(split, "\n")
	}

	return output
}
