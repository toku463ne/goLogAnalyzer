package logan

import (
	"goLogAnalyzer/pkg/utils"
)

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

func (lgsh *logGroupsHistory) detectAnomaly(groupId int64,
	stdThreshold, minOccurrences float64,
	minEpoch int64) (epochs []int64) {
	i, ok := lgsh.groupIdsMap[groupId]
	if !ok {
		return
	}
	epochs = make([]int64, 0)

	values := make([]float64, 0)
	for _, cnt := range lgsh.counts[i] {
		values = append(values, float64(cnt))
	}
	mean, stdDev := utils.CalculateStats(values)
	upperThreshold := mean + stdThreshold*stdDev
	lowerThreshold := mean - stdThreshold*stdDev

	if utils.DetectPeriodicityByThreshold(values, upperThreshold, lowerThreshold) {
		return
	}

	for j := range lgsh.counts[i][1:] {
		epoch := lgsh.timeline[j]
		if minEpoch > epoch {
			continue
		}
		// Sudden disappearance
		if values[i-1] >= minOccurrences && values[i] < minOccurrences {
			if values[i-1] > upperThreshold {
				epochs = append(epochs, epoch)
			}
		}

		// Below lower threshold anomaly
		if values[i] < lowerThreshold {
			epochs = append(epochs, epoch)
		}
	}
	return
}
