package cmd

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nilptrderef/bgate/search"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var download = &cobra.Command{
	Use:   "download",
	Short: "Download a translation of the Bible for local usage rather than reaching out to BibleGateway",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("translation", cmd.Flag("translation"))
		viper.BindPFlag("delay", cmd.Flag("delay"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		translation := viper.GetString("translation")
		delay := viper.GetInt("delay")

		remote := search.NewRemote(translation)

		books, err := remote.Booklist()
		cobra.CheckErr(err)
		if len(books) == 0 {
			cobra.CheckErr(fmt.Errorf("No books found for translation: %s", translation))
		}

		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		bgatepath := path.Join(home, ".bgate")
		err = os.MkdirAll(bgatepath, 0755)
		cobra.CheckErr(err)

		sqlpath := path.Join(bgatepath, fmt.Sprintf("%s.sql", translation))
		os.Remove(sqlpath)

		db, err := sqlx.Open("sqlite3", sqlpath)
		cobra.CheckErr(err)
		defer db.Close()

		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS verses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			book TEXT,
			chapter INTEGER,
			number INTEGER,
			part INTEGER,
			text TEXT,
			title TEXT
		)`)
		cobra.CheckErr(err)

		for _, book := range books {
			fmt.Printf("Downloading %s...\n", book.Name)

			for chapter := range book.Chapters {
				verses, err := remote.Query(fmt.Sprintf("%s %d", book.Name, chapter+1))
				cobra.CheckErr(err)

				for _, verse := range verses {
					_, err = db.Exec("insert into verses (book, chapter, number, part, text, title) values (?, ?, ?, ?, ?, ?)", book.Name, verse.Chapter, verse.Number, verse.Part, verse.Text, verse.Title)
					cobra.CheckErr(err)
				}

				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
		}
	},
}

func init() {
	download.Flags().StringP("translation", "t", "ESV", "The translation of the Bible to search for.")
	download.Flags().IntP("delay", "d", 100, "Number of milliseconds to wait between requests.")
	root.AddCommand(download)
}
