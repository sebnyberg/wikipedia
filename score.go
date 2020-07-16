package wikirel

func Score(a *LinkedPage, b *LinkedPage) float32 {
	aLinks := make(map[string]bool, len(a.Links))
	aLinks[a.Title] = true
	for _, link := range a.Links {
		aLinks[link.TargetTitle] = true
	}

	bLinks := make(map[string]bool, len(b.Links))
	bLinks[b.Title] = true
	for _, link := range b.Links {
		bLinks[link.TargetTitle] = true
	}

	var score float32 = 0
	for title := range bLinks {
		if _, exists := aLinks[title]; exists {
			score++
		}
	}

	score = score / float32(len(b.Links))

	return score
}
