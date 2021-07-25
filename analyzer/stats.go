package analyzer

import (
	"fmt"
	"math"

	"github.com/pkg/errors"
)

func newColStats() *colStats {
	s := new(colStats)
	return s
}

func newStats(dataDir string, maxBlocks, maxRowsInBlock int) (*stats, error) {
	s := new(stats)
	s.colStats = newColStats()
	s.currBlock = newColStats()
	s.countPerScore = make([]int, cCountbyScoreLen)
	s.maxBlocks = maxBlocks
	s.maxRowsInBlock = maxRowsInBlock
	s.blockNo = 0
	s.rowNo = 0
	s.seqNo = 0
	if dataDir != "" {
		d, err := newDB(dataDir, "stats")
		if err != nil {
			return nil, err
		}
		s.db = d
		if err := s.prepareTables(); err != nil {
			return nil, err
		}
		if err := s.loadDB(); err != nil {
			return nil, err
		}
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

func (s *stats) registerScore(score float64) error {
	scoreSqr := score * score
	scoreStage := s.getScoreStage(score)

	s.scoreCount++
	s.scoreSum += score
	s.scoreSqrSum += scoreSqr
	s.countPerScore[scoreStage]++

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
	if s.db != nil {
		s.db.close()
	}
}

func (s *stats) nextBlock() error {
	if s.dataDir != "" {
		err := s.commit(true)
		if err != nil {
			return err
		}
	}

	s.currBlock = newColStats()
	s.countPerScore = make([]int, cCountbyScoreLen)
	s.blockNo++
	s.rowNo = 0
	if s.blockNo >= s.maxBlocks {
		s.blockNo = 0
	}
	return nil
}

func (s *stats) prepareTables() error {
	if err := s.createTable("statistics"); err != nil {
		return err
	}
	if err := s.createTable("scores"); err != nil {
		return err
	}
	return nil
}

func (s *stats) commit(completed bool) error {
	if s.rowNo == 0 {
		return nil
	}

	cb := s.currBlock

	stmt, err := s.conn.Prepare(`REPLACE INTO scores(
seqNo, blockNo, rowCount, scoreSum, scoreSqrSum, completed) 
VALUES(?,?,?,?,?,?);`)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = stmt.Exec(s.seqNo, s.blockNo, s.rowNo,
		cb.scoreSum, cb.scoreSqrSum, completed)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = s.exec(fmt.Sprintf(`DELETE FROM statistics WHERE seqNo=%d AND blockNo=%d`,
		s.seqNo, s.blockNo))
	if err != nil {
		return errors.WithStack(err)
	}

	stmt, err = s.conn.Prepare(`INSERT INTO statistics(
		seqNo, blockNo, scoreStage, itemCount) VALUES(?,?,?,?);`)
	if err != nil {
		return errors.WithStack(err)
	}

	for i, cnt := range s.countPerScore {
		if cnt == 0 {
			continue
		}
		_, err = stmt.Exec(s.seqNo, s.blockNo, i, cnt)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	s.seqNo++

	return nil
}

func (s *stats) blockExists(blockNo int) bool {
	cnt := s.count("scores", fmt.Sprintf("blockNo = %d", blockNo))
	return cnt > 0
}

func (s *stats) loadDB() error {
	cb := s.currBlock
	var completed bool

	if cnt := s.count("scores", ""); cnt == 0 {
		return nil
	}

	if err := s.select1rec(`SELECT 
seqNo, blockNo, rowCount, scoreSum, scoreSqrSum, completed 
FROM scores 
WHERE seqNo = (SELECT MAX(seqNo) FROM scores);`,
		&s.seqNo, &s.blockNo, &s.rowNo, &cb.scoreSum, &cb.scoreSqrSum, &completed); err != nil {
		return err
	}
	s.currBlock.scoreCount = int64(s.rowNo)

	rows, err := s.query(`SELECT rowCount, scoreSum, scoreSqrSum FROM scores;)`)
	if err != nil {
		return err
	}
	var scoreCount int64
	var scoreSum float64
	var scoreSqrSum float64
	for rows.Next() {
		if err := rows.Scan(&scoreCount, &scoreSum, &scoreSqrSum); err != nil {
			return errors.WithStack(err)
		}
		s.scoreCount += scoreCount
		s.scoreSum += scoreSum
		s.scoreSqrSum += scoreSqrSum
	}

	if completed {
		return nil
	}

	var scoreStage int
	var itemCount int
	rows, err = s.query(fmt.Sprintf(`SELECT scoreStage, itemCount 
FROM statistics
WHERE seqNo=%d AND blockNo=%d;)`, s.seqNo, s.blockNo))
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
