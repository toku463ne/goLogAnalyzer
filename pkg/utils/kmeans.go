package utils

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"gonum.org/v1/gonum/mat"
)

// EuclideanDistance calculates the distance between two points.
func EuclideanDistance(a, b []float64) float64 {
	var sum float64
	for i := range a {
		sum += math.Pow(a[i]-b[i], 2)
	}
	return math.Sqrt(sum)
}

// Mean calculates the centroid of a cluster of points.
func Mean(cluster [][]float64) []float64 {
	if len(cluster) == 0 {
		return nil
	}

	dim := len(cluster[0])
	centroid := make([]float64, dim)

	for _, point := range cluster {
		for i := 0; i < dim; i++ {
			centroid[i] += point[i]
		}
	}

	for i := range centroid {
		centroid[i] /= float64(len(cluster))
	}
	return centroid
}

// AssignClusters assigns each point to the nearest centroid.
func AssignClusters(data [][]float64, centroids [][]float64) []int {
	assignments := make([]int, len(data))

	for i, point := range data {
		minDist := math.MaxFloat64
		nearestCluster := 0

		for j, centroid := range centroids {
			if len(centroid) == 0 {
				continue
			}
			dist := EuclideanDistance(point, centroid)
			if dist < minDist {
				minDist = dist
				nearestCluster = j
			}
		}
		assignments[i] = nearestCluster
	}

	return assignments
}

// PrintClusterResults prints the clustering results.
func PrintClusterResults(data [][]float64, clusters [][]int, centroids [][]float64) {
	for i, cluster := range clusters {
		fmt.Printf("Cluster %d (Centroid: %v):\n", i, centroids[i])
		for _, idx := range cluster {
			fmt.Println(data[idx])
		}
	}
}

// IntraClusterDistance calculates the average distance between a point and all other points in the same cluster.
func IntraClusterDistance(point []float64, cluster [][]float64) float64 {
	if len(cluster) <= 1 {
		return 0.0 // Single-point clusters have no intra-cluster distance
	}

	var totalDist float64
	for _, other := range cluster {
		if len(other) == 0 {
			continue
		}
		totalDist += EuclideanDistance(point, other)
	}
	return totalDist / float64(len(cluster)-1) // Exclude self
}

// InterClusterDistance calculates the average distance between a point and all points in another cluster.
func InterClusterDistance(point []float64, otherCluster [][]float64) float64 {
	var totalDist float64
	for _, other := range otherCluster {
		if len(other) == 0 {
			continue
		}
		totalDist += EuclideanDistance(point, other)
	}
	return totalDist / float64(len(otherCluster))
}

// SilhouetteScore calculates the silhouette score for all points.
func SilhouetteScore(data [][]float64, clusters [][]int, centroids [][]float64) float64 {
	var totalScore float64
	var numPoints int

	// Build clusters as slices of points
	clusterPoints := make([][][]float64, len(clusters))
	for clusterID, indices := range clusters {
		for _, idx := range indices {
			clusterPoints[clusterID] = append(clusterPoints[clusterID], data[idx])
		}
	}

	// Calculate silhouette score for each point
	for clusterID, indices := range clusters {
		for _, idx := range indices {
			point := data[idx]

			// Calculate a(i): intra-cluster distance
			a := IntraClusterDistance(point, clusterPoints[clusterID])

			// Calculate b(i): nearest cluster distance
			b := math.MaxFloat64
			for otherClusterID, otherCluster := range clusterPoints {
				if clusterID == otherClusterID {
					continue
				}
				interDist := InterClusterDistance(point, otherCluster)
				if interDist < b {
					b = interDist
				}
			}

			// Calculate silhouette score for the point
			s := (b - a) / math.Max(a, b)
			totalScore += s
			numPoints++
		}
	}

	return totalScore / float64(numPoints) // Average silhouette score
}

func KMeans(data [][]float64, k, maxIterations int) ([][]int, [][]float64, []int) {
	rand.Seed(time.Now().UnixNano())

	// Randomly initialize centroids
	centroids := make([][]float64, k)
	for i := 0; i < k; i++ {
		centroids[i] = data[rand.Intn(len(data))]
	}

	var clusters [][]int
	var clusterSizes []int

	for iteration := 0; iteration < maxIterations; iteration++ {
		// Assign points to clusters
		assignments := AssignClusters(data, centroids)

		// Group points by cluster
		clusters = make([][]int, k)
		clusterSizes = make([]int, k)
		for i, clusterID := range assignments {
			clusters[clusterID] = append(clusters[clusterID], i)
			clusterSizes[clusterID]++
		}

		// Recalculate centroids
		newCentroids := make([][]float64, k)
		for i := 0; i < k; i++ {
			clusterPoints := [][]float64{}
			for _, idx := range clusters[i] {
				if clusters[i] == nil {
					continue
				}
				clusterPoints = append(clusterPoints, data[idx])
			}
			newCentroids[i] = Mean(clusterPoints)
		}

		// Reassign single-member clusters
		for i := range clusters {
			if len(clusters[i]) == 1 {
				// Find the nearest cluster
				minDist := math.MaxFloat64
				nearestCluster := -1
				for j := range clusters {
					if clusterSizes[j] <= 1 {
						continue
					}
					if i != j {
						point := data[clusters[i][0]]
						dist := EuclideanDistance(point, centroids[j])
						if dist < minDist {
							minDist = dist
							nearestCluster = j
						}
					}
				}
				if nearestCluster != -1 {
					// Reassign the member to the nearest cluster
					clusters[nearestCluster] = append(clusters[nearestCluster], clusters[i][0])
					// Remove the single-member cluster
					clusters[i] = nil
				}
			}
		}

		// Check for convergence
		converged := true
		for i := range centroids {
			if newCentroids[i] == nil {
				continue
			}
			if !mat.EqualApprox(mat.NewVecDense(len(centroids[i]), centroids[i]), mat.NewVecDense(len(newCentroids[i]), newCentroids[i]), 1e-6) {
				converged = false
				break
			}
		}
		centroids = newCentroids

		if converged {
			break
		}
	}

	return clusters, centroids, clusterSizes
}

// CompactnessScore calculates the average intra-cluster distance for the top N clusters.
func CompactnessScore(data [][]float64, clusters [][]int, clusterSizes []int, topN int) (float64, []float64) {
	// Sort clusters by size (descending order)
	type ClusterInfo struct {
		ID   int
		Size int
	}
	clusterInfos := make([]ClusterInfo, len(clusterSizes))
	for i, size := range clusterSizes {
		clusterInfos[i] = ClusterInfo{ID: i, Size: size}
	}
	sort.Slice(clusterInfos, func(i, j int) bool {
		return clusterInfos[i].Size > clusterInfos[j].Size
	})

	// Calculate average intra-cluster distance for top N clusters
	var totalScore float64
	var count int
	clusterScores := make([]float64, 0)
	for i := 0; i < topN && i < len(clusterInfos); i++ {
		clusterID := clusterInfos[i].ID
		if len(clusters[clusterID]) == 0 {
			clusterScores = append(clusterScores, -1)
			continue
		}
		clusterPoints := [][]float64{}
		for _, idx := range clusters[clusterID] {
			clusterPoints = append(clusterPoints, data[idx])
		}
		clusterScore := 0.0
		for _, point := range clusterPoints {
			clusterScore += IntraClusterDistance(point, clusterPoints)
		}
		clusterScore /= float64(len(clusterPoints)) // Average intra-cluster distance
		clusterScores = append(clusterScores, clusterScore)
		totalScore += clusterScore
		count++
	}

	// Average score across top N clusters
	return totalScore / float64(count), clusterScores
}

func BestKMeans(data [][]float64, k, maxIterations, topN, trials int) ([][]int, [][]float64,
	[]int, float64, []float64) {
	bestScore := math.MaxFloat64
	var bestCentroids [][]float64
	var bestClusterScores []float64
	var bestClusters [][]int
	var bestSizes []int

	for t := 0; t < trials; t++ {
		clusters, centroids, clusterSizes := KMeans(data, k, maxIterations)
		score, clusterScores := CompactnessScore(data, clusters, clusterSizes, topN)

		if score > 0 && score < bestScore {
			bestScore = score
			bestCentroids = centroids
			bestClusters = clusters
			bestSizes = clusterSizes
			bestClusterScores = clusterScores
		}
	}

	return bestClusters, bestCentroids, bestSizes, bestScore, bestClusterScores
}

// FilterOutliers removes points that are more than 2σ away from the centroid of their cluster.
func FilterOutliers(sigma float64, data [][]float64, clusters [][]int, centroids [][]float64) ([][]int, [][]float64) {
	filteredClusters := make([][]int, len(clusters))
	filteredCentroids := make([][]float64, len(centroids))

	for clusterID, indices := range clusters {
		if len(indices) == 0 {
			continue
		}

		// Compute distances of all points to the centroid
		distances := make([]float64, len(indices))
		clusterPoints := make([][]float64, len(indices))
		for i, idx := range indices {
			point := data[idx]
			distances[i] = EuclideanDistance(point, centroids[clusterID])
			clusterPoints[i] = point
		}

		// Compute mean and standard deviation of distances
		mean := MeanValue(distances)
		stdDev := StandardDeviation(distances, mean)

		// Retain only points within 2σ
		newCluster := []int{}
		for i, dist := range distances {
			if math.Abs(dist-mean) <= sigma*stdDev {
				newCluster = append(newCluster, indices[i])
			}
		}

		// Update filtered cluster and recompute centroid
		filteredClusters[clusterID] = newCluster
		if len(newCluster) > 0 {
			filteredClusterPoints := [][]float64{}
			for _, idx := range newCluster {
				filteredClusterPoints = append(filteredClusterPoints, data[idx])
			}
			filteredCentroids[clusterID] = Mean(filteredClusterPoints)
		}
	}

	return filteredClusters, filteredCentroids
}

// MeanValue calculates the mean of a slice of float64 values.
func MeanValue(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// StandardDeviation calculates the standard deviation of a slice of float64 values.
func StandardDeviation(values []float64, mean float64) float64 {
	var variance float64
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	return math.Sqrt(variance / float64(len(values)))
}
