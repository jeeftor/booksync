package matcher

import (
	"testing"

	"github.com/jeeftor/bookSync/internal/absclient"
	"github.com/jeeftor/bookSync/internal/kindleclient"
)

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"Project Hail Mary: A Novel":          "project hail mary",
		"Project Hail Mary (Unabridged)":      "project hail mary",
		"  Dune   ":                           "dune",
		"The Hobbit, or There and Back Again": "the hobbit or there and back again",
	}
	for in, want := range cases {
		if got := normalize(in); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSimilarity(t *testing.T) {
	if s := similarity("dune", "dune"); s != 1 {
		t.Errorf("identical strings should score 1, got %v", s)
	}
	if s := similarity("", ""); s != 1 {
		t.Errorf("two empty strings should score 1, got %v", s)
	}
	if s := similarity("dune", "xxxx"); s != 0 {
		t.Errorf("fully mismatched same-length strings should score 0, got %v", s)
	}
	close := similarity("project hail mary", "project hail mary")
	far := similarity("project hail mary", "the hobbit")
	if close <= far {
		t.Errorf("expected close match (%v) to outscore unrelated title (%v)", close, far)
	}
}

func TestSuggestPicksBestMatchAboveThreshold(t *testing.T) {
	kindleBooks := []kindleclient.Book{
		{ASIN: "B001", Title: "Project Hail Mary", Authors: []string{"Andy Weir"}},
		{ASIN: "B002", Title: "Some Totally Unrelated Kindle Book", Authors: []string{"Nobody"}},
	}
	absItems := []absclient.LibraryItem{
		{ID: "abs-1", Title: "Project Hail Mary (Unabridged)", Authors: []string{"Andy Weir"}},
		{ID: "abs-2", Title: "Dune", Authors: []string{"Frank Herbert"}},
	}

	got := Suggest(kindleBooks, absItems, nil)

	if len(got) != 1 {
		t.Fatalf("expected exactly 1 suggestion above MinConfidence, got %d: %+v", len(got), got)
	}
	c := got[0]
	if c.KindleASIN != "B001" || c.ABSItemID != "abs-1" {
		t.Errorf("expected B001 <-> abs-1, got %s <-> %s", c.KindleASIN, c.ABSItemID)
	}
	if c.Confidence < MinConfidence {
		t.Errorf("returned candidate confidence %v below MinConfidence %v", c.Confidence, MinConfidence)
	}
}

func TestSuggestSortedByConfidenceDescending(t *testing.T) {
	kindleBooks := []kindleclient.Book{
		{ASIN: "B001", Title: "Dune", Authors: []string{"Frank Herbert"}},
		{ASIN: "B002", Title: "Project Hail Mary", Authors: []string{"Andy Weir"}},
	}
	absItems := []absclient.LibraryItem{
		{ID: "abs-1", Title: "Dune (Unabridged)", Authors: []string{"Frank Herbert"}},
		{ID: "abs-2", Title: "Project Hail Mary: A Novel", Authors: []string{"Andy Weir"}},
	}

	got := Suggest(kindleBooks, absItems, nil)
	for i := 1; i < len(got); i++ {
		if got[i].Confidence > got[i-1].Confidence {
			t.Fatalf("results not sorted by descending confidence: %+v", got)
		}
	}
}

func TestSuggestSkipsRejectedPairAndFallsBackToNextBest(t *testing.T) {
	kindleBooks := []kindleclient.Book{
		{ASIN: "B001", Title: "Dune", Authors: []string{"Frank Herbert"}},
	}
	absItems := []absclient.LibraryItem{
		{ID: "abs-1", Title: "Dune (Unabridged)", Authors: []string{"Frank Herbert"}},
		{ID: "abs-2", Title: "Dune (Special Edition)", Authors: []string{"Frank Herbert"}},
	}

	isRejected := func(asin, absItemID string) bool {
		return asin == "B001" && absItemID == "abs-1"
	}

	got := Suggest(kindleBooks, absItems, isRejected)
	if len(got) != 1 {
		t.Fatalf("expected 1 suggestion, got %d: %+v", len(got), got)
	}
	if got[0].ABSItemID != "abs-2" {
		t.Errorf("expected fallback to abs-2 once abs-1 rejected, got %s", got[0].ABSItemID)
	}
}

func TestSuggestExcludesBelowThreshold(t *testing.T) {
	kindleBooks := []kindleclient.Book{
		{ASIN: "B001", Title: "Completely Unrelated Title", Authors: []string{"Someone"}},
	}
	absItems := []absclient.LibraryItem{
		{ID: "abs-1", Title: "Nothing Alike Whatsoever", Authors: []string{"Someone Else"}},
	}

	got := Suggest(kindleBooks, absItems, nil)
	if len(got) != 0 {
		t.Fatalf("expected no suggestions below MinConfidence, got %+v", got)
	}
}
