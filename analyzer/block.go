package analyzer

type block struct {
	blockID     int
	blockCnt    int
	lastEpoch   int64
	scoreSum    float64
	scoreSqrSum float64
	countPerGap []int
	//nTopRareLogs    []*logRec
	minTopRareScore float64
	completed       bool
}

func newBlock(blockID int) *block {
	b := new(block)
	b.blockID = blockID
	b.completed = false
	b.countPerGap = make([]int, cCountbyScoreLen)
	//b.nTopRareLogs = make([]*logRec, cNTopRareRecords)
	return b
}
