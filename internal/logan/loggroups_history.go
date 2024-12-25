package logan

import (
	"fmt"
	"goLogAnalyzer/pkg/utils"
	"math"
	"sort"
	"strconv"
	"time"
)

type logGroupsHistory struct {
	timeline       []int64
	counts         [][]int
	totalCounts    []int
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
	totalCounts := make([]int, len(groupIds))
	for i, groupId := range groupIds {
		gr_timeline := make([]int, len(timeline))
		for epoch, cnt := range lgs.alllg[groupId].countHistory {
			gr_timeline[timelineMap[epoch]] = cnt
		}
		counts[i] = gr_timeline
		totalCounts[i] = lgs.alllg[groupId].count
	}
	lgsh.counts = counts
	lgsh.totalCounts = totalCounts

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

	for j := range lgsh.counts[i] {
		if j == 0 {
			continue
		}
		epoch := lgsh.timeline[j]
		if minEpoch > epoch {
			continue
		}
		// Sudden disappearance
		if values[j-1] >= minOccurrences && values[j] < lowerThreshold {
			epochs = append(epochs, epoch)
		}

		// Below lower threshold anomaly
		if values[j] > upperThreshold {
			epochs = append(epochs, epoch)
		}
	}
	return
}

// Build rows from log group history
func (lgsh *logGroupsHistory) buildRows(timeFormat string, topN int) (rows [][]string) {
	rows = make([][]string, 0)
	type groupTotal struct {
		groupId    int64
		totalCount int
	}
	groupTotals := make([]groupTotal, len(lgsh.groupIds))
	for i, groupId := range lgsh.groupIds {
		groupTotals[i] = groupTotal{groupId, lgsh.totalCounts[i]}
	}

	// Sort groups by totalCount in descending order
	sort.Slice(groupTotals, func(i, j int) bool {
		return groupTotals[i].totalCount > groupTotals[j].totalCount
	})

	// Take topN groups
	if topN > len(groupTotals) {
		topN = len(groupTotals)
	}
	topGroups := groupTotals[:topN]

	for _, group := range topGroups {
		i := lgsh.groupIdsMap[group.groupId]
		for j, timestamp := range lgsh.timeline {
			count := lgsh.counts[i][j]
			// Add rows only for non-zero counts
			if count > 0 {
				epoch := time.Unix(timestamp, 0)
				rows = append(rows, []string{
					fmt.Sprint(epoch),         // time
					epoch.Format(timeFormat),  // timestr
					fmt.Sprint(group.groupId), // metric
					strconv.Itoa(count),       // value
				})
			}
		}
	}
	return rows
}

// Sum per kmeans group per epoch and build rows from log group history
func (lgsh *logGroupsHistory) buildRowsByKmeans(groupIds []int64, maxIterations, topN, trials int,
	timeFormat string) (rows [][]string, kms *kmClusters) {
	sums, kms := lgsh.sumByKmeans(groupIds, maxIterations, topN, trials)
	for i, sum := range sums {
		for j, timestamp := range lgsh.timeline {
			count := sum[j]
			// Add rows only for non-zero counts
			if count > 0 {
				epoch := time.Unix(timestamp, 0)
				if kms.clusters[i].id >= 0 {
					rows = append(rows, []string{
						fmt.Sprint(epoch),            // time
						epoch.Format(timeFormat),     // timestr
						fmt.Sprintf("cluster_%d", i), // metric
						strconv.Itoa(int(count)),     // value
					})
				} else {
					rows = append(rows, []string{
						fmt.Sprint(epoch),                       // time
						epoch.Format(timeFormat),                // timestr
						fmt.Sprint(kms.clusters[i].groupIds[0]), // metric
						strconv.Itoa(int(count)),                // value
					})
				}
			}
		}
	}
	return rows, kms
}

// plot: make plots per kmeans group. x-axis: time, y-axis: count
func (lgsh *logGroupsHistory) plotByKmeans(groupIds []int64, maxIterations, topN, trials int,
	timeFormat, title, xlabel, ylabel, output string) error {
	sums, kms := lgsh.sumByKmeans(groupIds, maxIterations, topN, trials)
	plotData := make([][]float64, len(sums))
	for i, sum := range sums {
		plotData[i] = make([]float64, len(lgsh.timeline))
		for j, count := range sum {
			plotData[i][j] = float64(count)
		}
	}
	plotLabels := make([]string, len(sums))
	for i, cluster := range kms.clusters {
		if cluster.id >= 0 {
			plotLabels[i] = fmt.Sprintf("cluster_%d", i)
		} else {
			plotLabels[i] = fmt.Sprint(cluster.groupIds[0])
		}
	}
	return utils.Plot(lgsh.timeline, plotData, plotLabels, title, xlabel, ylabel, output, timeFormat)
}

func (lgsh *logGroupsHistory) sumByKmeans(groupIds []int64,
	maxIterations, topN, trials int) ([][]int64, *kmClusters) {
	k := int(math.Ceil(float64(len(groupIds))*cKmeansKRate)) + 1
	if k < cKmeansMinK {
		k = cKmeansMinK
	}
	if k > len(groupIds) {
		k = int(math.Ceil(float64(len(groupIds))*cKmeansKRate)) + 1
	}

	counts := lgsh.counts
	ncounts := make([][]float64, len(counts))
	for i, cnt := range counts {
		ncounts[i] = utils.NormalizeInt(cnt)
	}
	clusters, centroids, sizes, score, clusterScores := utils.BestKMeans(ncounts, k, maxIterations, topN, trials)
	clusters, _ = utils.FilterOutliers(2.0, ncounts, clusters, centroids)

	kms := newKmClusters(score)
	for i, cluster := range clusters {
		if i >= len(clusterScores) {
			break
		}
		groupIds := make([]int64, 0)
		totalCount := 0
		for _, j := range cluster {
			groupIds = append(groupIds, lgsh.groupIds[j])
			totalCount += lgsh.totalCounts[j]
		}
		kms.addCluster(i, sizes[i], clusterScores[i], cluster, groupIds, counts, totalCount)
	}

	// Identify log groups that do not belong to any k-means cluster and have a size larger than the total count
	// of each cluster, then add these log groups to the sums
	for i, totalCount := range lgsh.totalCounts {
		inCluster := false
		for _, cluster := range clusters {
			for _, j := range cluster {
				if j == i {
					inCluster = true
					break
				}
			}
			if inCluster {
				break
			}
		}
		if !inCluster {
			kms.addCluster(-1, 0, 0, []int{i}, []int64{lgsh.groupIds[i]}, [][]int{counts[i]}, totalCount)
		}
	}

	// sort kmClusters by logCountTotal
	kms.sortByLogCountTotal()
	kms.filterTopN(topN)

	sums := make([][]int64, len(kms.clusters))
	for i, cluster := range kms.clusters {
		sums[i] = make([]int64, len(lgsh.timeline))
		for j, _ := range lgsh.timeline {
			for _, k := range cluster.memberIds {
				sums[i][j] += int64(counts[k][j])
			}
		}
	}

	return sums, kms
}
