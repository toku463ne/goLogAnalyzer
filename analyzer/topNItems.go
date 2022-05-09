package analyzer

func newTopNItems(n int) *topNItems {
	t := new(topNItems)
	t.n = n
	t.itemIDs = make([]int, n)
	t.scores = make([]float64, n)
	t.terms = make([]string, n)
	return t
}

func (t *topNItems) register(itemID int, score float64, term string) {
	if t.n == 0 {
		return
	}
	if score <= 0 || t.scores[t.n-1] > 0 && t.minScoreInTopN > score {
		return
	}
	if t.minScoreInTopN == 0 || score < t.minScoreInTopN {
		t.minScoreInTopN = score
	}
	newItemIDs := make([]int, t.n)
	newScores := make([]float64, t.n)
	newTerms := make([]string, t.n)
	j := 0
	for i := range newScores {
		if score >= t.scores[j] {
			newScores[i] = score
			newItemIDs[i] = itemID
			newTerms[i] = term
			score = -1
			continue
		}
		if t.scores[j] == 0 {
			break
		}
		if t.itemIDs[j] == itemID {
			j++
			if j >= t.n {
				break
			}
		}
		newScores[i] = t.scores[j]
		newItemIDs[i] = t.itemIDs[j]
		newTerms[i] = t.terms[j]
		j++
		if j >= t.n {
			break
		}
	}
	t.itemIDs = newItemIDs
	t.scores = newScores
	t.terms = newTerms
}
