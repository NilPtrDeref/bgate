package search

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/jmoiron/sqlx"
	"github.com/woodywood117/bgate/model"
)

func TranslationHasLocal(translation string) (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}

	bgatepath := path.Join(home, ".bgate")
	sqlpath := path.Join(bgatepath, fmt.Sprintf("%s.sql", translation))

	_, err = os.Stat(sqlpath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	return !errors.Is(err, os.ErrNotExist), nil
}

type Local struct {
	db *sqlx.DB
}

func NewLocal(translation string) (*Local, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	bgatepath := path.Join(home, ".bgate")
	sqlpath := path.Join(bgatepath, fmt.Sprintf("%s.sql", translation))
	db, err := sqlx.Open("sqlite3", sqlpath)
	if err != nil {
		return nil, err
	}
	return &Local{db}, nil
}

func (l *Local) Query(query string) ([]model.Verse, error) {
	query, err := parsequery(query)
	if err != nil {
		return nil, err
	}
	query = fmt.Sprintf("SELECT book, chapter, number, part, text, title FROM verses WHERE %s", query)

	var verses []model.Verse
	err = l.db.Select(&verses, query)
	if err != nil {
		return nil, err
	}

	return verses, nil
}

func (l *Local) Booklist() ([]model.Book, error) {
	var books []model.Book
	err := l.db.Select(&books, "SELECT distinct(book) as name, max(chapter) as chapters FROM verses group by book order by id")
	if err != nil {
		return nil, err
	}
	return books, nil
}

func (l *Local) Close() error {
	return l.db.Close()
}
