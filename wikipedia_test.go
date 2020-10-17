package wikipedia_test

import (
	"testing"

	"github.com/sebnyberg/wikipedia"
	"google.golang.org/protobuf/proto"
)

func Test_ScoreArticles(t *testing.T) {
	a := &wikipedia.LinkedPage{
		PageTitle: "a",
		Links: []*wikipedia.Link{
			{TargetTitle: "b"},
			{TargetTitle: "c"},
			{TargetTitle: "d"},
		},
	}
	var b *wikipedia.LinkedPage = proto.Clone(a).(*wikipedia.LinkedPage)
	b.Links = append(b.Links, &wikipedia.Link{TargetTitle: "e"})

	score := wikipedia.Score(a, b)
	t.Fatal(score)
	if score < 1 {
		t.Fatal("hehe")
	}
}
