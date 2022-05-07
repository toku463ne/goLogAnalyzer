package analyzer

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func newReport(jsonFile string, nDays int) (*report, error) {
	r := new(report)
	var err error
	r.conf, err = newLogConfRoot(jsonFile)
	if err != nil {
		return nil, err
	}
	InitLog(r.conf.RootDir)
	r.confGroups = newLogConfGroups(r.conf)
	if nDays == 0 {
		nDays = CDefaultDaysToReport
	}
	r.daysToShow = nDays
	return r, nil
}

func (r *report) getAnalyzer(node *LogNode) (*rarityAnalyzer, error) {
	var err error
	ac := NewAnalConf(node.dataDir)
	ac.LogPathRegex = node.LogPath
	ac.BlockSize = node.BlockSize
	ac.MaxBlocks = node.MaxBlocks
	ac.MaxItemBlocks = node.MaxItemBlocks
	ac.DatetimeStartPos = node.DatetimeStartPos
	ac.DatetimeLayout = node.DatetimeLayout
	ac.ScoreStyle = node.ScoreStyle
	ac.ScoreNSize = node.ScoreNSize
	ac.MinGapToRecord = node.MinGapToRecord
	ac.NTopRecordsCount = node.TopN
	ac.NRareTerms = node.NRareTerms
	if node.ModeblockPerFile == cIntTrue {
		ac.ModeblockPerFile = true
	} else {
		ac.ModeblockPerFile = false
	}

	a, err := newRarityAnalyzer(ac)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *report) getStartEndEpoch(node *LogNode) (int64, int64, error) {
	end := int64(0)
	start := int64(0)
	var err error
	if node.FromDate != "" {
		if start, err = Str2Epoch(node.DatetimeLayout, node.FromDate); err != nil {
			return 0, 0, err
		}
	}
	if node.ToDate != "" {
		if end, err = Str2Epoch(node.DatetimeLayout, node.ToDate); err != nil {
			return 0, 0, err
		}
	}
	if node.FromDate == "" && node.ToDate == "" && node.DatetimeLayout != "" {
		end = getCurrentEpoch()
		start = end - 3600*24*CDefaultDaysToReport
	}
	return start, end, nil
}

func (r *report) insertHtmlTag(te string, emp map[string][]string) string {
	for k, l := range emp {
		for _, v := range l {
			a := strings.Split(v, " ")
			re := regexp.MustCompile(fmt.Sprintf("(?i)%s", k))
			matches := re.FindAllString(te, -1)
			matches = UniqueStringSplit(matches)
			for _, s := range matches {
				te = strings.ReplaceAll(te, s, fmt.Sprintf("<%s>%s</%s>", v, s, a[0]))
			}
		}
	}
	return te
}

func (r *report) createDetailedReport(node *LogNode,
	stats *stats, n int,
	keyRareTerms map[string][]string,
	records []*colLogRecord) error {
	out := "<html>"

	// count per stats
	ost, _, err := stats.getCountPerStatsHtml(0)
	if err != nil {
		return err
	}
	out += ost
	out += "<br>"

	out += fmt.Sprintf("<h3>Top %d rare %s records</h3><br>", node.TopN, node.Name)
	out += "<table border=1 ~~~ style='table-layout:fixed;width:100%;'>"
	out += "<tr><td width=4%>count</td><td width=10%>lastUpdate</td><td width=6%>score</td><td width=10%>rowID</td><td>text</td></tr>"
	topScore := 0.0
	for i, logr := range records {
		if logr == nil {
			break
		}
		if topScore == 0 {
			topScore = logr.score
		}
		te := ""
		if len(logr.record) > cMaxCharsToShowInTopN {
			te = string([]rune(logr.record)[:cMaxCharsToShowInTopN])
		} else {
			te = logr.record
		}

		te = r.insertHtmlTag(te, keyRareTerms)
		te = r.insertHtmlTag(te, node.KeyEmphasize)

		out += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%8.2f</td><td>%10d</td><td>%s</td></tr>",
			logr.count, logr.lastDate, logr.score, logr.rowid, te)

		if logr.score == 0 {
			break
		}

		if i+1 >= n {
			break
		}
	}
	out += "</table>"
	out += "</html>"
	if err := ensureDir(node.reportDir); err != nil {
		return err
	}
	filePath := fmt.Sprintf("%s/%s.html", node.reportDir, node.Name)
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

func (r *report) run() error {
	log.Printf("Creating reports")
	done := make(map[string]bool, 0)
	for groupName, g := range r.confGroups.g {
		out := "<html>"
		out += fmt.Sprintf("<head><title>%s</title></head>", groupName)
		out += "<body>"
		out += fmt.Sprintf("<h2>group %s digest</h2><br><br>", groupName)
		out += "<table border=1 ~~~ style='table-layout:fixed;width:100%;'>"
		out += "<tr><td width=10%>name</td>"
		out += "<td width=4%>count</td>"
		out += "<td>text</td></tr>"

		for _, node := range g {
			a, err := r.getAnalyzer(node)
			if err != nil {
				log.Printf("%+v", err)
				a.close()
				continue
			}
			if !done[node.Name] && node.LogPath != "" && (len(node.Categories) > 0 || node.isEnd) {
				log.Printf("[%s] blockSize=%d maxBlocks=%d maxItemBlocks=%d minGap=%1.1f",
					node.Name, a.BlockSize, a.MaxBlocks, a.MaxItemBlocks, a.MinGapToRecord)
				err = a.analyze(0)
				if err != nil {
					return err
				}
				done[node.Name] = true
				if len(node.Categories) > 0 {
					for _, cat := range node.Categories {
						done[cat.Name] = true
					}
				}
			}
			start, end, err := r.getStartEndEpoch(node)
			if err != nil {
				return err
			}
			ntop, err := a.getNTop(node.Name, node.TopN, start, end,
				node.Search, node.Exclude, node.MinScore, node.MaxScore, node.NRareTerms)
			if err != nil {
				return err
			}
			if ntop.getLen() == 0 {
				continue
			}
			records := ntop.getRecords2()
			keyRareTerms := make(map[string][]string, 0)
			rareTerms := ntop.getRareTerms()
			for _, term := range rareTerms {
				if term == "" {
					continue
				}
				keyRareTerms[term] = []string{cHtmlRareEmphTag}
			}
			out += fmt.Sprintf("<tr><td rowspan='%d'>%s</td>", len(records), node.Name)
			for _, rec := range records {
				if rec == nil {
					break
				}
				if rec.count > cMaxCountToShowInDigest {
					break
				}

				out += fmt.Sprintf("<td>%d</td>", rec.count)
				txt := rec.record
				txt = r.insertHtmlTag(txt, keyRareTerms)
				txt = r.insertHtmlTag(txt, node.KeyEmphasize)
				out += fmt.Sprintf("<td>%s</td>", txt)
				out += "</tr>"
			}

			if err := r.createDetailedReport(node,
				a.stats, ntop.n, keyRareTerms, records); err != nil {
				log.Printf("%+v", err)
			}

			a.close()
			a = nil
		}

		out += "</table>"
		out += "</body></html>"
		if err := ensureDir(r.confGroups.reportDir); err != nil {
			return err
		}
		reportPath := fmt.Sprintf("%s/%s.html", r.confGroups.reportDir, groupName)
		fw, err := os.Create(reportPath)
		if err != nil {
			return err
		}
		defer fw.Close()
		if _, err := fw.WriteString(out); err != nil {
			return err
		}
	}

	return nil
}
