package search

import (
	"github.com/woodywood117/bgate/reader/model"
)

type Searcher interface {
	Query(query string) ([]model.Verse, error)
	Booklist() ([]model.Book, error)
}
