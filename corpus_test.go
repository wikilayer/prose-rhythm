package proserhythm

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type writtenTally struct {
	Sentences           int `yaml:"sentences"`
	Words               int `yaml:"words"`
	SquaredWords        int `yaml:"squared_words"`
	LongSentences       int `yaml:"long_sentences"`
	Longest             int `yaml:"longest"`
	Paragraphs          int `yaml:"paragraphs"`
	UniformParagraphs   int `yaml:"uniform_paragraphs"`
	SubjectRunSentences int `yaml:"subject_run_sentences"`
	DialogueParagraphs  int `yaml:"dialogue_paragraphs"`
}

func (w writtenTally) in(code string) Tally {
	return Tally{
		Language:            code,
		Sentences:           w.Sentences,
		Words:               w.Words,
		SquaredWords:        w.SquaredWords,
		LongSentences:       w.LongSentences,
		Longest:             w.Longest,
		Paragraphs:          w.Paragraphs,
		UniformParagraphs:   w.UniformParagraphs,
		SubjectRunSentences: w.SubjectRunSentences,
		DialogueParagraphs:  w.DialogueParagraphs,
	}
}

func read(t *testing.T, name string, into any) {
	t.Helper()
	written, err := os.ReadFile("corpus/" + name)
	require.NoError(t, err, "the cases every port answers to are unreadable; run make sync-corpus in prose-rhythm-python")
	decoder := yaml.NewDecoder(bytes.NewReader(written))
	decoder.KnownFields(true)
	require.NoError(t, decoder.Decode(into), "%s does not parse, or holds something this port does not read", name)
}

func english(t *testing.T, code string) *language {
	t.Helper()
	rules, err := rulesFor(code)
	require.NoError(t, err)
	return rules
}

func TestSentencesBreakAsTheCorpusSays(t *testing.T) {
	var held struct {
		Language string `yaml:"language"`
		Cases    []struct {
			Name      string  `yaml:"name"`
			Paragraph string  `yaml:"paragraph"`
			Sentences [][]any `yaml:"sentences"`
		} `yaml:"cases"`
	}
	read(t, "sentences.yaml", &held)
	require.NotEmpty(t, held.Cases, "no cases, so this test cannot fail")
	rules := english(t, held.Language)
	for _, one := range held.Cases {
		t.Run(one.Name, func(t *testing.T) {
			want := []Sentence{}
			for _, pair := range one.Sentences {
				require.Len(t, pair, 2)
				want = append(want, Sentence{Text: pair[0].(string), Words: pair[1].(int)})
			}
			found := rules.sentencesOf(one.Paragraph)
			if found == nil {
				found = []Sentence{}
			}
			assert.Equal(t, want, found)
		})
	}
}

func TestOpeningsAsTheCorpusSays(t *testing.T) {
	var held struct {
		Language string `yaml:"language"`
		Cases    []struct {
			Sentence     string `yaml:"sentence"`
			SubjectFirst bool   `yaml:"subject_first"`
		} `yaml:"cases"`
	}
	read(t, "openings.yaml", &held)
	require.NotEmpty(t, held.Cases, "no cases, so this test cannot fail")
	rules := english(t, held.Language)
	for _, one := range held.Cases {
		t.Run(one.Sentence, func(t *testing.T) {
			assert.Equal(t, one.SubjectFirst, rules.opensWithSubject(one.Sentence))
		})
	}
}

func TestDialogueAsTheCorpusSays(t *testing.T) {
	var held struct {
		Language string `yaml:"language"`
		Cases    []struct {
			Name      string `yaml:"name"`
			Paragraph string `yaml:"paragraph"`
			Dialogue  bool   `yaml:"dialogue"`
		} `yaml:"cases"`
	}
	read(t, "dialogue.yaml", &held)
	require.NotEmpty(t, held.Cases, "no cases, so this test cannot fail")
	rules := english(t, held.Language)
	for _, one := range held.Cases {
		t.Run(one.Name, func(t *testing.T) {
			assert.Equal(t, one.Dialogue, rules.isDialogue(one.Paragraph))
		})
	}
}

func TestTalliesAsTheCorpusSays(t *testing.T) {
	var held struct {
		Language string            `yaml:"language"`
		Anchors  map[string]string `yaml:"anchors"`
		Cases    []struct {
			Name       string       `yaml:"name"`
			Paragraphs []string     `yaml:"paragraphs"`
			Tally      writtenTally `yaml:"tally"`
			Judged     bool         `yaml:"judged"`
			Findings   []struct {
				Kind  string  `yaml:"kind"`
				Value float64 `yaml:"value"`
				Bound string  `yaml:"bound"`
				Limit float64 `yaml:"limit"`
			} `yaml:"findings"`
			Places []struct {
				Paragraph int `yaml:"paragraph"`
				Sentences int `yaml:"sentences"`
			} `yaml:"places"`
		} `yaml:"cases"`
	}
	read(t, "tallies.yaml", &held)
	require.NotEmpty(t, held.Cases, "no cases, so this test cannot fail")
	for _, one := range held.Cases {
		t.Run(one.Name, func(t *testing.T) {
			counted, err := Count(one.Paragraphs, held.Language)
			require.NoError(t, err)
			assert.Equal(t, one.Tally.in(held.Language), counted)

			summed := Tally{Language: held.Language}
			for _, paragraph := range one.Paragraphs {
				single, err := Count([]string{paragraph}, held.Language)
				require.NoError(t, err)
				summed, err = summed.Add(single)
				require.NoError(t, err)
			}
			assert.Equal(t, counted, summed, "the paragraphs counted one by one do not add up to the whole")

			verdict, err := Judge(counted)
			require.NoError(t, err)
			assert.Equal(t, one.Judged, verdict.Judged)
			want := []Finding{}
			for _, finding := range one.Findings {
				want = append(want, Finding(finding))
			}
			assert.Equal(t, want, verdict.Findings)

			runs, err := Places(one.Paragraphs, held.Language)
			require.NoError(t, err)
			found := [][2]int{}
			for _, run := range runs {
				found = append(found, [2]int{run.Paragraph, len(run.Sentences)})
			}
			wanted := [][2]int{}
			for _, place := range one.Places {
				wanted = append(wanted, [2]int{place.Paragraph, place.Sentences})
			}
			assert.Equal(t, wanted, found)
		})
	}
}

func TestTalliesOfDifferentLanguagesDoNotAdd(t *testing.T) {
	_, err := Tally{Language: "en"}.Add(Tally{Language: "de"})
	assert.ErrorIs(t, err, ErrMixedLanguages)
}

func TestAnUnknownLanguageIsRefused(t *testing.T) {
	_, err := Count([]string{"He ran."}, "xx")
	assert.ErrorIs(t, err, ErrUnknownLanguage)
}
