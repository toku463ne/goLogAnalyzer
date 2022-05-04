package analyzer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
)

func newLogConfRoot(jsonFile string) (*LogConfRoot, error) {
	bytes, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}
	var lcr *LogConfRoot
	if err := json.Unmarshal(bytes, &lcr); err != nil {
		return nil, err
	}

	if lcr.ReportDir == "" {
		return nil, errors.New("ReportDir is mandatory")
	}

	for i, child := range lcr.Children {
		lcr.Children[i], err = child.inheritConf(lcr.LogConf, lcr.RootDir, lcr.ReportDir, lcr.Templates, false)
		if err != nil {
			return nil, err
		}
	}

	return lcr, nil
}

func (node *LogNode) inheritConf(parentConf *LogConf, parentDataDir, parentReportDir string,
	templates map[string]*LogConf, isCategory bool) (*LogNode, error) {
	var template *LogConf

	if node.Name == "" {
		return nil, errors.New("Name is mandatory")
	}

	if isCategory {
		if len(node.Children) > 0 {
			return nil, errors.New("Cannot define children inside category")
		}
		node.dataDir = parentDataDir
	} else {
		node.dataDir = fmt.Sprintf("%s/%s", parentDataDir, node.Name)
	}
	node.reportDir = fmt.Sprintf("%s/%s", parentReportDir, node.Name)

	if node.TemplateName != "" {
		template = templates[node.TemplateName]
	}

	node2 := new(LogNode)
	node2.LogConf = new(LogConf)

	for _, c := range []*LogConf{parentConf, template, node.LogConf} {
		if c == nil {
			continue
		}
		if c.LogPath != "" {
			node2.LogPath = c.LogPath
		}
		if c.TopN != 0 {
			node2.TopN = c.TopN
		}
		if c.ScoreStyle != 0 {
			node2.ScoreStyle = c.ScoreStyle
		}
		if c.Search != "" {
			node2.Search = c.Search
		}
		if c.Exclude != "" {
			node2.Exclude = c.Exclude
		}
		if c.BlockSize != 0 {
			node2.BlockSize = c.BlockSize
		}
		if c.MaxBlocks != 0 {
			node2.MaxBlocks = c.MaxBlocks
		}
		if c.MaxItemBlocks != 0 {
			node2.MaxItemBlocks = c.MaxItemBlocks
		}
		if c.MinGapToRecord != 0 {
			node2.MinGapToRecord = c.MinGapToRecord
		}
		if c.DatetimeStartPos != 0 {
			node2.DatetimeStartPos = c.DatetimeStartPos
		}
		if c.DatetimeLayout != "" {
			node2.DatetimeLayout = c.DatetimeLayout
		}
		if c.KeyEmphasize != nil {
			node2.KeyEmphasize = c.KeyEmphasize
		}
		if c.ModeblockPerFile != 0 {
			node2.ModeblockPerFile = c.ModeblockPerFile
		}
		if c.MaxScore != 0 {
			node2.MaxScore = c.MaxScore
		}
		if c.MinScore != 0 {
			node2.MinScore = c.MinScore
		}
		if c.FromDate != "" {
			node2.FromDate = c.FromDate
		}
		if c.ToDate != "" {
			node2.ToDate = c.ToDate
		}
		if c.NRareTerms == 0 {
			node2.NRareTerms = CDefaultNRareTerms
		}

	}
	node.LogConf = node2.LogConf

	if (node.FromDate != "" || node.ToDate != "") && node.DatetimeLayout == "" {
		return nil, errors.WithStack(errors.New("dateLayout is mandatory if you have FromDate or ToDate"))
	}

	var err error
	for i, child := range node.Children {
		node.Children[i], err = child.inheritConf(node.LogConf,
			node.dataDir, node.reportDir, templates, false)
		if err != nil {
			return nil, err
		}
	}
	for i, category := range node.Categories {
		node.Categories[i], err = category.inheritConf(node.LogConf,
			node.dataDir, node.reportDir, templates, true)
		if err != nil {
			return nil, err
		}
	}
	if (node.Children == nil || len(node.Children) == 0) && (node.Categories == nil || len(node.Categories) == 0) {
		node.isEnd = true
	} else {
		node.isEnd = false
	}

	node.isCategory = isCategory

	return node, nil
}
