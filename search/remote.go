package search

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/woodywood117/bgate/reader/model"
)

type Remote struct {
	// URL format: https://www.biblegateway.com/passage/?search=Genesis+1&version=LSB
	translation string
}

func NewRemote(translation string) *Remote {
	return &Remote{translation}
}

func (r *Remote) Query(query string) ([]model.Verse, error) {
	base, err := url.Parse("https://www.biblegateway.com/passage/")
	if err != nil {
		return nil, err
	}

	values := base.Query()
	values.Set("search", query)
	values.Set("version", r.translation)
	base.RawQuery = values.Encode()

	request, err := http.NewRequest("GET", base.String(), nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("unable to retrieve passage")
	}

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}
	document.Find(".crossreference").Remove()
	document.Find(".footnote").Remove()

	verses := []model.Verse{}
	document.Find(".passage-table").Each(func(pi int, passage *goquery.Selection) {
		passage.Find(".translation").Remove()

		var book string
		{
			ptext := passage.Find(".dropdown-display-text").Text()
			psplit := strings.Split(ptext, " ")
			book = strings.Join(psplit[:len(psplit)-1], " ")
		}

		var title *string
		var part int
		passage.Find(".text").Each(func(li int, line *goquery.Selection) {
			// Store title for the next verse
			if strings.HasPrefix(line.Parent().Nodes[0].Data, "h") {
				t := line.Text()
				title = &t
				return
			}

			class, ok := line.Attr("class")
			if !ok {
				fmt.Println("No class on text line")
				os.Exit(1)
			}

			csplit := strings.Split(class, " ")
			if len(csplit) != 2 {
				fmt.Println("Unexpected class format")
				os.Exit(1)
			}

			csplit = strings.Split(csplit[1], "-")
			if len(csplit) != 3 {
				fmt.Println("Unexpected inner class format")
				os.Exit(1)
			}

			cnum, err := strconv.Atoi(csplit[1])
			if err != nil {
				fmt.Println("Failed to parse chapter number:", err)
				os.Exit(1)
			}
			vnum, err := strconv.Atoi(csplit[2])
			if err != nil {
				fmt.Println("Failed to parse verse number:", err)
				os.Exit(1)
			}

			if line.Find(".versenum").Remove().Length() > 0 || line.Find(".chapternum").Remove().Length() > 0 {
				verses = append(verses, model.Verse{
					Book:    book,
					Chapter: cnum,
					Number:  vnum,
					Part:    1,
					Text:    line.Text(),
					Title:   title,
				})
				part = 1
			} else {
				verses = append(verses, model.Verse{
					Book:    book,
					Chapter: cnum,
					Number:  vnum,
					Part:    part + 1,
					Text:    line.Text(),
					Title:   title,
				})
				part++
			}

			title = nil
		})
	})

	return verses, nil
}

func (r *Remote) Booklist() ([]model.Book, error) {
	var base = "https://www.biblegateway.com"
	var booklist string
	var books []model.Book

	{
		base, err := url.Parse(base + "/passage/")
		if err != nil {
			return nil, err
		}

		values := base.Query()
		values.Set("search", "Genesis 1")
		values.Set("version", r.translation)
		base.RawQuery = values.Encode()

		request, err := http.NewRequest("GET", base.String(), nil)
		if err != nil {
			return nil, err
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return nil, errors.New("unable to retrieve passage")
		}

		document, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			return nil, err
		}

		link := document.Find(".publisher-info-bottom").Find("a")
		if link.Length() == 0 {
			return nil, errors.New("unable to retrieve book list")
		}

		booklist, _ = link.First().Attr("href")
		booklist = booklist + "#booklist"
	}

	{
		request, err := http.NewRequest("GET", base+booklist, nil)
		if err != nil {
			return nil, err
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return nil, errors.New("unable to retrieve booklist")
		}

		document, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			return nil, err
		}

		rows := document.Find(".infotable").Find("tr").Find(".book-name")
		rows.Find("svg").Remove()
		rows.Each(func(i int, row *goquery.Selection) {
			ctext := row.Find(".num-chapters").Text()
			row.Find(".num-chapters").Remove()
			chapters, ierr := strconv.Atoi(ctext)
			if ierr != nil {
				err = ierr
				return
			}

			name := strings.TrimSpace(row.Text())
			books = append(books, model.Book{
				Name:     name,
				Chapters: chapters,
			})
		})
		if err != nil {
			return nil, err
		}
	}

	return books, nil
}

func (r *Remote) Translation() string {
	return r.translation
}
