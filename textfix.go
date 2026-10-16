package main

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// kind décrit la nature d'un morceau de texte (token).
type kind int

const (
	word    kind = iota // un mot : "hello", "1E", "well-known"
	punct               // un groupe de ponctuation : ",", "...", "!?"
	quote               // une apostrophe isolée ' (ouvrante ou fermante)
	command             // une commande normalisée : "(up,2)", "(hex,)"
)

type token struct {
	kind kind
	text string
}

var (
	// Repère une commande dans le texte brut : (up), (low, 3), (hex)...
	cmdRe = regexp.MustCompile(`\(\s*(hex|bin|up|low|cap)\s*(?:,\s*(\d+)\s*)?\)`)
	// Reconnaît une commande une fois normalisée : "(up,2)" ou "(up,)".
	tokRe = regexp.MustCompile(`^\((hex|bin|up|low|cap),(\d*)\)$`)
)

// Process applique toutes les transformations au texte, ligne par ligne
// (les retours à la ligne sont donc conservés).
func Process(text string) string {
	text = strings.ReplaceAll(text, "\r", "")
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = processLine(line)
	}
	return strings.Join(lines, "\n")
}

func processLine(line string) string {
	// "(up, 2)" -> " (up,2) " : la commande devient un bloc sans espace interne.
	line = cmdRe.ReplaceAllString(line, " ($1,$2) ")

	tokens := tokenize(line)
	tokens = applyCommands(tokens)
	fixArticles(tokens)
	return render(tokens)
}

// ---------------------------------------------------------------------------
// 1. Découpage en tokens
// ---------------------------------------------------------------------------

func isPunct(r rune) bool { return strings.ContainsRune(".,!?:;", r) }

func isAlnum(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }

// tokenize découpe la ligne en mots, ponctuation, apostrophes et commandes.
func tokenize(line string) []token {
	var tokens []token
	for _, field := range strings.Fields(line) {
		if tokRe.MatchString(field) {
			tokens = append(tokens, token{command, field})
			continue
		}
		tokens = append(tokens, splitField(field)...)
	}
	return tokens
}

// splitField sépare la ponctuation et les apostrophes collées à un mot.
// Exemple : "times," -> "times" + "," ; "'awesome'" -> "'" + "awesome" + "'".
func splitField(field string) []token {
	runes := []rune(field)
	var tokens []token
	var buf []rune

	flush := func() {
		if len(buf) > 0 {
			tokens = append(tokens, token{word, string(buf)})
			buf = nil
		}
	}

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case isPunct(r):
			if isDecimalPoint(runes, i) { // 3.14 reste un seul mot
				buf = append(buf, r)
				continue
			}
			flush()
			j := i
			for j < len(runes) && isPunct(runes[j]) { // "...", "!?" restent groupés
				j++
			}
			tokens = append(tokens, token{punct, string(runes[i:j])})
			i = j - 1
		case r == '\'' && isQuote(runes, i):
			flush()
			tokens = append(tokens, token{quote, "'"})
		default:
			buf = append(buf, r)
		}
	}
	flush()
	return tokens
}

// isDecimalPoint : vrai pour le '.' de "3.14" (chiffre avant et après).
func isDecimalPoint(runes []rune, i int) bool {
	return runes[i] == '.' && i > 0 && i+1 < len(runes) &&
		unicode.IsDigit(runes[i-1]) && unicode.IsDigit(runes[i+1])
}

// isQuote : une apostrophe au milieu d'un mot (don't) n'est pas un guillemet.
func isQuote(runes []rune, i int) bool {
	prevIsWord := i > 0 && isAlnum(runes[i-1])
	nextIsWord := i+1 < len(runes) && isAlnum(runes[i+1])
	return !(prevIsWord && nextIsWord)
}

// ---------------------------------------------------------------------------
// 2. Commandes (hex) (bin) (up) (low) (cap)
// ---------------------------------------------------------------------------

// applyCommands construit une nouvelle liste : les mots normaux sont copiés,
// et chaque commande modifie les mots déjà présents dans `out`.
func applyCommands(tokens []token) []token {
	var out []token
	for _, t := range tokens {
		if t.kind != command {
			out = append(out, t)
			continue
		}

		m := tokRe.FindStringSubmatch(t.text)
		name := m[1]
		n := 1
		if m[2] != "" {
			n, _ = strconv.Atoi(m[2])
		}

		switch name {
		case "hex":
			convertBase(out, 16)
		case "bin":
			convertBase(out, 2)
		default:
			changeCase(out, name, n)
		}
	}
	return out
}

// lastWordIndexes renvoie les positions des n derniers mots (la ponctuation
// est ignorée : "Ready, set, go (up) !" cible bien "go").
func lastWordIndexes(tokens []token, n int) []int {
	var idx []int
	for i := len(tokens) - 1; i >= 0 && len(idx) < n; i-- {
		if tokens[i].kind == word {
			idx = append(idx, i)
		}
	}
	return idx
}

// convertBase remplace le mot précédent par sa valeur décimale.
// Si la conversion échoue, le mot reste inchangé.
func convertBase(tokens []token, base int) {
	idx := lastWordIndexes(tokens, 1)
	if len(idx) == 0 {
		return
	}
	i := idx[0]
	if v, err := strconv.ParseInt(tokens[i].text, base, 64); err == nil {
		tokens[i].text = strconv.FormatInt(v, 10)
	}
}

// changeCase applique up / low / cap aux n derniers mots.
func changeCase(tokens []token, name string, n int) {
	for _, i := range lastWordIndexes(tokens, n) {
		switch name {
		case "up":
			tokens[i].text = strings.ToUpper(tokens[i].text)
		case "low":
			tokens[i].text = strings.ToLower(tokens[i].text)
		case "cap":
			tokens[i].text = capitalize(tokens[i].text)
		}
	}
}

// capitalize : première lettre en majuscule, le reste en minuscules.
func capitalize(w string) string {
	r := []rune(strings.ToLower(w))
	if len(r) > 0 {
		r[0] = unicode.ToUpper(r[0])
	}
	return string(r)
}

// ---------------------------------------------------------------------------
// 3. a -> an
// ---------------------------------------------------------------------------

// fixArticles transforme "a" en "an" si le mot suivant commence par
// une voyelle (a e i o u) ou un h.
func fixArticles(tokens []token) {
	for i := 0; i < len(tokens)-1; i++ {
		cur, next := tokens[i], tokens[i+1]
		if cur.kind != word || (cur.text != "a" && cur.text != "A") || next.kind != word {
			continue
		}
		first := unicode.ToLower([]rune(next.text)[0])
		if strings.ContainsRune("aeiouh", first) {
			tokens[i].text += "n"
		}
	}
}

// ---------------------------------------------------------------------------
// 4. Reconstruction du texte (espaces, ponctuation, apostrophes)
// ---------------------------------------------------------------------------

// render recolle les tokens en respectant les règles d'espacement :
//   - la ponctuation se colle au mot précédent ;
//   - une apostrophe ouvrante se colle au mot suivant ;
//   - une apostrophe fermante se colle au mot précédent ;
//   - sinon, un seul espace entre deux éléments.
func render(tokens []token) string {
	var sb strings.Builder
	glue := false // vrai : pas d'espace avant le prochain mot
	open := false // vrai : on est entre deux apostrophes

	for _, t := range tokens {
		switch t.kind {
		case punct:
			sb.WriteString(t.text)
		case quote:
			if !open { // apostrophe ouvrante
				if sb.Len() > 0 {
					sb.WriteByte(' ')
				}
				sb.WriteByte('\'')
				glue = true
			} else { // apostrophe fermante
				sb.WriteByte('\'')
			}
			open = !open
		default: // word
			if sb.Len() > 0 && !glue {
				sb.WriteByte(' ')
			}
			sb.WriteString(t.text)
			glue = false
		}
	}
	return sb.String()
}
