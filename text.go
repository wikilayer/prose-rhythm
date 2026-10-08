package proserhythm

import (
	"slices"
	"strings"
	"unicode"
)

const spaces = "\t\n\v\f\r \u0085"

// Sentence is one sentence of narration and the number of words in it.
type Sentence struct {
	Text  string
	Words int
}

func isSpace(r rune) bool {
	return strings.ContainsRune(spaces, r) || unicode.In(r, unicode.Zs, unicode.Zl, unicode.Zp)
}

func isWord(r rune) bool {
	return unicode.In(r, unicode.L, unicode.N)
}

func isMark(r rune) bool {
	return unicode.In(r, unicode.M)
}

func isUpper(r rune) bool {
	return unicode.Is(unicode.Lu, r)
}

func isDecimal(r rune) bool {
	return unicode.Is(unicode.Nd, r)
}

func lowered(word string) string {
	return strings.Map(unicode.ToLower, word)
}

func stripped(text string) string {
	return strings.TrimFunc(text, isSpace)
}

func (l *language) words(text string) []string {
	var found []string
	var word []rune
	for _, r := range text + " " {
		continues := strings.ContainsRune(l.WordJoiners, r) || isMark(r)
		if isWord(r) || (len(word) > 0 && continues) {
			word = append(word, r)
		} else if len(word) > 0 {
			found = append(found, strings.TrimRight(string(word), l.WordJoiners))
			word = nil
		}
	}
	return found
}

func (l *language) sentenceBreaks(paragraph []rune) []int {
	var found []int
	end := len(paragraph)
	position := 0
	for position < end {
		if !strings.ContainsRune(l.SentenceEnds, paragraph[position]) {
			position++
			continue
		}
		for position < end && strings.ContainsRune(l.SentenceEnds, paragraph[position]) {
			position++
		}
		for position < end && strings.ContainsRune(l.Closers, paragraph[position]) {
			position++
		}
		space := position
		for space < end && isSpace(paragraph[space]) {
			space++
		}
		if space > position {
			found = append(found, space)
			position = space
		}
	}
	return found
}

func (l *language) opensWithSubject(sentence string) bool {
	words := l.words(sentence)
	if len(words) == 0 {
		return false
	}
	first := lowered(words[0])
	if slices.Contains(l.NonSubjectOpeners, first) {
		return false
	}
	if slices.Contains(l.SubjectDespiteSuffix, first) {
		return true
	}
	for _, suffix := range l.NonSubjectSuffixes {
		if strings.HasSuffix(first, suffix) {
			return false
		}
	}
	return true
}

func (l *language) stopsShort(sentence []rune) bool {
	last := string(sentence[lastWordStart(sentence):])
	last = strings.TrimLeft(strings.TrimRight(last, l.Closers), l.Openers)
	if slices.Contains(l.Abbreviations, last) {
		return true
	}
	letters := []rune(last)
	initial := len(letters) == 2 && isUpper(letters[0]) && letters[1] == '.'
	return initial && !slices.Contains(l.InitialExceptions, last)
}

func (l *language) beginsSentence(text []rune) bool {
	if len(text) == 0 {
		return true
	}
	first := text[0]
	return isUpper(first) || isDecimal(first) || strings.ContainsRune(l.Openers, first)
}

func (l *language) isDialogue(paragraph string) bool {
	text := []rune(paragraph)
	if len(text) == 0 {
		return false
	}
	quoted := 0
	for _, pair := range l.QuotePairs {
		quoted += quotedLength(text, []rune(pair[0]), []rune(pair[1]))
	}
	return float64(quoted)/float64(len(text)) >= l.DialogueQuotedShare
}

func (l *language) sentencesOf(paragraph string) []Sentence {
	text := []rune(paragraph)
	var texts []string
	start := 0
	for _, end := range l.sentenceBreaks(text) {
		candidate := []rune(stripped(string(text[start:end])))
		if l.stopsShort(candidate) || !l.beginsSentence(text[end:]) {
			continue
		}
		texts = append(texts, string(candidate))
		start = end
	}
	texts = append(texts, stripped(string(text[start:])))
	var sentences []Sentence
	for _, each := range texts {
		if words := len(l.words(each)); words > 0 {
			sentences = append(sentences, Sentence{Text: each, Words: words})
		}
	}
	return sentences
}

func lastWordStart(text []rune) int {
	position := len(text)
	for position > 0 && !isSpace(text[position-1]) {
		position--
	}
	return position
}

func quotedLength(text, opening, closing []rune) int {
	total := 0
	position := 0
	for {
		start := index(text, opening, position)
		for start > 0 && isWord(text[start-1]) {
			start = index(text, opening, start+1)
		}
		if start == -1 {
			return total
		}
		end := index(text, closing, start+len(opening))
		for end != -1 && wordFollows(text, end+len(closing)) {
			end = index(text, closing, end+1)
		}
		if end == -1 {
			return total
		}
		position = end + len(closing)
		total += position - start
	}
}

func index(text, part []rune, from int) int {
	for at := from; at+len(part) <= len(text); at++ {
		if slices.Equal(text[at:at+len(part)], part) {
			return at
		}
	}
	return -1
}

func wordFollows(text []rune, position int) bool {
	return position < len(text) && isWord(text[position])
}
