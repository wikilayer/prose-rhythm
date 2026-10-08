package proserhythm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	proserhythm "github.com/wikilayer/prose-rhythm"
)

func TestTheReadmeExampleRuns(t *testing.T) {
	paragraphs := []string{
		"He climbed the stairs to the attic. The boxes stood where his father had left them.",
		"Rain. Under the eave the stone stayed dry.",
	}
	total := proserhythm.Tally{Language: proserhythm.English}
	for _, paragraph := range paragraphs {
		counted, err := proserhythm.Count([]string{paragraph}, proserhythm.English)
		require.NoError(t, err)
		total, err = total.Add(counted)
		require.NoError(t, err)
	}
	verdict, err := proserhythm.Judge(total)
	require.NoError(t, err)

	whole, err := proserhythm.Count(paragraphs, proserhythm.English)
	require.NoError(t, err)
	assert.Equal(t, whole, total)
	assert.Equal(t, 4, total.Sentences)
	assert.False(t, verdict.Judged)
}

func TestLanguagesNameEnglish(t *testing.T) {
	codes, err := proserhythm.Languages()
	require.NoError(t, err)
	assert.Contains(t, codes, proserhythm.English)
}
