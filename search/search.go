package search

import (
	"github.com/nilptrderef/bgate/reader/model"
)

type Searcher interface {
	Query(query string) ([]model.Verse, error)
	Booklist() ([]model.Book, error)
	Translation() string
}
