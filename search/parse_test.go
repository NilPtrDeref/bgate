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
	}

	for _, test := range tests {
		tokens, err := tokenize(test.query)
		if err != test.err {
			t.Errorf("Expected error %v, got %v", test.err, err)
		}
		if len(tokens) != len(test.tokens) {
			t.Errorf("Expected %d tokens, got %d", len(test.tokens), len(tokens))
		}
		for i := range tokens {
			if tokens[i] != test.tokens[i] {
				t.Errorf("Expected token %v, got %v", test.tokens[i], tokens[i])
			}
		}
	}
}
