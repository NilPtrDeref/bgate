package view

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/woodywood117/bgate/model"
	"github.com/woodywood117/bgate/search"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Reader struct {
	searcher search.Searcher
	query    string // Only the initial query, should never be written to after intialization

	verses    []model.Verse
	wrap      bool
	padding   int
	lines     []string
	scroll    int
	maxscroll int
	vheight   int
	vwidth    int
	books     []model.Book
	Error     error
}

func NewReader(searcher search.Searcher, query string, wrap bool, padding int) *Reader {
	return &Reader{
		searcher: searcher,
		query:    query,
		wrap:     wrap,
		padding:  padding,
		scroll:   0,
		vheight:  20,
		vwidth:   20,
	}
}

func (r *Reader) Init() tea.Cmd {
	err := r.ChangePassage(r.query)
	if err != nil {
		r.Error = err
		return tea.Quit
	}
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

		chunked := r.chunks(line, width)
		lines = append(lines, chunked...)
	}

	r.lines = lines
}

func (r *Reader) ChangePassage(query string) (err error) {
	r.verses, err = r.searcher.Query(query)
	if err != nil {
		return err
	}

	if len(r.verses) == 0 {
		return errors.New("No verses found")
	}

	r.resize(r.vwidth - 2*r.padding)
	return nil
}

func (r *Reader) SetWindowSize(width, height int) {
	r.vheight = height

	r.vwidth = width
	r.resize(width - 2*r.padding)

	r.maxscroll = max(0, len(r.lines)-1)
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
			r.scroll = max(0, (r.maxscroll-r.vheight)+2)
		case "p":
			// Previous chapter
			first := r.verses[0]
			chapter := first.Chapter

			if r.books == nil {
				var err error
				r.books, err = r.searcher.Booklist()
				if err != nil {
					r.Error = err
					return r, tea.Quit
				}
			}

			// Handle being beginning of book
			book := first.Book
			if chapter == 1 {
				index := slices.IndexFunc(r.books, func(b model.Book) bool {
					return b.Name == first.Book
				})
				if index == -1 {
					r.Error = errors.New("Book not found")
					return r, tea.Quit
				} else if index == 0 {
					book = r.books[len(r.books)-1].Name
					chapter = r.books[len(r.books)-1].Chapters + 1
				} else {
					book = r.books[index-1].Name
					chapter = r.books[index-1].Chapters + 1
				}
			}
			query := book + " " + strconv.Itoa(chapter-1)
			err := r.ChangePassage(query)
			if err != nil {
				r.Error = err
				return r, tea.Quit
			}
			return r, tea.SetWindowTitle(query)
		case "n":
			// Next chapter
			last := r.verses[len(r.verses)-1]
			chapter := last.Chapter

			if r.books == nil {
				var err error
				r.books, err = r.searcher.Booklist()
				if err != nil {
					r.Error = err
					return r, tea.Quit
				}
			}

			index := slices.IndexFunc(r.books, func(b model.Book) bool {
				return b.Name == last.Book
			})
			if index == -1 {
				r.Error = errors.New("Book not found")
				return r, tea.Quit
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
			query := book + " " + strconv.Itoa(chapter+1)
			err := r.ChangePassage(query)
			if err != nil {
				r.Error = err
				return r, tea.Quit
			}
			return r, tea.SetWindowTitle(query)
		}
	case tea.MouseMsg:
		switch msg.String() {
		case "wheel down":
			if r.scroll < r.maxscroll {
				r.scroll++
			}
		case "wheel up":
			if r.scroll > 0 {
				r.scroll--
			}
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
