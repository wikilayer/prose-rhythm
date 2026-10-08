package proserhythm

import (
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const leadingPort = "https://raw.githubusercontent.com/wikilayer/prose-rhythm-python/main/"

var copied = map[string]string{
	"rhythm.yaml":           "src/prose_rhythm/rhythm.yaml",
	"corpus/sentences.yaml": "corpus/sentences.yaml",
	"corpus/openings.yaml":  "corpus/openings.yaml",
	"corpus/dialogue.yaml":  "corpus/dialogue.yaml",
	"corpus/tallies.yaml":   "corpus/tallies.yaml",
	"corpus/novels.yaml":    "corpus/novels.yaml",
}

func TestEveryCopyIsTheFileTheLeadingPortHolds(t *testing.T) {
	client := http.Client{Timeout: 20 * time.Second}
	for copy, original := range copied {
		t.Run(copy, func(t *testing.T) {
			held, err := os.ReadFile(copy)
			require.NoError(t, err, "%s is missing; run make sync-corpus in prose-rhythm-python", copy)

			answer, err := client.Get(leadingPort + original)
			require.NoError(t, err, "the leading port cannot be reached, so this test proves nothing")
			defer answer.Body.Close()
			require.Equal(t, http.StatusOK, answer.StatusCode,
				"the leading port answered %d for %s, so this test proves nothing", answer.StatusCode, original)
			leading, err := io.ReadAll(answer.Body)
			require.NoError(t, err)

			assert.Equal(t, string(leading), string(held),
				"%s is not the file the leading port holds, so this port answers older rules; run make sync-corpus in prose-rhythm-python", copy)
		})
	}
}
