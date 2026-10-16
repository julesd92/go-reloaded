package main

import "testing"

func TestProcess(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"hex", "1E (hex) files were added", "30 files were added"},
		{"bin", "It has been 10 (bin) years", "It has been 2 years"},
		{"up", "Ready, set, go (up) !", "Ready, set, GO!"},
		{"low", "I should stop SHOUTING (low)", "I should stop shouting"},
		{"cap", "Welcome to the Brooklyn bridge (cap)", "Welcome to the Brooklyn Bridge"},
		{"up avec nombre", "This is so exciting (up, 2)", "This is SO EXCITING"},
		{"nombre trop grand", "so exciting (up, 10)", "SO EXCITING"},
		{"ponctuation", "I was sitting over there ,and then BAMM !!", "I was sitting over there, and then BAMM!!"},
		{"groupe ...", "I was thinking ... You were right", "I was thinking... You were right"},
		{"apostrophes 1 mot", "I am exactly how they describe me: ' awesome '", "I am exactly how they describe me: 'awesome'"},
		{"apostrophes plusieurs mots",
			"As Elton John said: ' I am the most well-known homosexual in the world '",
			"As Elton John said: 'I am the most well-known homosexual in the world'"},
		{"a -> an", "There it was. A amazing rock!", "There it was. An amazing rock!"},
		{"a -> an devant h", "bearing a honest man", "bearing an honest man"},
		{"a reste a", "a good day", "a good day"},
		{"hex + bin", "Simply add 42 (hex) and 10 (bin) and you will see the result is 68.",
			"Simply add 66 and 2 and you will see the result is 68."},
		{"exemple 1 du sujet",
			"it (cap) was the best of times, it was the worst of times (up) , it was the age of wisdom, it",
			"It was the best of times, it was the worst of TIMES, it was the age of wisdom, it"},
		{"exemple ponctuation", "Punctuation tests are ... kinda boring ,what do you think ?",
			"Punctuation tests are... kinda boring, what do you think?"},
		{"decimal conservé", "pi is 3.14 ok", "pi is 3.14 ok"},
		{"apostrophe dans un mot", "don't stop", "don't stop"},
		{"plusieurs lignes", "hello (up)\nworld (cap)", "HELLO\nWorld"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Process(tc.in); got != tc.want {
				t.Errorf("\nentrée  : %q\nattendu : %q\nobtenu  : %q", tc.in, tc.want, got)
			}
		})
	}
}
