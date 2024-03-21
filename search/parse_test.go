package search

import "testing"

func TestTokenize(t *testing.T) {
	tests := []struct {
		query  string
		tokens []token
		err    error
	}{
		{"", []token{}, nil},
		{" ", []token{}, nil},
		{"1john1:1", []token{
			{token_number, "1"},
			{token_word, "john"},
			{token_number, "1"},
			{token_colon, ":"},
			{token_number, "1"},
		}, nil},
		{"1john1-2", []token{
			{token_number, "1"},
			{token_word, "john"},
			{token_number, "1"},
			{token_dash, "-"},
			{token_number, "2"},
		}, nil},
	}

	for _, test := range tests {
		tokens, err := tokenize(test.query)
		if err != test.err {
			t.Fatalf("Expected error %v, got %v", test.err, err)
		}
		if len(tokens) != len(test.tokens) {
			t.Fatalf("Expected %d tokens, got %d", len(test.tokens), len(tokens))
		}
		for i := range tokens {
			if tokens[i] != test.tokens[i] {
				t.Fatalf("Expected token %v, got %v", test.tokens[i], tokens[i])
			}
		}
	}
}

func TestParseQuery(t *testing.T) {
	output, err := parsequery("1john1:1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output != "book = '1 John' and chapter = 1 and number = 1" {
		t.Fatalf("Unexpected output: %s", output)
	}

	output, err = parsequery("1john1-2")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "id >= (select id from verses where book = '1 John' and chapter = 1 order by id limit 1) and id <= (select id from verses where book = '1 John' and chapter = 2 order by id desc limit 1)"
	if output != expected {
		t.Fatalf("Unexpected output:\nExpected: %s\nActual: %s", expected, output)
	}
}
