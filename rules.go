package proserhythm

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed rhythm.yaml
var writtenRules []byte

// Bounds say on which side of its limit a check is satisfied.
const (
	Min = "min"
	Max = "max"
)

// Kinds of check a language may hold.
const (
	Spread            = "spread"
	LongSentences     = "long_sentences"
	UniformParagraphs = "uniform_paragraphs"
	SubjectRuns       = "subject_runs"
)

var kinds = []string{Spread, LongSentences, UniformParagraphs, SubjectRuns}

// ErrUnreadable wraps a failure to read the rules shipped with the package.
var ErrUnreadable = errors.New("rhythm.yaml is unreadable")

// ErrUnknownLanguage is returned for a language the rules do not hold.
var ErrUnknownLanguage = errors.New("no rules for this language")

// ErrMixedLanguages is returned when tallies of different languages are added.
var ErrMixedLanguages = errors.New("tallies of different languages do not add up")

type check struct {
	Kind  string  `yaml:"kind"`
	Bound string  `yaml:"bound"`
	Limit float64 `yaml:"limit"`
}

func (c check) fails(value float64) bool {
	if c.Bound == Min {
		return value < c.Limit
	}
	return value > c.Limit
}

type thresholds struct {
	MinimumSentences          int    `yaml:"minimum_sentences"`
	LongWords                 int    `yaml:"long_words"`
	UniformParagraphSentences [2]int `yaml:"uniform_paragraph_sentences"`
	SubjectRun                int    `yaml:"subject_run"`
}

type language struct {
	code                 string
	Thresholds           thresholds  `yaml:"thresholds"`
	Checks               []check     `yaml:"checks"`
	WordJoiners          string      `yaml:"word_joiners"`
	SentenceEnds         string      `yaml:"sentence_ends"`
	Closers              string      `yaml:"closers"`
	Openers              string      `yaml:"openers"`
	QuotePairs           [][2]string `yaml:"quote_pairs"`
	DialogueQuotedShare  float64     `yaml:"dialogue_quoted_share"`
	InitialExceptions    []string    `yaml:"initial_exceptions"`
	Abbreviations        []string    `yaml:"abbreviations"`
	NonSubjectSuffixes   []string    `yaml:"non_subject_suffixes"`
	SubjectDespiteSuffix []string    `yaml:"subject_despite_suffix"`
	NonSubjectOpeners    []string    `yaml:"non_subject_openers"`
}

var (
	loaded     map[string]*language
	loadFailed error
	loadOnce   sync.Once
)

func languages() (map[string]*language, error) {
	loadOnce.Do(func() {
		loaded, loadFailed = readRules(writtenRules)
	})
	return loaded, loadFailed
}

func readRules(source []byte) (map[string]*language, error) {
	var written struct {
		Languages map[string]*language `yaml:"languages"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(source))
	decoder.KnownFields(true)
	if err := decoder.Decode(&written); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnreadable, err)
	}
	if len(written.Languages) == 0 {
		return nil, fmt.Errorf("%w: no languages", ErrUnreadable)
	}
	for code, rules := range written.Languages {
		rules.code = code
		if err := rules.complete(); err != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrUnreadable, code, err)
		}
	}
	return written.Languages, nil
}

func (l *language) complete() error {
	limits := l.Thresholds
	if limits.MinimumSentences <= 0 || limits.LongWords <= 0 || limits.SubjectRun <= 0 {
		return errors.New("thresholds must all be positive")
	}
	if limits.UniformParagraphSentences[0] <= 0 ||
		limits.UniformParagraphSentences[1] < limits.UniformParagraphSentences[0] {
		return errors.New("uniform_paragraph_sentences must be a range of positive counts")
	}
	if len(l.Checks) == 0 {
		return errors.New("no checks")
	}
	for _, c := range l.Checks {
		if !slices.Contains(kinds, c.Kind) {
			return fmt.Errorf("no check is called %q", c.Kind)
		}
		if c.Bound != Min && c.Bound != Max {
			return fmt.Errorf("%s is bound by neither %s nor %s", c.Kind, Min, Max)
		}
	}
	required := map[string]int{
		"word_joiners":           len(l.WordJoiners),
		"sentence_ends":          len(l.SentenceEnds),
		"closers":                len(l.Closers),
		"openers":                len(l.Openers),
		"quote_pairs":            len(l.QuotePairs),
		"initial_exceptions":     len(l.InitialExceptions),
		"abbreviations":          len(l.Abbreviations),
		"non_subject_suffixes":   len(l.NonSubjectSuffixes),
		"subject_despite_suffix": len(l.SubjectDespiteSuffix),
		"non_subject_openers":    len(l.NonSubjectOpeners),
	}
	for key, size := range required {
		if size == 0 {
			return fmt.Errorf("%s is missing", key)
		}
	}
	if l.DialogueQuotedShare <= 0 {
		return errors.New("dialogue_quoted_share must be positive")
	}
	return nil
}

func rulesFor(code string) (*language, error) {
	known, err := languages()
	if err != nil {
		return nil, err
	}
	rules, ok := known[code]
	if !ok {
		codes := make([]string, 0, len(known))
		for each := range known {
			codes = append(codes, each)
		}
		sort.Strings(codes)
		return nil, fmt.Errorf("%w: %q, only for %s", ErrUnknownLanguage, code, strings.Join(codes, ", "))
	}
	return rules, nil
}

// Languages lists the BCP 47 codes the rules hold, sorted.
func Languages() ([]string, error) {
	known, err := languages()
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(known))
	for code := range known {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes, nil
}
