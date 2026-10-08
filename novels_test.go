//go:build novels

package proserhythm

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEveryNovelCountsAsRecordedAndPasses(t *testing.T) {
	var held struct {
		Language string `yaml:"language"`
		Novels   []struct {
			ID    int          `yaml:"id"`
			Title string       `yaml:"title"`
			Tally writtenTally `yaml:"tally"`
		} `yaml:"novels"`
	}
	read(t, "novels.yaml", &held)
	require.NotEmpty(t, held.Novels, "no novels, so this test cannot fail")
	for _, novel := range held.Novels {
		t.Run(novel.Title, func(t *testing.T) {
			path := fmt.Sprintf("data/%d.txt", novel.ID)
			text, err := os.ReadFile(path)
			require.NoError(t, err, "%s is missing; run make data", path)

			counted, err := Count(paragraphs(string(text)), held.Language)
			require.NoError(t, err)
			assert.Equal(t, novel.Tally.in(held.Language), counted)

			verdict, err := Judge(counted)
			require.NoError(t, err)
			assert.True(t, verdict.Judged)
			assert.Empty(t, verdict.Findings)
		})
	}
}

func paragraphs(text string) []string {
	var found []string
	var held []string
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if line = stripped(line); line != "" {
			held = append(held, line)
		} else if len(held) > 0 {
			found = append(found, strings.Join(held, " "))
			held = nil
		}
	}
	if len(held) > 0 {
		found = append(found, strings.Join(held, " "))
	}
	return found
}
