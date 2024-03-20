package search

import (
	"github.com/woodywood117/bgate/model"
)

type Searcher interface {
	Query(translation, query string) ([]model.Verse, error)
	Booklist(translation string) ([]model.Book, error)
}
