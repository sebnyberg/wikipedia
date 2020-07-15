package wikirel_test

import (
	"testing"

	"github.com/sebnyberg/wikirel"
)

func Test_ScoreArticles(t *testing.T) {
	a := &wikirel.Page{
		Title: "a",
		Links: []*wikirel.Link{
			{TargetTitle: "b"},
			{TargetTitle: "c"},
			{TargetTitle: "d"},
		},
	}
	b := new(wikirel.Page)
	*b = *a
	b.Links = append(b.Links, &wikirel.Link{TargetTitle: "e"})

	score := wikirel.Score(a, b)
	t.Fatal(score)
	if score < 1 {
		t.Fatal("hehe")
	}
}
