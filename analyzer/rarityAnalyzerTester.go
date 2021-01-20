package analyzer

type rarityAnalyzerTester struct {
	rarityAnalyzer
	movingScore  []float64
	movingSum    float64
	movingSqrSum float64
	pos          int
}

func newRarityAnalyzerTester(logPathRegex,
	rootDir string,
	filterRe, xFilterRe string,
	gapThreshold, rowsRateToShow float64,
	linesInBlock, maxBlocks int) (*rarityAnalyzerTester, error) {

	at := new(rarityAnalyzerTester)
	at.movingScore = make([]float64, linesInBlock)
	at.pos = -1

	if err := at.init(logPathRegex,
		rootDir,
		filterRe, xFilterRe,
		gapThreshold,
		linesInBlock, maxBlocks); err != nil {
		return nil, err
	}

	at.outputRes = func(rowID int64, score, scoreGap, scoreAvg, scoreStd float64, text string) error {
		at.pos++
		if at.pos >= at.linesInBlock {
			at.pos = -1
		}
		oldScore := at.movingScore[at.pos]
		at.movingSum -= oldScore
		at.movingSqrSum -= oldScore * oldScore
		at.movingScore[at.pos] = score

		return nil
	}

	return at, nil
}
