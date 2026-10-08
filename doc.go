// Package proserhythm measures how sentence rhythm varies in prose and says
// where it goes flat.
//
// Count tallies plain-text paragraphs, already free of markup, into a Tally of
// integers. Tallies of one language add up, so a block, a page and a whole
// collection are counted alike and judged once by Judge. Places names the
// paragraphs holding runs of sentences that open with their subject.
//
// The checks, their limits and each language's word lists are data in
// rhythm.yaml, shared with the leading Python port together with the test
// cases in corpus/.
package proserhythm
