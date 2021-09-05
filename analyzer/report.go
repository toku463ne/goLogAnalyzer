package analyzer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type scoreHistJson struct {
	Date string  `json:"date"`
	Avg  float64 `json:"avg"`
	Std  float64 `json:"std"`
	Max  float64 `json:"max"`
}

type logInfo struct {
	LogPath        string  `json:"path"`
	Search         string  `json:"search"`
	Exclude        string  `json:"exclude"`
	LinesInBlock   int     `json:"linesInBlock"`
	MaxBlocks      int     `json:"maxBlocks"`
	MaxItemBlocks  int     `json:"maxItemBlocks"`
	TopN           int     `json:"topN"`
	MinGapToRecord float64 `json:"minGapToRecord"`
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

	if ls.TopN <= 0 {
		ls.TopN = defaultNTopRecords
	}

	if err := ensureDir(ls.DataDir); err != nil {
		return err
	}

	reportDir := fmt.Sprintf("%s/results", ls.DataDir)
	if err := ensureDir(reportDir); err != nil {
		return err
	}
	abnormalDir := fmt.Sprintf("%s/abnormals", ls.DataDir)
	if err := ensureDir(abnormalDir); err != nil {
		return err
	}

	for name, l := range ls.Logs {
		log.Printf("\n\nProcessing %s", name)
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
		if l.TopN <= 0 {
			l.TopN = ls.TopN
		}
		if ls.HistSize == 0 {
			ls.HistSize = defaultHistSize
		}
		if l.MinGapToRecord == 0.0 {
			l.MinGapToRecord = defaultMinGapToRecord
		}

		if match, err := filepath.Glob(l.LogPath); err != nil {
			log.Println(err)
			continue
		} else if match == nil {
			log.Printf("%s does not exist", l.LogPath)
			continue
		}

		if err := a.open(l.LogPath, l.Search, l.Exclude,
			defaultMinGapToRecord,
			l.MaxBlocks, l.MaxItemBlocks, l.LinesInBlock, l.TopN); err != nil {
			log.Println(err)
			continue
		}
		_, err := a.analyze(0)
		if err != nil {
			log.Println(err)
			continue
		}

		fw, err := os.Create(reportPath)
		if err != nil {
			log.Println(err)
			continue
		}
		defer fw.Close()

		out, g, err := a.stats.getCountPerStatsString()
		if err != nil {
			log.Println(err)
			continue
		}

		sum := 0.0
		sqrSum := 0.0
		cnt := len(g)
		for _, v := range g {
			sum += float64(v)
			sqrSum += float64(v * v)
		}
		mean := float64(sum) / float64(len(g))
		std := math.Sqrt((sqrSum - 2*sum*mean + float64(cnt)*mean*mean) / float64(cnt))
		defaultMin := int(mean - std*3)
		border := mean - std*3

		stages := make([]int, 0)
		min := defaultMin
		passedBottom := false
		for i := cnt - 1; i >= 0; i-- {
			if g[i] == 0 {
				continue
			}
			if min == 0 || (g[i] <= defaultMin && g[i] < min) {
				min = g[i]
				passedBottom = false
			}
			if g[i] > min || len(stages) == 0 {
				if !passedBottom {
					stages = append(stages, i)
				}
				min = defaultMin
				passedBottom = true
			}
		}
		//out += fmt.Sprintf("score border %f\n", border)
		ex := ""
		if l.Exclude == "" {
			ex = fmt.Sprintf("(?i)(%s)", cErrorKeywords)
		} else {
			ex = fmt.Sprintf("(?i)(%s|%s)", l.Exclude, cErrorKeywords)
		}
		for stagei, stage := range stages {
			out += ("\n")
			minStage := 0.0
			if stagei+1 < len(stages) {
				minStage = float64(stages[stagei+1])
			}

			msg := fmt.Sprintf("%d top rare records <= %d", l.TopN, stage+1)

			out2 := ""
			out2, _, err = a.getNTopString(msg,
				l.TopN, startEpoch, 0,
				l.Search, ex, true, minStage, float64(stage+1))
			if err != nil {
				log.Println(err)
				continue
			}
			out += out2
			out += "\n"
			out += "\n"

			inc := fmt.Sprintf("(?i)(%s)", cErrorKeywords)
			msg = fmt.Sprintf("%d top rare %s <= %d", l.TopN, cErrorKeywords, stage+1)
			out3 := ""
			out3, _, err = a.getNTopString(msg,
				ls.TopN, startEpoch, 0,
				inc, l.Exclude, true, minStage, float64(stage+1))
			if err != nil {
				log.Println(err)
				continue
			}
			out += out3
			out += "\n"
			out += "\n"

			if out2, err := a.stats.getRecentStatsString(ls.HistSize); err != nil {
				log.Println(err)
				continue
			} else {
				out += out2
			}
		}

		println(out)
		if _, err := fw.WriteString(out); err != nil {
			log.Println(err)
			continue
		}

		_, topScore, err := a.getNTopString("",
			l.TopN, startEpoch, 0,
			l.Search, ex, true, 0, 0)
		if err != nil {
			log.Println(err)
			continue
		}
		if topScore > border {
			copyFile(reportPath, fmt.Sprintf("%s/%s.txt", abnormalDir, name))
		}
	}
	return nil
}
