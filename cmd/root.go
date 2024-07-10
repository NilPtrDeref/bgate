package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/nilptrderef/bgate/reader"
	"github.com/nilptrderef/bgate/search"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var root = &cobra.Command{
	Use:   "bgate [flags] <query>",
	Short: "A terminal interface to Bible Gateway",
	Args:  cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("translation", cmd.Flag("translation"))
		viper.BindPFlag("padding", cmd.Flag("padding"))
		viper.BindPFlag("wrap", cmd.Flag("wrap"))
		viper.BindPFlag("force-local", cmd.Flag("force-local"))
		viper.BindPFlag("force-remote", cmd.Flag("force-remote"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		translation := viper.GetString("translation")
		query := strings.Join(args, " ")
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

		r := reader.NewReader(searcher, query)
		r.SetPadding(padding)
		r.SetWrap(wrap)

		p := tea.NewProgram(r, tea.WithMouseCellMotion(), tea.WithAltScreen())
		p.SetWindowTitle(query)
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
	root.Flags().StringP("translation", "t", "ESV", "The translation of the Bible to search for.")
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
