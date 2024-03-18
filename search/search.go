package search

import (
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

func Passage(translation, query string) (*goquery.Document, error) {
	// URL format: https://www.biblegateway.com/passage/?search=Genesis+1&version=LSB
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

	// TODO: Check for status code

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	return document, nil
}
