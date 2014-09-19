package main

/**
	Shamelessly stolen from Apache commons math
 */

import (
	"errors"
	"time"
	"math"
	"math/rand"
)

// Represent a data point. The length of the array specifies the dimensions of the point
type Point []float64

// Just a cleaner name for a set of points
type Cluster []*Point

// A holder for a cluster and its center
type CentroidCluster struct {
	center *Point
	points *Cluster
}

// function used to calculate the distance between two points
type DistanceMeasure func(a, b *Point) (float64, error)
func Euclidean(a, b *Point)  (float64, error) {
	if len(*a) != len(*b) {
		return 0, errors.New("Cannot calculate distance between points of different dimensions")
	}

	acc := float64(0)
	for i,_ := range *a {
		acc += math.Pow((*a)[i] - (*b)[i], 2.0)
	}

	return math.Sqrt(acc), nil
}

// function to determine what to do with empty clusters that might appear in the algorithm
type EmptyClusterStrategy func(Config, []*CentroidCluster) *Point
func GetPointFromLargestVarianceCluster(config Config, clusters []*CentroidCluster) *Point {
	maxVariance := 0.0
	var selectedCluster *CentroidCluster = nil
	for _, cluster := range clusters {
		if len(*cluster.points) > 0 {
			variance := PopulationVariance(calculateDistances(config, cluster.points, cluster.center))
			if variance > maxVariance {
				maxVariance = variance
				selectedCluster = cluster
			}
		}
	}

	if selectedCluster == nil {
		for _, cluster := range clusters {
			if len(*cluster.points) > 0 {
				selectedCluster = cluster
				break
			}
		}
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := r.Intn(len(*selectedCluster.points))
	selectedPoint := (*selectedCluster.points)[randomIndex]
	return selectedPoint
}

func calculateDistances(config Config, cluster *Cluster, point *Point) []float64 {
	var distances []float64
	for _, pointA := range *cluster {
		// TODO: Handle error
		distance, _ := config.measure(pointA, point)
		distances = append(distances, distance)
	}
	return distances
}

type Config struct {
	k int /* number of clusters */
	maxIterations int32 /* a negative value means no max */
    measure DistanceMeasure
	emptyStrategy EmptyClusterStrategy
}

var DEFAULT_CONFIG Config = Config {-1, math.MaxInt32, Euclidean, GetPointFromLargestVarianceCluster}

func FindClusters(config Config, points []*Point) ([]*CentroidCluster, error) {
	if len(points) < config.k {
		return nil, errors.New("Too few datapoints")
	}

	clusters := chooseInitialCenters(config, points);
	var assignments []int
	for i := 0; i < len(points); i++ {
		assignments = append(assignments, 0)
	}
	assignPointsToClusters(config, clusters, points, assignments)

	var maxIterations int32 = config.maxIterations
	for i := int32(0); i < maxIterations; i++ {

		// Update cluster centers
		emptyCluster := false
		var newClusters []*CentroidCluster
		for _, cluster := range clusters {
			var newCenter *Point
			if len(*cluster.points) == 0 {
				emptyCluster = true
				newCenter = config.emptyStrategy(config, clusters)
			} else {
				newCenter = centroidOf(cluster.points, cluster.center, len(*cluster.center))
			}

			newClusters = append(newClusters, &CentroidCluster{newCenter, &Cluster{}})
		}

		pointsThatChangedCluster := assignPointsToClusters(config, newClusters, points, assignments)
		clusters = newClusters

		if pointsThatChangedCluster == 0 && !emptyCluster {
			return clusters, nil
		}
	}
	return clusters, nil
}

func chooseInitialCenters(config Config, points []*Point) []*CentroidCluster {
	numPoints := len(points)
	var taken []bool
	for i := 0; i < numPoints; i++ {
		taken = append(taken, false)
	}

	var resultSet []*CentroidCluster

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	firstPointIndex := random.Intn(numPoints)
	firstPoint := points[firstPointIndex]
	resultSet = append(resultSet, &CentroidCluster{firstPoint, &Cluster{}})
	taken[firstPointIndex] = true

	var minDistSquared []float64
	for i := 0; i < numPoints; i++ {
		minDistSquared = append(minDistSquared, 0.0)
	}

	for i := 0; i < numPoints; i++ {
		if i != firstPointIndex {
			// TODO: Handle error
			d, _ := config.measure(firstPoint, points[i])
			minDistSquared[i] = d * d
		}
	}

	for len(resultSet) < config.k {
		distSqSum := float64(0)
		for i := 0; i < numPoints; i++ {
			if !taken[i] {
				distSqSum += minDistSquared[i]
			}
		}

		r := random.NormFloat64() * distSqSum
		nextPointIndex := -1
		sum := 0.0
		for i := 0; i < numPoints; i++ {
			if !taken[i] {
				sum += minDistSquared[i]
				if sum >= r {
					nextPointIndex = i
					break
				}
			}
		}

		if nextPointIndex == -1 {
			for i := numPoints - 1; i >= 0; i-- {
				if !taken[i] {
					nextPointIndex = i;
					break;
				}
			}
		}

		if nextPointIndex >= 0 {
			nextPoint := points[nextPointIndex]
			resultSet = append(resultSet, &CentroidCluster{nextPoint, &Cluster{}})
			taken[nextPointIndex] = true

			if len(resultSet) < config.k {
				for j := 0; j < numPoints; j++ {
					if !taken[j] {
						// TODO: Handle error
						d, _ := config.measure(nextPoint, points[j])
						dSqr := d * d
						if dSqr < minDistSquared[j] {
							minDistSquared[j] = dSqr
						}
					}
				}
			}
		} else {
			break
		}
	}
	return resultSet
}

func centroidOf(cluster *Cluster, center *Point, dimensions int) *Point {
	var centroid Point
	for i := 0; i < dimensions; i++ {
		centroid = append(centroid, 0.0)
	}

	for _, point := range *cluster {
		for i := 0; i < dimensions; i++ {
			centroid[i] += (*point)[i]
		}
	}

	for i := 0; i < len(centroid); i++ {
		centroid[i] = centroid[i] / float64(len(*cluster))
	}

	return &centroid
}

func assignPointsToClusters(config Config, clusters []*CentroidCluster, points []*Point, assignments []int) int {
	assignedDifferently := 0

	for pointIndex,point := range points {
		clusterIndex := getNearestCluster(config, clusters, point)
		if clusterIndex != assignments[pointIndex] {
			assignedDifferently++
		}
		*clusters[clusterIndex].points = append(*clusters[clusterIndex].points, point)
		assignments[pointIndex] = clusterIndex
	}

	return assignedDifferently
}

func getNearestCluster(config Config, clusters []*CentroidCluster, point *Point) int {
	minDistance := math.MaxFloat64
	minCluster := 0

	for clusterIndex, cluster := range clusters {
		// TODO: Handle error
		distance, _ := config.measure(point, cluster.center)
		if distance < minDistance {
			minDistance = distance
			minCluster = clusterIndex
		}
	}

	return minCluster
}


type Quality struct {
	rootSquaredError float64
	dataPoints int
}

func CalculateClusterQuality(config Config, clusters []*CentroidCluster) []Quality {
	var qualities []Quality
	for _, cluster := range clusters {
		clusterQuality := 0.0
		for _, point := range *cluster.points {
			// TODO: Handle error
			distance,_ := config.measure(point, cluster.center)
			clusterQuality += math.Sqrt(distance * distance)
		}
		clusterQuality /= float64(len(*cluster.points))
		qualities = append(qualities, Quality{clusterQuality, len(*cluster.points)})
	}

	return qualities
}
