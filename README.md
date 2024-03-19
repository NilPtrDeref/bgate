# bgate
```
A terminal interface to Bible Gateway

Usage:
  bgate [flags] <query>
  bgate [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  list        List all books of the Bible and how many chapters they have.

Flags:
  -c, --config string        Config file to use. (default "~/.config/bgate/config.json")
  -h, --help                 help for bgate
  -i, --interactive          Interactive view, allows you to scroll using j/up and k/down.
  -p, --padding int          Horizontal padding in character count.
  -t, --translation string   The translation of the Bible to search for.

Use "bgate [command] --help" for more information about a command.
```

## Interactive Controls
* `j` - Down
* `k` - Up
* `g` - Top
* `G` - Bottom
* `q/esc/ctrl+c` - Quit
