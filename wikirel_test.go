package wikirel_test

import (
	"testing"

	"github.com/sebnyberg/wikirel"
	"google.golang.org/protobuf/proto"
)

func Test_ScoreArticles(t *testing.T) {
	a := &wikirel.LinkedPage{
		PageTitle: "a",
		Links: []*wikirel.Link{
			{TargetTitle: "b"},
			{TargetTitle: "c"},
			{TargetTitle: "d"},
		},
	}
	var b *wikirel.LinkedPage = proto.Clone(a).(*wikirel.LinkedPage)
	b.Links = append(b.Links, &wikirel.Link{TargetTitle: "e"})

	score := wikirel.Score(a, b)
	t.Fatal(score)
	if score < 1 {
		t.Fatal("hehe")
	}
}
