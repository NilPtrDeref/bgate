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
		viper.BindPFlag("force-local", cmd.Flag("force-local"))
		viper.BindPFlag("force-remote", cmd.Flag("force-remote"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		translation := viper.GetString("translation")
		query := args[0]
		interactive := viper.GetBool("interactive")
		padding := viper.GetInt("padding")
		wrap := viper.GetBool("wrap")

		local, err := search.TranslationHasLocal(translation)
		cobra.CheckErr(err)

		if !local && viper.GetBool("force-local") {
			cobra.CheckErr(errors.New("No local copy of translation found. Please use download command for requested translation."))
		}

		var searcher search.Searcher
		if local && !viper.GetBool("force-remote") {
			searcher, err = search.NewLocal(translation)
			cobra.CheckErr(err)
		} else {
			searcher = search.NewRemote(translation)
		}

		r := view.NewReader(searcher, query, wrap, padding)
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

		p := tea.NewProgram(r, tea.WithMouseAllMotion())
		p.SetWindowTitle(query)
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
			os.Exit(1)
		}
		if r.Error != nil {
			fmt.Fprintf(os.Stderr, "Error running program: %v\n", r.Error)
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
	root.Flags().StringP("translation", "t", "ESV", "The translation of the Bible to search for.")
	root.Flags().BoolP("interactive", "i", false, "Interactive view, allows you to scroll using j/up and k/down.")
	root.Flags().IntP("padding", "p", 0, "Horizontal padding in character count.")
	root.Flags().BoolP("wrap", "w", false, "Wrap verses, this will cause it to not start each verse on a new line.")
	root.Flags().Bool("force-local", false, "Force the program to crash if there isn't a local copy of the translation you're trying to read.")
	root.Flags().Bool("force-remote", false, "Force the program to use the remote searcher even if there is a local copy of the translation.")

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
