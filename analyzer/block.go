package analyzer

type block struct {
	blockID       int
	blockCnt      int
	lastEpoch     int64
	scoreSum      float64
	scoreSqrSum   float64
	countPerScore []int
	//nTopRareLogs    []*logRec
	completed bool
}

func newBlock(blockID int) *block {
	b := new(block)
	b.blockID = blockID
	b.completed = false
	b.countPerScore = make([]int, cCountbyScoreLen)
	//b.nTopRareLogs = make([]*logRec, cNTopRareRecords)
	return b
}
