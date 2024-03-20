package search

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/jmoiron/sqlx"
)

func TranslationHasLocal(translation string) (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}

	bgatepath := path.Join(home, ".bgate")
	sqlpath := path.Join(bgatepath, fmt.Sprintf("%s.sql", translation))

	_, err = os.Stat(sqlpath)
	return errors.Is(err, os.ErrNotExist), nil
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

// func (l *Local) Query(translation, query string) ([]model.Verse, error) {
// }
//
// func (l *Local) Booklist(translation string) ([]model.Book, error) {
// }

func (l *Local) Close() error {
	return l.db.Close()
}
