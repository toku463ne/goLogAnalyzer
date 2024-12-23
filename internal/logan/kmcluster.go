package logan

import (
	"encoding/json"
	"sort"
)

type kmCluster struct {
	id            int
	size          int
	logCountTotal int
	clusterScore  float64
	memberIds     []int
	groupIds      []int64
	history       []int
}

type kmClusters struct {
	score    float64
	clusters []*kmCluster
}

func newKmCluster(id int, size int, clusterScore float64,
	memberIds []int, groupIds []int64, history []int, logCountTotal int) *kmCluster {
	km := new(kmCluster)
	km.id = id
	km.size = size
	km.clusterScore = clusterScore
	km.memberIds = memberIds
	km.groupIds = groupIds
	km.history = history
	km.logCountTotal = logCountTotal
	return km
}

func newKmClusters(score float64) *kmClusters {
	kms := new(kmClusters)
	kms.clusters = make([]*kmCluster, 0)
	kms.score = score
	return kms
}

func (kms *kmClusters) addCluster(id int, size int, clusterScore float64,
	memberIds []int, groupIds []int64, counts [][]int, totalCount int) {
	history := make([]int, len(counts[0]))
	if len(counts) == 1 {
		history = counts[0]
	} else {
		for _, j := range memberIds {
			for i := 0; i < len(counts[0]); i++ {
				history[i] += counts[j][i]
			}
		}
	}
	kms.clusters = append(kms.clusters, newKmCluster(id, size, clusterScore, memberIds, groupIds, history, totalCount))
}

// filter the first n clusters
func (kms *kmClusters) filterTopN(n int) {
	if n < 0 || n >= len(kms.clusters) {
		return
	}
	kms.clusters = kms.clusters[:n]
}

// sort kmClusters by logCountTotal
func (kms *kmClusters) sortByLogCountTotal() {
	// sort by logCountTotal
	sort.Slice(kms.clusters, func(i, j int) bool {
		return kms.clusters[i].logCountTotal > kms.clusters[j].logCountTotal
	})
}

// convert kmClusters to JSON string
func (kms *kmClusters) toJSON() (string, error) {
	type jsonKmCluster struct {
		ID            int     `json:"id"`
		Size          int     `json:"size"`
		LogCountTotal int     `json:"log_count_total"`
		ClusterScore  float64 `json:"cluster_score"`
		GroupIds      []int64 `json:"group_ids"`
	}

	type jsonKmClusters struct {
		Score    float64         `json:"score"`
		Clusters []jsonKmCluster `json:"clusters"`
	}

	jsonClusters := jsonKmClusters{
		Score: kms.score,
		Clusters: func() []jsonKmCluster {
			clusters := make([]jsonKmCluster, len(kms.clusters))
			for i, cluster := range kms.clusters {
				clusters[i] = jsonKmCluster{
					ID:            cluster.id,
					Size:          cluster.size,
					LogCountTotal: cluster.logCountTotal,
					ClusterScore:  cluster.clusterScore,
					GroupIds:      cluster.groupIds,
				}
			}
			return clusters
		}(),
	}

	data, err := json.MarshalIndent(jsonClusters, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
