package logan

type logGroup struct {
	displayString string // lastValue with rare terms replaced with "*"
	count         int    // total count of this log group
	retentionPos  int64
	created       int64 // first epoch in the current block
	updated       int64 // last epoch in the current block
	countHistory  map[int64]int
	rareScore     float64
}

func (lg *logGroup) calcScore(tokens []int, te *terms) {
	scores := make([]float64, 0)
	for _, itemID := range tokens {
		if itemID >= 0 {
			scores = append(scores, te.getIdf(itemID))
		}
	}
	ss := 0.0
	for _, v := range scores {
		ss += v
	}
	score := 0.0
	if len(scores) > 0 {
		score = ss / float64(len(scores))
	}
	lg.rareScore = score
}
