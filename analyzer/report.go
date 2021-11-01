package analyzer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func newReports() *reports {
	rs := new(reports)
	rs.rep = make(map[string]*report)
	return rs
}

func (rs *reports) run(jsonFile string, recentNdays int, outFormat string,
	defaultMinGapToRecord float64,
	defaultMaxBlocks, defaultMaxItemBlocks,
	defaultLinesInBlock, defaultNTopRecords, defaultHistSize int,
	defaultDatetimeStartPos int, defaultDatetimeLayout string) error {
	lm, err := newLogInfoMap(jsonFile)
	if err != nil {
		return err
	}

	if rs.dataDir == "" {
		rs.dataDir = lm.DataDir
	}

	if lm.DataDir == "" {
		lm.DataDir = rs.dataDir
	}

	if lm.TopN <= 0 {
		lm.TopN = defaultNTopRecords
	}
	if lm.HistSize <= 0 {
		lm.HistSize = defaultHistSize
	}
	if lm.ScoreStyle <= 0 {
		lm.ScoreStyle = cScoreNDistAvg
	}
	if lm.LinesInBlock <= 0 {
		lm.LinesInBlock = defaultLinesInBlock
	}
	if lm.MaxBlocks <= 0 {
		lm.MaxBlocks = defaultMaxBlocks
	}
	if lm.MaxItemBlocks <= 0 {
		lm.MaxItemBlocks = defaultMaxItemBlocks
	}
	if lm.MinGapToRecord <= 0 {
		lm.MinGapToRecord = defaultMinGapToRecord
	}

	startEpoch := int64(0)
	if recentNdays > 0 {
		startEpoch = getCurrentEpoch() - int64(24*60*60*recentNdays)
	}

	for name, l := range lm.Logs {
		if l.DataDir == "" && lm.DataDir == "" {
			log.Print("LogDir must not be empty")
		}
		if l.DataDir == "" {
			l.DataDir = lm.DataDir
		}
		if l.LogPath == "" && lm.LogPath == "" {
			log.Print("LogPath must not be empty")
		}
		if l.LogPath == "" {
			l.LogPath = lm.LogPath
		}
		if match, err := filepath.Glob(l.LogPath); err != nil {
			log.Println(err)
			continue
		} else if match == nil {
			log.Printf("%s does not exist", l.LogPath)
			continue
		}

		if l.TopN <= 0 {
			l.TopN = lm.TopN
		}
		if l.HistSize <= 0 {
			l.HistSize = lm.HistSize
		}
		if l.ScoreStyle <= 0 {
			l.ScoreStyle = lm.ScoreStyle
		}
		if l.LinesInBlock <= 0 {
			l.LinesInBlock = lm.LinesInBlock
		}
		if l.MaxBlocks <= 0 {
			l.MaxBlocks = lm.MaxBlocks
		}
		if l.MaxItemBlocks <= 0 {
			l.MaxItemBlocks = lm.MaxItemBlocks
		}
		if l.MinGapToRecord <= 0 {
			l.MinGapToRecord = lm.MinGapToRecord
		}

		r, err := rs.runAnalyzer(name, startEpoch, l)
		if err != nil {
			log.Printf("%+v", err)
			continue
		}
		if outFormat == cFormatHtml {
			if err := r.writeHtmlReport(); err != nil {
				log.Printf("%+v", err)
			}
		}
		rs.rep[name] = r
	}
	if outFormat == cFormatHtml {
		if err := rs.writeHtmlDiffSummary(); err != nil {
			return err
		}
	}
	return nil
}

func (rs *reports) runAnalyzer(name string, startEpoch int64, l LogInfo) (*report, error) {
	log.Printf("Processing %s", name)
	dataPath := fmt.Sprintf("%s/%s", l.DataDir, name)
	a := newRarityAnalyzer(dataPath)
	if err := a.open(l.LogPath, "", "",
		l.MinGapToRecord,
		l.MaxBlocks, l.MaxItemBlocks, l.LinesInBlock, l.TopN,
		l.DatetimeStartPos, l.DatetimeLayout, l.ScoreStyle); err != nil {
		return nil, err
	}
	_, err := a.analyze(0)
	if err != nil {
		return nil, err
	}

	r := new(report)
	r.name = name
	r.info = l

	ex := ""
	ex1 := fmt.Sprintf("(?i)(%s)", cErrorKeywords)
	ex2 := fmt.Sprintf("(?i)(%s|%s)", l.Exclude, cErrorKeywords)

	if l.Exclude == "" {
		ex = ex1
	} else {
		ex = ex2
	}

	minScore := 0.0

	ntopNorm, err := a.getNTop(name, l.TopN, startEpoch, 0,
		l.Search, ex, minScore, 0)
	if err != nil {
		return nil, err
	}
	if err := ntopNorm.save(); err != nil {
		return nil, err
	}

	r.nTopNorm = ntopNorm

	inc := fmt.Sprintf("(?i)(%s)", cErrorKeywords)

	r.includePhrase = l.Search

	nTopErr, err := a.getNTop(fmt.Sprintf("%s_errors", name), l.TopN, startEpoch, 0,
		inc, l.Exclude, minScore, 0)
	if err != nil {
		return nil, err
	}
	if err := nTopErr.save(); err != nil {
		return nil, err
	}
	r.nTopErr = nTopErr
	r.st = a.stats
	log.Printf("Completed %s\n", name)
	return r, nil
}

func (r *report) writeHtmlReport() error {
	out := "<html>"

	// count per stats
	ost, _, err := r.st.getCountPerStatsHtml(0)
	if err != nil {
		return err
	}
	out += ost
	out += "<br>"

	// normal topN logs
	if r.nTopNorm.getLen() > 0 {
		ote, _, err := r.nTopNorm.nTop2html(fmt.Sprintf("%d top %s",
			r.info.TopN, r.includePhrase), r.info.TopN)
		if err != nil {
			return err
		}
		out += ote
		out += "<br>"

		if r.nTopNorm.diff.getLen() > 0 {
			ote, _, err := r.nTopNorm.diff.nTop2html(fmt.Sprintf("[diff] %d top %s",
				r.info.TopN, r.includePhrase), r.info.TopN)
			if err != nil {
				return err
			}
			out += ote
			out += "<br>"
		}
	}

	// error topN logs
	if r.nTopErr.getLen() > 0 {
		ote, _, err := r.nTopErr.nTop2html(fmt.Sprintf("%d top %s",
			r.info.TopN, cErrorKeywords), r.info.TopN)
		if err != nil {
			return err
		}
		out += ote
		out += "<br>"

		if r.nTopErr.diff.getLen() > 0 {
			ote, _, err := r.nTopErr.diff.nTop2html(fmt.Sprintf("[diff] %d top %s",
				r.info.TopN, cErrorKeywords), r.info.TopN)
			if err != nil {
				return err
			}
			out += ote
			out += "<br>"
		}
	}

	out += "</html>"
	reportPath := fmt.Sprintf("%s/results", r.info.DataDir)
	if err := ensureDir(reportPath); err != nil {
		return err
	}
	filePath := fmt.Sprintf("%s/%s.html", reportPath, r.name)
	fw, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fw.Close()
	if _, err := fw.WriteString(out); err != nil {
		return err
	}
	return nil
}

func (rs *reports) writeHtmlDiffSummary() error {
	log.Println("Processing diff summary")
	out := "<html>"
	out += "<table border=1 ~~~ style='table-layout:fixed;width:100%;'><tr><td width=8%>name</td><width=4%>count</td><width=6%>score</td><td>text</td></tr>"
	subNames := []string{"errors", ""}
	for _, r := range rs.rep {
		for i, ntop := range []*nTopRecords{r.nTopErr, r.nTopNorm} {
			diffRecs := ntop.getDiffRecords()
			if len(diffRecs) == 0 {
				continue
			}

			out += fmt.Sprintf("<tr><td rowspan='%d'>%s</td>", len(diffRecs),
				fmt.Sprintf("%s %s", r.name, subNames[i]))
			for _, diffRec := range diffRecs {
				te := ""
				if len(diffRec.record) > cMaxCharsToShowInTopN {
					te = string([]rune(diffRec.record)[:cMaxCharsToShowInTopN])
				} else {
					te = diffRec.record
				}
				out += fmt.Sprintf("<td>%d</td><td>%5.2f</td><td>%s</td></tr>",
					diffRec.count, diffRec.score, te)
				if i+1 < len(diffRecs) || i+1 < ntop.n {
					out += "<tr>"
				}
			}
			out += "</tr>"
		}
	}
	out += "</table>"
	out += "</html>"
	reportPath := fmt.Sprintf("%s/results/diffs.html", rs.dataDir)
	fw, err := os.Create(reportPath)
	if err != nil {
		return err
	}
	defer fw.Close()
	if _, err := fw.WriteString(out); err != nil {
		return err
	}
	log.Println("Completed diff summary")
	return nil
}
