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

func (r *report) analyzeAndCreateReport(node *LogNode) error {
	a, err := r.getAnalyzer(node)
	if err != nil {
		return err
	}

	if !node.isCategory && node.LogPath != "" && (len(node.Categories) > 0 || node.isEnd) {
		log.Printf("[%s] blockSize=%d maxBlocks=%d maxItemBlocks=%d minGap=%1.1f",
			node.Name, a.BlockSize, a.MaxBlocks, a.MaxItemBlocks, a.MinGapToRecord)

		err = a.analyze(0)
		if err != nil {
			return err
		}
	}
	if node.isEnd {
		start, end, err := r.getStartEndEpoch(node)
		if err != nil {
			return err
		}
		ar, err := a.getNTop(node.Name, node.TopN, start, end,
			node.Search, node.Exclude, node.MinScore, node.MaxScore)
		if err != nil {
			return err
		}
		if ar.getLen() == 0 {
			return nil
		}
		out := "<html>"
		out += fmt.Sprintf("<h2>Top %d rare %s records</h2><br><br>", node.TopN, node.Name)
		tmp, _, err := ar.getHtmlTable(node.TopN)
		if err != nil {
			return err
		}
		out += tmp
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
	} else {
		for _, child := range node.Children {
			if err := r.analyzeAndCreateReport(child); err != nil {
				return err
			}
		}
		for _, cat := range node.Categories {
			if err := r.analyzeAndCreateReport(cat); err != nil {
				return err
			}
		}
	}
	return nil
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

func (r *report) createDigestReport() error {
	for groupName, g := range r.confGroups.g {
		out := "<html>"
		out += fmt.Sprintf("<head><title>%s</title></head>", groupName)
		out += "<body>"
		out += fmt.Sprintf("<h2>group %s digest</h2><br><br>", groupName)
		out += "<table border=1 ~~~ style='table-layout:fixed;width:100%;'>"
		out += "<tr><td width=10%>name</td>"
		out += "<td width=4%>count</td>"
		//out += "<td width=6%>score</td>"
		//out += "<td width=10%>lastUpdate</td>"
		out += "<td>text</td></tr>"

		for _, node := range g {
			a, err := r.getAnalyzer(node)
			if err != nil {
				return err
			}
			start, end, err := r.getStartEndEpoch(node)
			if err != nil {
				return err
			}

			ar, err := a.getNTop(node.Name, node.TopN, start, end,
				node.Search, node.Exclude, node.MinScore, node.MaxScore)
			if err != nil {
				return err
			}
			if ar.getLen() == 0 {
				continue
			}
			records := ar.getRecords2()
			keyRareTerms := make(map[string][]string, 0)
			for _, term := range ar.getRareTerms(a.NRareTerms, records) {
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

				out += fmt.Sprintf("<td>%d</td>", rec.count)
				//out += fmt.Sprintf("<td>%5.2f</td>", rec.score)
				//out += fmt.Sprintf("<td>%s</td>", rec.lastDate)
				txt := rec.record
				txt = r.insertHtmlTag(txt, keyRareTerms)
				txt = r.insertHtmlTag(txt, node.KeyEmphasize)
				out += fmt.Sprintf("<td>%s</td>", txt)
				out += "</tr>"
			}
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

func (r *report) run() error {
	// analyze and save report per logs
	for _, child := range r.conf.Children {
		if child.Name != "" {
			log.Printf("Processing %s", child.Name)
		}
		if err := r.analyzeAndCreateReport(child); err != nil {
			return err
		}
	}

	// create digests
	log.Printf("Creating reports")
	if err := r.createDigestReport(); err != nil {
		return err
	}
	return nil
}
