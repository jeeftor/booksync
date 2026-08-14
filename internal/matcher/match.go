// Package matcher fuzzy-matches Kindle library entries against Audiobookshelf
// library items by title/author, since filenames and metadata between the
// two rarely line up exactly. It only proposes candidates; a human confirms
// or rejects them via the API before a mapping becomes active.
package matcher

import (
	"regexp"
	"sort"
	"strings"

	"github.com/agnivade/levenshtein"

	"github.com/jeeftor/bookSync/internal/absclient"
	"github.com/jeeftor/bookSync/internal/kindleclient"
)

// Candidate is a proposed link between one Kindle book and one Audiobookshelf
// library item, ranked by Confidence in [0, 1].
type Candidate struct {
	KindleASIN  string
	KindleTitle string
	ABSItemID   string
	ABSTitle    string
	Confidence  float64
}

// MinConfidence is the default threshold below which candidates are not
// surfaced as suggestions at all (still findable via manual matching).
const MinConfidence = 0.55

// Suggest proposes the single best Audiobookshelf item for each Kindle book,
// sorted with the strongest matches first. Callers should filter/display by
// Confidence and let a human confirm before persisting a mapping.
//
// isRejected, if non-nil, is consulted per (ASIN, ABS item ID) pair; matches
// it reports as previously rejected are skipped entirely, so the next-best
// candidate for that Kindle book can surface instead of resurfacing the same
// rejected pairing.
func Suggest(kindleBooks []kindleclient.Book, absItems []absclient.LibraryItem, isRejected func(asin, absItemID string) bool) []Candidate {
	var out []Candidate

	for _, kb := range kindleBooks {
		var best *Candidate
		for _, ai := range absItems {
			if isRejected != nil && isRejected(kb.ASIN, ai.ID) {
				continue
			}
			score := score(kb.Title, kb.Authors, ai.Title, ai.Authors)
			if best == nil || score > best.Confidence {
				best = &Candidate{
					KindleASIN:  kb.ASIN,
					KindleTitle: kb.Title,
					ABSItemID:   ai.ID,
					ABSTitle:    ai.Title,
					Confidence:  score,
				}
			}
		}
		if best != nil && best.Confidence >= MinConfidence {
			out = append(out, *best)
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Confidence > out[j].Confidence })
	return out
}

// score combines title similarity (weighted most heavily) with author
// similarity into a single 0..1 confidence value.
func score(titleA string, authorsA []string, titleB string, authorsB []string) float64 {
	titleScore := similarity(normalize(titleA), normalize(titleB))
	authorScore := bestAuthorSimilarity(authorsA, authorsB)
	return 0.7*titleScore + 0.3*authorScore
}

func bestAuthorSimilarity(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	best := 0.0
	for _, x := range a {
		for _, y := range b {
			if s := similarity(normalize(x), normalize(y)); s > best {
				best = s
			}
		}
	}
	return best
}

// similarity is 1 - normalized Levenshtein distance, in [0, 1].
func similarity(a, b string) float64 {
	if a == "" && b == "" {
		return 1
	}
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 1
	}
	dist := levenshtein.ComputeDistance(a, b)
	return 1 - float64(dist)/float64(maxLen)
}

var (
	subtitleRe   = regexp.MustCompile(`[:(].*$`) // drop subtitles/parentheticals like "Book 1)"
	nonAlnumRe   = regexp.MustCompile(`[^a-z0-9 ]+`)
	whitespaceRe = regexp.MustCompile(`\s+`)
)

// normalize lowercases, strips subtitles/parentheticals and punctuation, and
// collapses whitespace so titles like "Project Hail Mary: A Novel" and
// "Project Hail Mary (Unabridged)" compare closely.
func normalize(s string) string {
	s = strings.ToLower(s)
	s = subtitleRe.ReplaceAllString(s, "")
	s = nonAlnumRe.ReplaceAllString(s, "")
	s = whitespaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}
