package cmd

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/woodywood117/bgate/search"
	"github.com/woodywood117/bgate/view"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var root = &cobra.Command{
	Use:   "bgate [flags] <query>",
	Short: "A terminal interface to Bible Gateway",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("translation", cmd.Flag("translation"))
		viper.BindPFlag("interactive", cmd.Flag("interactive"))
		viper.BindPFlag("padding", cmd.Flag("padding"))
		viper.BindPFlag("wrap", cmd.Flag("wrap"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		translation := viper.GetString("translation")
		query := args[0]
		interactive := viper.GetBool("interactive")
		padding := viper.GetInt("padding")
		wrap := viper.GetBool("wrap")

		verses, err := search.Query(translation, query)
		cobra.CheckErr(err)

		if len(verses) == 0 {
			cobra.CheckErr(errors.New("No content found"))
		}

		r := view.NewReader(verses, wrap, padding)
		if !interactive {
			width, _, err := term.GetSize(0)
			if err != nil {
				panic(err)
			}
			r.SetWindowSize(width, math.MaxInt32)
			v := r.View()
			fmt.Print(v)
			return
		}

		p := tea.NewProgram(r)
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	var config string
	root.PersistentFlags().StringVarP(&config, "config", "c", "~/.config/bgate/config.json", "Config file to use.")
	root.Flags().StringP("translation", "t", "", "The translation of the Bible to search for.")
	root.Flags().BoolP("interactive", "i", false, "Interactive view, allows you to scroll using j/up and k/down.")
	root.Flags().IntP("padding", "p", 0, "Horizontal padding in character count.")
	root.Flags().BoolP("wrap", "w", false, "Wrap verses, this will cause it to not start each verse on a new line.")

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	config = strings.ReplaceAll(config, "~", home)
	viper.SetConfigFile(config)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			panic(err)
		}
	}
}
