package logan

import "encoding/json"

type kmCluster struct {
	id            int
	size          int
	logCountTotal int
	clusterScore  float64
	groupIds      []int64
}

type kmClusters struct {
	score    float64
	clusters []*kmCluster
}

func newKmCluster(id int, size int, clusterScore float64,
	groupIds []int64, logCountTotal int) *kmCluster {
	km := new(kmCluster)
	km.id = id
	km.size = size
	km.clusterScore = clusterScore
	km.groupIds = groupIds
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
	groupIds []int64, totalCount int) {
	kms.clusters = append(kms.clusters, newKmCluster(id, size, clusterScore, groupIds, totalCount))
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
