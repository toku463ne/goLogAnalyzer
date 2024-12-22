package logan

import (
	"fmt"
	"goLogAnalyzer/pkg/utils"
	"math"
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
func (lgsh *logGroupsHistory) buildRows(timeFormat string) (rows [][]string) {
	rows = make([][]string, 0)
	for i, groupId := range lgsh.groupIds {
		for j, timestamp := range lgsh.timeline {
			count := lgsh.counts[i][j]
			// Add rows only for non-zero counts
			if count > 0 {
				rows = append(rows, []string{
					time.Unix(timestamp, 0).Format(timeFormat), // time
					fmt.Sprint(groupId),                        // metric
					strconv.Itoa(count),                        // value
				})
			}
		}
	}
	return rows
}

// Sum per kmeans group per epoch and build rows from log group history
func (lgsh *logGroupsHistory) buildRowsByKmeans(groupIds []int64, maxIterations, topN, trials int,
	timeFormat string) (rows [][]string, kms *kmClusters) {
	sums, kms, nonClusterGroupIds := lgsh.sumByKmeans(groupIds, maxIterations, topN, trials)
	lastKmId := len(kms.clusters) - 1
	nonClusterI := 0
	for i, sum := range sums {
		for j, timestamp := range lgsh.timeline {
			count := sum[j]
			// Add rows only for non-zero counts
			if count > 0 {
				if i <= lastKmId {
					rows = append(rows, []string{
						time.Unix(timestamp, 0).Format(timeFormat), // time
						fmt.Sprintf("cluster_%d", i),               // metric
						strconv.Itoa(int(count)),                   // value
					})
				} else {
					rows = append(rows, []string{
						time.Unix(timestamp, 0).Format(timeFormat),  // time
						fmt.Sprint(nonClusterGroupIds[nonClusterI]), // metric
						strconv.Itoa(int(count)),                    // value
					})
					nonClusterI++
				}
			}
		}
	}
	return rows, kms
}

func (lgsh *logGroupsHistory) sumByKmeans(groupIds []int64,
	maxIterations, topN, trials int) ([][]int64, *kmClusters, []int64) {
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
		groupIds := make([]int64, 0)
		totalCount := 0
		for _, j := range cluster {
			groupIds = append(groupIds, lgsh.groupIds[j])
			totalCount += lgsh.totalCounts[j]
		}
		kms.addCluster(i, sizes[i], clusterScores[i], groupIds, totalCount)
	}

	// sum up the counts per cluster per epoch
	sums := make([][]int64, len(clusters))
	for i, cluster := range clusters {
		sums[i] = make([]int64, len(counts[0]))
		for _, j := range cluster {
			for l, cnt := range counts[j] {
				sums[i][l] += int64(cnt)
			}
		}
	}

	nonClusterGroupIds := make([]int64, 0)

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
			for _, size := range sizes {
				if totalCount > size {
					newRow := make([]int64, len(counts[i]))
					for l, cnt := range counts[i] {
						newRow[l] = int64(cnt)
					}
					sums = append(sums, newRow)
					break
				}
			}
			nonClusterGroupIds = append(nonClusterGroupIds, lgsh.groupIds[i])
		}
	}

	return sums, kms, nonClusterGroupIds
}
