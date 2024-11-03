package logan

type logGroupsHistory struct {
	timeline       []int64
	counts         [][]int
	groupIds       []int64
	displayStrings []string
	groupIdsMap    map[int64]int
	timelineMap    map[int64]int
}

func newLogGroupsHistory(lgs *logGroups, start, end,
	unitSecs int64, groupIds []int64) *logGroupsHistory {
	lgsh := new(logGroupsHistory)
	timeline := make([]int64, 0)
	groupIdsMap := make(map[int64]int)
	timelineMap := make(map[int64]int)

	if groupIds == nil {
		groupIds = make([]int64, 0)
	}

	i := 0
	pos := start
	for pos <= end {
		timeline = append(timeline, pos)
		timelineMap[pos] = i
		i++
		pos += unitSecs
	}
	lgsh.timeline = timeline
	lgsh.timelineMap = timelineMap

	if len(groupIds) == 0 {
		for groupId := range lgs.alllg {
			groupIds = append(groupIds, groupId)
		}
	}
	lgsh.groupIds = groupIds
	for i, groupId := range groupIds {
		groupIdsMap[groupId] = i
	}
	lgsh.groupIdsMap = groupIdsMap

	counts := make([][]int, len(groupIds))
	for i, groupId := range groupIds {
		gr_timeline := make([]int, len(timeline))
		for epoch, cnt := range lgs.alllg[groupId].countHistory {
			gr_timeline[timelineMap[epoch]] = cnt
		}
		counts[i] = gr_timeline
	}
	lgsh.counts = counts

	displayStrings := make([]string, len(lgs.alllg))
	for i, groupId := range groupIds {
		displayStrings[i] = lgs.alllg[groupId].displayString
	}
	lgsh.displayStrings = displayStrings
	return lgsh
}

func (lgsh *logGroupsHistory) getCount(groupId, epoch int64) int {
	i, ok := lgsh.groupIdsMap[groupId]
	if !ok {
		return -1
	}
	j, ok := lgsh.timelineMap[epoch]
	if !ok {
		return -1
	}
	return lgsh.counts[i][j]
}
