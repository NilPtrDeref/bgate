package search

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type tokentype int

var (
	token_number tokentype = 0
	token_word   tokentype = 1
	token_colon  tokentype = 2
	token_dash   tokentype = 3
	// TODO:
	// token_comma tokentype = 4
	// TODO:
	// token_semicolon tokentype = 5
)

type token struct {
	_type tokentype
	value string
}

func tokenize(query string) ([]token, error) {
	var tokens []token
	runes := []rune(query)
	for i := 0; i < len(runes); i++ {
		if unicode.IsSpace(runes[i]) {
			continue
		} else if unicode.IsDigit(runes[i]) {
			num := []rune{runes[i]}
			for i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
				num = append(num, runes[i+1])
				i++
			}
			tokens = append(tokens, token{_type: token_number, value: string(num)})
		} else if unicode.IsLetter(runes[i]) {
			word := []rune{runes[i]}
			for i+1 < len(runes) && unicode.IsLetter(runes[i+1]) {
				word = append(word, runes[i+1])
				i++
			}
			tokens = append(tokens, token{_type: token_word, value: string(word)})
		} else if runes[i] == ':' {
			tokens = append(tokens, token{_type: token_colon, value: ":"})
		} else if runes[i] == '-' {
			tokens = append(tokens, token{_type: token_dash, value: "-"})
		} else {
			return nil, errors.New("Invalid character when tokenizing query")
		}
	}
	return tokens, nil
}

func parsebook(tokens []token) (string, []token, error) {
	book := ""

	if len(tokens) == 0 {
		return book, tokens, errors.New("No book found")
	}

	if tokens[0]._type == token_word {
		book = tokens[0].value
		if book, ok := abbreviations[book]; ok {
			return book, tokens[1:], nil
		}
	}
	if tokens[0]._type == token_number {
		book = tokens[0].value

		if !(len(tokens) > 1 && tokens[1]._type == token_word) {
			return book, tokens, errors.New("invalid token in book parsing")
		}

		book += tokens[1].value
		if book, ok := abbreviations[book]; ok {
			return book, tokens[2:], nil
		}
	}

	return book, tokens, errors.New("invalid token in book parsing")
}

func parsechapter(tokens []token) (string, []token, error) {
	if len(tokens) == 0 {
		return "", tokens, errors.New("No chapter found")
	}
	if tokens[0]._type != token_number {
		return "", tokens, errors.New("Invalid chapter")
	}
	return tokens[0].value, tokens[1:], nil
}

func parseverse(tokens []token) (string, []token, error) {
	if len(tokens) == 0 {
		return "", tokens, nil
	}
	if tokens[0]._type != token_colon {
		return "", tokens, nil
	}
	if len(tokens) == 1 {
		return "", tokens, errors.New("No verse found")
	}
	if tokens[1]._type != token_number {
		return "", tokens, errors.New("Invalid verse")
	}
	return tokens[1].value, tokens[2:], nil
}

func parsepart(tokens []token) (string, []token, error) {
	book, tokens, err := parsebook(tokens)
	if err != nil {
		return "", tokens, err
	}

	chapter, tokens, err := parsechapter(tokens)
	if err != nil {
		return "", tokens, err
	}
	part := fmt.Sprintf("book = '%s' and chapter = %s", book, chapter)

	verse, tokens, err := parseverse(tokens)
	if err != nil {
		return "", tokens, err
	}
	if verse != "" {
		part += fmt.Sprintf(" and number = %s", verse)
	}

	if len(tokens) > 0 && tokens[0]._type == token_dash {
		tokens = tokens[1:]
		vrange := fmt.Sprintf("id >= (select id from verses where %s order by id limit 1)", part)

		var newbook string
		newbook, tokens, err = parsebook(tokens)
		if err != nil {
			newbook = book
		}

		var newchapter string
		newchapter, tokens, err = parsechapter(tokens)
		if err != nil {
			return "", tokens, errors.Join(errors.New("failed in second chapter parse"), err)
		}

		var newverse string
		newverse, tokens, err = parseverse(tokens)
		if err != nil {
			return "", tokens, err
		}

		if verse != "" && newverse != "" {
			// Range across chapters possibly
			part = fmt.Sprintf("book = '%s' and chapter = %s and number = %s", newbook, newchapter, newverse)
		} else if verse != "" && newverse == "" {
			// Continuation of chapter verse->verse
			part = fmt.Sprintf("book = '%s' and chapter = %s and number = %s", newbook, chapter, newchapter)
		} else {
			part = fmt.Sprintf("book = '%s' and chapter = %s", newbook, newchapter)
			if newverse != "" {
				part += fmt.Sprintf(" and number = %s", newverse)
			}
		}

		vrange += fmt.Sprintf(" and id <= (select id from verses where %s order by id desc limit 1)", part)
		return vrange, tokens, nil
	}

	return part, tokens, nil
}

func parsequery(query string) (string, error) {
	query = strings.TrimSpace(query)
	query = strings.ToLower(query)

	tokens, err := tokenize(query)
	if err != nil {
		return "", err
	}

	part, tokens, err := parsepart(tokens)
	if err != nil {
		return "", err
	}

	return part, nil
}
