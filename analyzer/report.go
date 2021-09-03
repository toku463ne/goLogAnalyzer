package analyzer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
)

type scoreHistJson struct {
	Date string  `json:"date"`
	Avg  float64 `json:"avg"`
	Std  float64 `json:"std"`
	Max  float64 `json:"max"`
}

type logInfo struct {
	LogPath       string `json:"path"`
	Search        string `json:"search"`
	Exclude       string `json:"exclude"`
	LinesInBlock  int    `json:"linesInBlock"`
	MaxBlocks     int    `json:"maxBlocks"`
	MaxItemBlocks int    `json:"maxItemBlocks"`
}

type logSetInfo struct {
	DataDir  string             `json:"dataDir"`
	TopN     int                `json:"topN"`
	HistSize int                `json:"histSize"`
	Logs     map[string]logInfo `json:"logs"`
}

func newLogSetInfo(jsonFile string) (*logSetInfo, error) {
	bytes, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}
	var ls *logSetInfo
	if err := json.Unmarshal(bytes, &ls); err != nil {
		return nil, err
	}
	return ls, nil
}

func (ls *logSetInfo) run(recentNdays int,
	defaultMinGapToRecord float64,
	defaultMaxBlocks, defaultMaxItemBlocks,
	defaultLinesInBlock, defaultNTopRecords, defaultHistSize int) error {

	startEpoch := int64(0)
	if recentNdays > 0 {
		startEpoch = getCurrentEpoch() - int64(24*60*60*recentNdays)
	}

	if ls.DataDir == "" {
		return errors.New("DataDir cannot be empty")
	}

	if err := ensureDir(ls.DataDir); err != nil {
		return err
	}

	reportDir := fmt.Sprintf("%s/results", ls.DataDir)
	if err := ensureDir(reportDir); err != nil {
		return err
	}

	for name, l := range ls.Logs {
		log.Printf("Processing %s", name)
		dataPath := fmt.Sprintf("%s/%s", ls.DataDir, name)
		reportPath := fmt.Sprintf("%s/%s.txt", reportDir, name)

		a := newRarityAnalyzer(dataPath)

		if l.MaxBlocks <= 0 {
			l.MaxBlocks = defaultMaxBlocks
		}
		if l.MaxItemBlocks <= 0 {
			l.MaxItemBlocks = defaultMaxItemBlocks
		}
		if l.LinesInBlock <= 0 {
			l.LinesInBlock = defaultLinesInBlock
		}
		if ls.TopN <= 0 {
			ls.TopN = defaultNTopRecords
		}
		if ls.HistSize == 0 {
			ls.HistSize = defaultHistSize
		}

		if err := a.open(l.LogPath, l.Search, l.Exclude,
			defaultMinGapToRecord,
			l.MaxBlocks, l.MaxItemBlocks, l.LinesInBlock, ls.TopN); err != nil {
			return err
		}
		_, err := a.analyze(0)
		if err != nil {
			return err
		}

		fw, err := os.Create(reportPath)
		if err != nil {
			return err
		}
		defer fw.Close()

		out, border, err := a.stats.getCountPerStatsString()
		if err != nil {
			return err
		}

		out += fmt.Sprintf("score border %f\n", border)

		out += ("\n")
		msg := fmt.Sprintf("%d top rare records", ls.TopN)
		out2, err := a.getNTopString(msg,
			ls.TopN, startEpoch, 0,
			l.Search, l.Exclude, true)
		if err != nil {
			return err
		}
		out += out2
		out += "\n"
		out += "\n"

		if out2, err := a.stats.getRecentStatsString(ls.HistSize); err != nil {
			return err
		} else {
			out += out2
		}

		println(out)
		if _, err := fw.WriteString(out); err != nil {
			return err
		}
	}
	return nil
}
