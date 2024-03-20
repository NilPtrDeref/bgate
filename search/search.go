package search

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/woodywood117/bgate/model"
)

/// URL format: https://www.biblegateway.com/passage/?search=Genesis+1&version=LSB

func Passage(translation, query string) ([]model.Content, error) {
	base, err := url.Parse("https://www.biblegateway.com/passage/")
	if err != nil {
		return nil, err
	}

	values := base.Query()
	values.Set("search", query)
	values.Set("version", translation)
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

	content := []model.Content{}
	document.Find(".passage-content").Each(func(pi int, passage *goquery.Selection) {
		passage.Find(".text").Each(func(li int, line *goquery.Selection) {
			if strings.HasPrefix(line.Parent().Nodes[0].Data, "h") {
				content = append(content, model.Content{
					Type:    model.Section,
					Content: line.Text(),
				})
				return
			}

			chapter := line.Find(".chapternum")
			if chapter.Length() > 0 {
				c := model.Content{
					Type:   model.Chapter,
					Number: chapter.Text(),
				}
				chapter.Remove()
				content = append(content, c)

				c.Type = model.Verse
				c.Number = "1 "
				c.Content = line.Text()
				content = append(content, c)
				return
			}

			verse := line.Find(".versenum")
			if verse.Length() > 0 {
				c := model.Content{
					Type:   model.Verse,
					Number: verse.Text(),
				}
				verse.Remove()

				c.Content = line.Text()
				content = append(content, c)
				return
			}

			content[len(content)-1].Content += " " + line.Text()
		})
	})

	return content, nil
}

func Booklist(translation string) ([]model.Book, error) {
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
		values.Set("version", translation)
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
