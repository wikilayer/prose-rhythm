# prose-rhythm

[![Tests](https://github.com/wikilayer/prose-rhythm/actions/workflows/tests.yml/badge.svg)](https://github.com/wikilayer/prose-rhythm/actions/workflows/tests.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/wikilayer/prose-rhythm.svg)](https://pkg.go.dev/github.com/wikilayer/prose-rhythm)

Measures how sentence rhythm varies in prose and says where it goes flat: how
much sentence length varies, how many sentences run long, how many paragraphs
are of one size, and how many sentences sit in runs that open with their
subject.

This is the Go port. The rules and test cases are copies of those held by the
leading [Python port](https://github.com/wikilayer/prose-rhythm-python), whose
README explains the measures; matching major and minor versions promise the
same counts.

```sh
go get github.com/wikilayer/prose-rhythm@v0.3.0
```

## Using

The package takes plain text: paragraphs already free of markup. `Count` tallies
them into integers, tallies of one language add up, and `Judge` judges the sum
once, so a block, a page and a whole collection are counted alike.

```go
paragraphs := []string{
	"He climbed the stairs to the attic. The boxes stood where his father had left them.",
	"Rain. Under the eave the stone stayed dry.",
}
total := proserhythm.Tally{Language: proserhythm.English}
for _, paragraph := range paragraphs {
	counted, err := proserhythm.Count([]string{paragraph}, proserhythm.English)
	if err != nil {
		return err
	}
	if total, err = total.Add(counted); err != nil {
		return err
	}
}
verdict, err := proserhythm.Judge(total)
```

`Judge` leaves fewer than 30 sentences of narration unjudged. `Places` names the
paragraphs holding runs of sentences that open with their subject, by index.

## Shared rules and cases

[`rhythm.yaml`](https://github.com/wikilayer/prose-rhythm/blob/main/rhythm.yaml)
and [`corpus/`](https://github.com/wikilayer/prose-rhythm/tree/main/corpus) are
copies; `make sync-corpus` in the leading port writes them, and a test compares
each with the leading port's `main`. Changes to the rules or cases go there
first.

`make measure` prepares the eight reference novels with the leading port's
`tools/prose.py`, from a checkout next to this one, and checks that this port
counts them exactly as recorded.

## Lines of Code

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/wikilayer/prose-rhythm/main/.github/loc-history-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/wikilayer/prose-rhythm/main/.github/loc-history-light.svg">
  <img alt="Lines of Code graph" src="https://raw.githubusercontent.com/wikilayer/prose-rhythm/main/.github/loc-history-light.svg">
</picture>
