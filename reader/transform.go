package reader

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func ResizeString(s string, width int, indentation string) []string {
	words := strings.Split(s, " ")
	var chunks []string
	var current []string
	var ccount int

	for _, word := range words {
		var wsize, _ = lipgloss.Size(word)
		var size = width
		if len(chunks) > 0 {
			size -= lipgloss.Width(indentation)
		}

		if ccount+wsize > size {
			if len(chunks) > 0 {
				current[0] = indentation + current[0]
			}
			chunks = append(chunks, strings.Join(current, " "))

			ccount = 0
			current = nil
		}

		ccount += wsize + 1
		current = append(current, word)
	}

	if len(chunks) > 0 {
		current[0] = indentation + current[0]
	}
	chunks = append(chunks, strings.Join(current, " "))

	return chunks
}
