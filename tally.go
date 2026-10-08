package proserhythm

import (
	"fmt"
	"math"
)

// English is the code of the language the rules hold first.
const English = "en"

// Tally counts the narration of one or more paragraphs. Tallies of the same
// language add up, so any part of a text and the whole are counted alike.
type Tally struct {
	Language            string
	Sentences           int
	Words               int
	SquaredWords        int
	LongSentences       int
	Longest             int
	Paragraphs          int
	UniformParagraphs   int
	SubjectRunSentences int
	DialogueParagraphs  int
}

// Finding is a check the tally fails: its measure, which side of the limit
// is wanted, and the limit.
type Finding struct {
	Kind  string
	Value float64
	Bound string
	Limit float64
}

// Verdict says whether a tally holds enough narration to judge and, if so,
// which checks it fails.
type Verdict struct {
	Judged   bool
	Findings []Finding
}

// Run is a run of sentences opening with their subject, in the paragraph of
// that index among those given.
type Run struct {
	Paragraph int
	Sentences []Sentence
}

// Add returns the sum of two tallies of the same language.
func (t Tally) Add(other Tally) (Tally, error) {
	if other.Language != t.Language {
		return Tally{}, fmt.Errorf("%w: %s and %s", ErrMixedLanguages, t.Language, other.Language)
	}
	return Tally{
		Language:            t.Language,
		Sentences:           t.Sentences + other.Sentences,
		Words:               t.Words + other.Words,
		SquaredWords:        t.SquaredWords + other.SquaredWords,
		LongSentences:       t.LongSentences + other.LongSentences,
		Longest:             max(t.Longest, other.Longest),
		Paragraphs:          t.Paragraphs + other.Paragraphs,
		UniformParagraphs:   t.UniformParagraphs + other.UniformParagraphs,
		SubjectRunSentences: t.SubjectRunSentences + other.SubjectRunSentences,
		DialogueParagraphs:  t.DialogueParagraphs + other.DialogueParagraphs,
	}, nil
}

// Mean is the mean sentence length in words.
func (t Tally) Mean() float64 {
	return share(t.Words, t.Sentences)
}

// Spread is the standard deviation of sentence length over its mean,
// computed from integer sums as sqrt(n·Σx² − (Σx)²) / Σx.
func (t Tally) Spread() float64 {
	if t.Words == 0 {
		return 0
	}
	return math.Sqrt(float64(t.Sentences*t.SquaredWords-t.Words*t.Words)) / float64(t.Words)
}

// LongShare is the share of sentences of the long length or more.
func (t Tally) LongShare() float64 {
	return share(t.LongSentences, t.Sentences)
}

// UniformShare is the share of paragraphs within the uniform range of sentences.
func (t Tally) UniformShare() float64 {
	return share(t.UniformParagraphs, t.Paragraphs)
}

// SubjectRunShare is the share of sentences inside runs opening with their subject.
func (t Tally) SubjectRunShare() float64 {
	return share(t.SubjectRunSentences, t.Sentences)
}

func (t Tally) metric(kind string) float64 {
	switch kind {
	case Spread:
		return t.Spread()
	case LongSentences:
		return t.LongShare()
	case UniformParagraphs:
		return t.UniformShare()
	default:
		return t.SubjectRunShare()
	}
}

func share(part, whole int) float64 {
	if whole == 0 {
		return 0
	}
	return float64(part) / float64(whole)
}

// Count tallies plain-text paragraphs, free of markup, in the given language.
func Count(paragraphs []string, code string) (Tally, error) {
	rules, err := rulesFor(code)
	if err != nil {
		return Tally{}, err
	}
	total := Tally{Language: rules.code}
	for _, paragraph := range paragraphs {
		total, _ = total.Add(rules.paragraphTally(paragraph))
	}
	return total, nil
}

// Judge leaves a tally of fewer sentences than its language needs unjudged
// and otherwise returns the checks it fails, in the order the rules hold them.
func Judge(t Tally) (Verdict, error) {
	rules, err := rulesFor(t.Language)
	if err != nil {
		return Verdict{}, err
	}
	if t.Sentences < rules.Thresholds.MinimumSentences {
		return Verdict{Judged: false, Findings: []Finding{}}, nil
	}
	findings := []Finding{}
	for _, c := range rules.Checks {
		if value := t.metric(c.Kind); c.fails(value) {
			findings = append(findings, Finding{Kind: c.Kind, Value: value, Bound: c.Bound, Limit: c.Limit})
		}
	}
	return Verdict{Judged: true, Findings: findings}, nil
}

// Places returns every run of sentences opening with their subject, by the
// index of its paragraph among those given; dialogue is left out.
func Places(paragraphs []string, code string) ([]Run, error) {
	rules, err := rulesFor(code)
	if err != nil {
		return nil, err
	}
	found := []Run{}
	for index, paragraph := range paragraphs {
		if rules.isDialogue(paragraph) {
			continue
		}
		for _, run := range rules.subjectRuns(rules.sentencesOf(paragraph)) {
			found = append(found, Run{Paragraph: index, Sentences: run})
		}
	}
	return found, nil
}

func (l *language) paragraphTally(paragraph string) Tally {
	if l.isDialogue(paragraph) {
		return Tally{Language: l.code, DialogueParagraphs: 1}
	}
	sentences := l.sentencesOf(paragraph)
	if len(sentences) == 0 {
		return Tally{Language: l.code}
	}
	counted := Tally{Language: l.code, Sentences: len(sentences), Paragraphs: 1}
	for _, sentence := range sentences {
		counted.Words += sentence.Words
		counted.SquaredWords += sentence.Words * sentence.Words
		counted.Longest = max(counted.Longest, sentence.Words)
		if sentence.Words >= l.Thresholds.LongWords {
			counted.LongSentences++
		}
	}
	low, high := l.Thresholds.UniformParagraphSentences[0], l.Thresholds.UniformParagraphSentences[1]
	if low <= len(sentences) && len(sentences) <= high {
		counted.UniformParagraphs = 1
	}
	for _, run := range l.subjectRuns(sentences) {
		counted.SubjectRunSentences += len(run)
	}
	return counted
}

func (l *language) subjectRuns(sentences []Sentence) [][]Sentence {
	var found [][]Sentence
	var run []Sentence
	for at := 0; at <= len(sentences); at++ {
		if at < len(sentences) && l.opensWithSubject(sentences[at].Text) {
			run = append(run, sentences[at])
			continue
		}
		if len(run) >= l.Thresholds.SubjectRun {
			found = append(found, run)
		}
		run = nil
	}
	return found
}
