package utils

import (
	"testing"
)

func Test_kmeans(t *testing.T) {
	// Example dataset
	data := [][]float64{
		{1.0, 2.0}, {1.5, 1.8}, {5.0, 8.0},
		{8.0, 8.0}, {1.0, 0.6}, {9.0, 11.0},
		{8.0, 10.0}, {1.0, 3.0}, {9.5, 9.5},
	}

	// Parameters
	k := 3
	maxIterations := 100
	trials := 10

	// Run K-means
	clusters, centroids, _, _, _ := BestKMeans(data, k, maxIterations, k, trials)
	if err := GetGotExpErr("len clusters", len(clusters), 3); err != nil {
		t.Error(err)
	}

	// Filter out outliers
	filteredClusters, filteredCentroids := FilterOutliers(2, data, clusters, centroids)

	if err := GetGotExpErr("len filtered centroids", len(filteredClusters), 3); err != nil {
		t.Error(err)
	}

	if err := GetGotExpErr("len filtered centroids", len(filteredCentroids), 3); err != nil {
		t.Error(err)
	}
}
