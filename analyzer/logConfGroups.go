package analyzer

func newLogConfGroups(lcr *LogConfRoot) *logConfGroups {
	lcg := new(logConfGroups)
	lcg.g = make(map[string][]*LogNode)
	for _, child := range lcr.Children {
		lcg.addValues(child)
	}
	lcg.reportDir = lcr.ReportDir
	return lcg
}

func (lcg *logConfGroups) addValues(node *LogNode) {
	var g []string
	if node.isEnd {
		if len(node.GroupNames) == 0 {
			g = []string{cDefaultGroupName}
		} else {
			g = node.GroupNames
		}
		for _, name := range g {
			lcg.g[name] = append(lcg.g[name], node)
		}
	} else {
		for _, child := range node.Children {
			lcg.addValues(child)
		}
		for _, category := range node.Categories {
			lcg.addValues(category)
		}

	}
}
