package analyzer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	LogPath          string  `json:"path"`
	Search           string  `json:"search"`
	Exclude          string  `json:"exclude"`
	LinesInBlock     int     `json:"linesInBlock"`
	MaxBlocks        int     `json:"maxBlocks"`
	MaxItemBlocks    int     `json:"maxItemBlocks"`
	TopN             int     `json:"topN"`
	MinGapToRecord   float64 `json:"minGapToRecord"`
	DatetimeStartPos int     `json:"dateStart"`
	DatetimeLayout   string  `json:"dateLayout"`
	ScoreStyle       int     `json:"scoreStyle"`
}

type logSetInfo struct {
	DataDir    string             `json:"dataDir"`
	TopN       int                `json:"topN"`
	HistSize   int                `json:"histSize"`
	Logs       map[string]logInfo `json:"logs"`
	ScoreStyle int                `json:"scoreStyle"`
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
	defaultLinesInBlock, defaultNTopRecords, defaultHistSize int,
	outFormat string,
	defaultDatetimeStartPos int, defaultDatetimeLayout string) error {

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

	if ls.ScoreStyle == 0 {
		ls.ScoreStyle = cScoreNAvg
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
		reportPath := fmt.Sprintf("%s/%s.%s", reportDir, name, outFormat)

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
		if l.ScoreStyle == 0 {
			l.ScoreStyle = ls.ScoreStyle
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
			l.MaxBlocks, l.MaxItemBlocks, l.LinesInBlock, l.TopN,
			l.DatetimeStartPos, l.DatetimeLayout, l.ScoreStyle); err != nil {
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

		out := ""
		newLine := "\n"
		switch outFormat {
		case cFormatText:
			out = ""
		case cFormatHtml:
			out = "<html>"
			newLine = "<br>"
		}

		ephasizeLine := func(s string) string {
			switch outFormat {
			case cFormatHtml:
				s = fmt.Sprintf("<b>%s<b>", s)
			}
			return s
		}

		out2, g, err := a.stats.getCountPerStats(0, outFormat)
		if err != nil {
			log.Println(err)
			continue
		}
		out += out2

		if startEpoch > 0 {
			out2, g2, err := a.stats.getCountPerStats(startEpoch, outFormat)
			if err != nil {
				log.Println(err)
				continue
			}
			g = g2
			out += out2
		}

		sum := 0.0
		cnt := 0
		for _, v := range g {
			if v > 0 {
				sum += float64(v)
				cnt++
			}
		}
		mean := float64(sum) / float64(cnt)
		defaultMin := int(mean)
		if defaultMin == 0 {
			defaultMin = 1
		}
		border := mean

		//out += fmt.Sprintf("score border %f\n", border)
		ex := ""
		ex1 := fmt.Sprintf("(?i)(%s)", cErrorKeywords)
		ex2 := fmt.Sprintf("(?i)(%s|%s)", l.Exclude, cErrorKeywords)

		msg := ephasizeLine(fmt.Sprintf("%d top rare records", l.TopN))
		msg += newLine

		if l.Exclude == "" {
			ex = ex1
		} else {
			ex = ex2
		}
		out2, border1, err := a.getNTop(msg,
			l.TopN, startEpoch, 0,
			l.Search, ex, true, 0, 0, outFormat)
		if err != nil {
			log.Println(err)
			continue
		}
		if border1 > border {
			border = border1
		}
		out += out2
		out += newLine
		out += newLine

		inc := fmt.Sprintf("(?i)(%s)", cErrorKeywords)
		msg = fmt.Sprintf("%d top rare %s", l.TopN, cErrorKeywords)
		out3, _, err := a.getNTop(msg,
			ls.TopN, startEpoch, 0,
			inc, l.Exclude, true, 0, 0, outFormat)
		if err != nil {
			log.Println(err)
			continue
		}
		out += out3
		out += newLine
		out += newLine
		if out2, err := a.stats.getRecentStats(ls.HistSize, outFormat); err != nil {
			log.Println(err)
			continue
		} else {
			out += out2
		}

		switch outFormat {
		case cFormatHtml:
			out += "</html>"
		}

		println(out)
		if _, err := fw.WriteString(out); err != nil {
			log.Println(err)
			continue
		}

		_, topScore, err := a.getNTop("",
			l.TopN, startEpoch, 0,
			l.Search, ex, true, 0, 0, outFormat)
		if err != nil {
			log.Println(err)
			continue
		}
		if topScore > border {
			copyFile(reportPath, fmt.Sprintf("%s/%s.%s", abnormalDir, name, outFormat))
		}
	}
	return nil
}
