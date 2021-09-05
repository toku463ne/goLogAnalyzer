package analyzer

import (
	"fmt"
	"math"
	"strconv"

	"github.com/pkg/errors"
	csvdb "github.com/toku463ne/goCsvDb"
)

func newColStats() *colStats {
	s := new(colStats)
	return s
}

func newStats(rootDir string, maxBlocks, maxRowsInBlock int) (*stats, error) {
	s := new(stats)
	s.colStats = newColStats()
	s.currBlock = newColStats()
	s.countPerScore = make([]int, cCountbyScoreLen)
	s.maxBlocks = maxBlocks
	s.maxRowsInBlock = maxRowsInBlock
	s.blockNo = 0
	s.rowNo = 0
	s.seqNo = 0
	s.rootDir = rootDir
	s.scoreMax = -1
	if s.rootDir != "" {
		db, err := csvdb.NewCsvDB(rootDir)
		if err != nil {
			return nil, err
		}
		s.CsvDB = db
		if err := s.prepareTables(); err != nil {
			return nil, err
		}
		//if err := s.load(); err != nil {
		//	return nil, err
		//}
	}

	return s, nil
}

func (s *stats) getScoreStage(score float64) int {
	scoreStage := int(math.Floor(score))
	if scoreStage < 0 {
		scoreStage = 0
	}
	if scoreStage >= cCountbyScoreLen {
		scoreStage = cCountbyScoreLen - 1
	}
	return scoreStage
}

func (s *stats) registerScore(score float64, fileEpoch int64) error {
	scoreSqr := score * score
	scoreStage := s.getScoreStage(score)

	s.scoreCount++
	s.scoreSum += score
	s.scoreSqrSum += scoreSqr
	s.countPerScore[scoreStage]++
	s.lastFileEpoch = fileEpoch

	cnt := float64(s.scoreCount)
	//score avg
	sa := s.scoreSum / cnt
	//score std
	ss := math.Sqrt((s.scoreSqrSum - 2*s.scoreSum*sa + cnt*sa*sa) / cnt)

	if ss > 0 {
		s.lastGap = (score - sa) / (ss)
	} else {
		s.lastGap = 0
	}
	s.lastAverage = sa
	s.lastStd = ss

	if s.scoreMax == -1 || s.scoreMax < score {
		s.scoreMax = score
	}

	s.currBlock.scoreCount++
	s.currBlock.scoreSum += score
	s.currBlock.scoreSqrSum += scoreSqr

	s.rowNo++

	if s.rowNo >= s.maxRowsInBlock {
		err := s.nextBlock()
		return err
	}
	return nil
}

func (s *stats) close() {
	s.CsvDB = nil
}

func (s *stats) nextBlock() error {
	if s.rootDir != "" {
		err := s.commit(true)
		if err != nil {
			return err
		}
	}

	s.currBlock = newColStats()
	if s.rootDir != "" {
		s.countPerScore = make([]int, cCountbyScoreLen)
		s.scoreMax = 0
	}
	s.blockNo++
	s.rowNo = 0
	if s.blockNo >= s.maxBlocks {
		s.blockNo = 0
	}
	return nil
}

func (s *stats) prepareTables() error {
	if t, err := s.CreateTableIfNotExists("statistics",
		tableDefs["statistics"], false, cDefaultBuffSize); err != nil {
		return err
	} else {
		s.statsTable = t
	}
	if t, err := s.CreateTableIfNotExists("scores",
		tableDefs["scores"], false, s.maxBlocks); err != nil {
		return err
	} else {
		s.scoresTable = t
	}
	if t, err := s.CreateTableIfNotExists("scoresHist",
		tableDefs["scoresHist"], false, s.maxBlocks); err != nil {
		return err
	} else {
		s.scoresHistTable = t
	}
	return nil
}

func (s *stats) commit(completed bool) error {
	if s.rootDir == "" {
		return nil
	}
	if s.rowNo == 0 {
		return nil
	}

	cb := s.currBlock
	blockIdx := s.scoresTable.GetColIdx("blockNo")
	if err := s.scoresTable.Upsert(func(v []string) bool {
		return v[blockIdx] == strconv.Itoa(s.blockNo)
	}, map[string]interface{}{
		"seqNo":         s.seqNo,
		"blockNo":       s.blockNo,
		"rowCount":      s.rowNo,
		"scoreSum":      cb.scoreSum,
		"scoreMax":      s.scoreMax,
		"scoreSqrSum":   cb.scoreSqrSum,
		"completed":     completed,
		"lastFileEpoch": s.lastFileEpoch,
	}); err != nil {
		return errors.WithStack(err)
	}
	blockNoidx := s.statsTable.GetColIdx("blockNo")
	scoreStageIdx := s.statsTable.GetColIdx("scoreStage")
	for i, cnt := range s.countPerScore {
		if cnt == 0 {
			continue
		}
		if err := s.statsTable.Upsert(func(v []string) bool {
			return v[blockNoidx] == strconv.Itoa(s.blockNo) && v[scoreStageIdx] == strconv.Itoa(i)
		}, map[string]interface{}{
			"seqNo":         s.seqNo,
			"blockNo":       s.blockNo,
			"scoreStage":    i,
			"itemCount":     cnt,
			"lastFileEpoch": s.lastFileEpoch,
		}); err != nil {
			return errors.WithStack(err)
		}
	}

	scoreMax := 0.0
	if err := s.scoresTable.Max(nil, "scoreMax", &scoreMax); err != nil {
		return errors.WithStack(err)
	}

	blockNoidx = s.scoresHistTable.GetColIdx("blockNo")
	if err := s.scoresHistTable.Upsert(func(v []string) bool {
		return v[blockNoidx] == strconv.Itoa(int(s.blockNo))
	}, map[string]interface{}{
		"seqNo":         s.seqNo,
		"blockNo":       s.blockNo,
		"avg":           s.lastAverage,
		"std":           s.lastStd,
		"max":           s.scoreMax,
		"lastFileEpoch": s.lastFileEpoch,
	}); err != nil {
		return errors.WithStack(err)
	}

	s.seqNo++

	return nil
}

func (s *stats) load(use_weigth bool) error {
	if s.rootDir == "" {
		return nil
	}
	cb := s.currBlock
	var completed bool

	if cnt := s.scoresTable.Count(nil); cnt == 0 {
		return nil
	}

	ma := 0.0
	if err := s.scoresTable.Max(nil, "seqNo", &ma); err != nil {
		return errors.WithStack(err)
	}

	seqIdx := s.scoresTable.GetColIdx("seqNo")
	maxStr := strconv.Itoa(int(ma))
	if err := s.scoresTable.Select1Row(func(v []string) bool {
		return v[seqIdx] == maxStr
	}, []string{"seqNo", "blockNo", "rowCount", "scoreSum", "scoreSqrSum", "completed"},
		&s.seqNo, &s.blockNo, &s.rowNo, &cb.scoreSum, &cb.scoreSqrSum, &completed); err != nil {
		return err
	}

	s.currBlock.scoreCount = int64(s.rowNo)

	rows, err := s.scoresTable.SelectRows(nil, []string{"seqNo", "rowCount", "scoreSum", "scoreSqrSum"})
	if err != nil {
		return err
	}
	if err := rows.OrderBy([]string{"seqNo"}, []string{"int"},
		csvdb.CorderByAsc); err != nil {
		return err
	}

	var seqNo int
	var scoreCount int64
	var scoreSum float64
	var scoreSqrSum float64
	idx := 0
	w := 1
	oldSeqNo := -1
	for rows.Next() {
		if use_weigth {
			if oldSeqNo == -1 || oldSeqNo != seqNo {
				idx++
				w = getWeight(idx)
			}
		}
		if err := rows.Scan(&seqNo, &scoreCount, &scoreSum, &scoreSqrSum); err != nil {
			return errors.WithStack(err)
		}
		s.scoreCount += scoreCount * int64(w)
		s.scoreSum += scoreSum * float64(w)
		s.scoreSqrSum += scoreSqrSum * float64(w)
	}

	if completed {
		return nil
	}

	seqIdx = s.statsTable.GetColIdx("seqNo")
	blockNoIdx := s.statsTable.GetColIdx("blockNo")
	var scoreStage int
	var itemCount int
	seqNoStr := strconv.Itoa(int(s.seqNo))
	blockNoStr := strconv.Itoa(s.blockNo)
	rows, err = s.statsTable.SelectRows(func(v []string) bool {
		return v[seqIdx] == seqNoStr && v[blockNoIdx] == blockNoStr
	}, []string{"scoreStage", "itemCount"})
	if err != nil {
		return err
	}
	for rows.Next() {
		if err := rows.Scan(&scoreStage, &itemCount); err != nil {
			return errors.WithStack(err)
		}
		s.countPerScore[scoreStage] += itemCount
	}

	return nil
}

func (s *stats) loadAllScorePerCount(use_weigth bool) ([]int, error) {
	var scoreStage int
	var itemCount int
	var seqNo int
	rows, err := s.statsTable.SelectRows(nil, []string{"seqNo", "scoreStage", "itemCount"})
	if err != nil {
		return nil, err
	}

	if err := rows.OrderBy([]string{"seqNo"}, []string{"int"},
		csvdb.CorderByAsc); err != nil {
		return nil, err
	}

	countPerScore := make([]int, cCountbyScoreLen)
	idx := 0
	w := 0
	oldSeqNo := -1
	for rows.Next() {
		if oldSeqNo == -1 || oldSeqNo != seqNo {
			idx++
			w = getWeight(idx)
		}
		if err := rows.Scan(&seqNo, &scoreStage, &itemCount); err != nil {
			return nil, errors.WithStack(err)
		}
		if use_weigth {
			countPerScore[scoreStage] += itemCount * w
		} else {
			countPerScore[scoreStage] += itemCount
		}

	}
	return countPerScore, nil
}

func (s *stats) loadRecentStats(showCounts int) ([]colScoresHist, error) {
	var lastFileEpoch int64
	var avg float64
	var std float64
	var max float64
	if showCounts == 0 {
		return nil, nil
	}
	rows, err := s.scoresHistTable.SelectRows(nil,
		[]string{"lastFileEpoch", "avg", "std", "max"})
	if err != nil {
		return nil, err
	}

	err = rows.OrderBy([]string{"seqNo"}, []string{"int"}, csvdb.CorderByDesc)
	if err != nil {
		return nil, err
	}

	scoresHists := make([]colScoresHist, showCounts)
	oldScore := new(colScoresHist)
	i := 0
	for rows.Next() {
		if err := rows.Scan(&lastFileEpoch, &avg, &std, &max); err != nil {
			return nil, errors.WithStack(err)
		}
		if oldScore.lastFileEpoch == 0 || oldScore.lastFileEpoch != lastFileEpoch {
			scoresHists[i] = colScoresHist{lastFileEpoch, avg, std, max}
			oldScore = &scoresHists[i]
			i++
		}
		if i >= showCounts {
			break
		}
	}
	return scoresHists, nil
}

func (s *stats) getCountPerStatsString() (string, []int, error) {
	var g []int
	var err error
	if s.rootDir == "" {
		g = s.countPerScore
	} else {
		g, err = s.loadAllScorePerCount(false)
		if err != nil {
			return "", nil, err
		}
	}

	out := "\n"
	out += "Counts per score\n"
	out += " score | count\n"

	out += " ------+--------------\n"
	for i := 0; i < cCountbyScoreLen; i++ {
		if g[i] > 0 {
			out += fmt.Sprintf(" %5.1f | %12d\n", float64(i), g[i])
		}
	}
	out += "\n"
	out += "\n"

	return out, g, nil
}

func (s *stats) getRecentStatsString(histSize int) (string, error) {
	h, err := s.loadRecentStats(histSize)
	if err != nil {
		return "", err
	}
	out := "\n"
	out += "score history\n"
	out += " last date           | average |     std |     max \n"
	out += " --------------------+---------+---------+---------\n"
	for _, rec := range h {
		if rec.lastFileEpoch == 0 {
			continue
		}
		out += fmt.Sprintf(" %s | %7.1f | %7.1f | %7.1f \n",
			epochToString(rec.lastFileEpoch), rec.avg, rec.std, rec.max)
	}
	return out, nil
}
