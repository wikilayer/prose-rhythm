# Changelog

Notable changes to `prose-rhythm` for Go are documented here in the format of
[Keep a Changelog](https://keepachangelog.com/). The rules and cases are copies
of those in the leading Python port; matching major and minor versions promise
the same counts.

The package remains below 1.0 while its public API is settling.

## 0.3.0 - 2026-10-08

The first release, answering every case of the leading port's 0.3.0.

### Added

- `Count` tallies plain-text paragraphs into a `Tally` of integers, and
  `Tally.Add` sums tallies of one language.
- `Judge` returns a `Verdict`: unjudged below the language's minimum of
  sentences, otherwise the failed checks as `Finding`s in the order the rules
  hold them.
- `Places` names the runs of sentences opening with their subject by paragraph
  index.
- `Languages` lists the codes the rules hold; other codes return
  `ErrUnknownLanguage`.
- The rules are read strictly: an unknown key, check or bound makes them
  unreadable rather than silently ignored.
