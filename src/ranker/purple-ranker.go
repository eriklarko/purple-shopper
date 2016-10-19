package ranker

import (
	"image/color"
	"math"
	"image"
	"errors"
	"image/png"
	"strings"
	"image/jpeg"
	"os"
	"log"
	"fmt"
	"time"
	"products"
	"coloralgorithms"
	"github.com/cenkalti/dominantcolor"
)

var purple color.RGBA = color.RGBA{0x80, 0x00, 0x80, 0xFF}
var rankThreshold int = 400
var percentageOfPixelsPerCluster float64 = 0.35
var config coloralgorithms.Config = coloralgorithms.Config {5, 1000000, coloralgorithms.Euclidean, coloralgorithms.GetPointFromLargestVarianceCluster}
var rankLogFile string = fmt.Sprintf("ranks/ranks-%s.txt", time.Now().Format("2006-01-02 15:04"))

func RankProductBasedOnAmountOfPurpleInImage(product *products.Product) *products.RankedProduct {
	imageFile, error := os.Open(product.Image)
	defer cleanUpFile(imageFile)

	if error != nil {
		log.Printf("Unable to find rank for %v. %v\n", product.Urls.Url, error)
	}

	rank, error := findAmountOfPurpleInImage(imageFile)
	if error == nil {
		//log.Printf("%s ranked at %d\n", product.Urls.Url, rank)
		logRankInfo(product, rank)

		if rank >= rankThreshold {
			return &products.RankedProduct{product, rank}
		}
	} else if !strings.Contains(error.Error(), "few datapoints") {
		log.Printf("Unable to find rank for %v. %v\n", product.Urls.Url, error)
	}

	return nil
}

func cleanUpFile(file *os.File) {
	if file == nil {
		log.Println("Tried to cleanup nil file")
	} else {

		error := os.Remove(file.Name())
		if error != nil {
			log.Printf("Unable to remove file %v: %v", file, error)
		}

		error = file.Close()
		if error != nil {
			log.Printf("Unable to close file %v: %v", file, error)
		}
	}
}

func findAmountOfPurpleInImage(imageFile *os.File) (int, error) {
	img, _, err := image.Decode(imageFile)
	if err != nil {
		return -1, err
	}
	dominantColor := dominantcolor.Find(img)
	distanceToPurple := coloralgorithms.DistanceBetweenColors(dominantColor, purple)

	// The distance should be as small as possible, but the rank should be as high as possible
	rank := int(coloralgorithms.MAX_DISTANCE - distanceToPurple)
	return rank, nil
}
func findAmountOfPurpleInImagee(imageFile *os.File) (int, error) {
	image, error := fileToImage(imageFile)
	if error != nil {
		return 0, error
	}

	points := pixelsToPoints(image)
	dominantColors, error := coloralgorithms.FindClusters(config, points)
	if error != nil {
		return 0, error
	}

	clusterQualities := coloralgorithms.CalculateClusterQuality(config, dominantColors)
	index, _ := coloralgorithms.LaMaximum(clusterQualities, int(float64(len(points)) * percentageOfPixelsPerCluster))
	if index < 0 {
		return -1, nil //errors.New("No cluster was large enough")
	}
	dominantColor := dominantColors[index]
	distanceToPurple := coloralgorithms.DistanceBetweenColors(pointToColor(*dominantColor.Center), purple)

	// The distance should be as small as possible, but the rank should be as high as possible
	rank := int(coloralgorithms.MAX_DISTANCE - distanceToPurple)
	return rank, nil
}

func fileToImage(file *os.File) (image.Image, error) {
	if file == nil {
		return nil, errors.New("Tried to decode nil image")
	}
	
	if strings.HasSuffix(file.Name(), "jpg") || strings.HasSuffix(file.Name(), "jpeg") {
		return jpeg.Decode(file)
	} else if strings.HasSuffix(file.Name(), "png") {
		return png.Decode(file)
	}

	return nil, errors.New("I don't know the format of " + file.Name())
}

func pixelsToPoints(image image.Image) []*coloralgorithms.Point {
	var colors []*coloralgorithms.Point

	b := image.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r32,g32,b32,_ := image.At(x, y).RGBA()
			r8, g8, b8 := uint8(r32), uint8(g32), uint8(b32)

			// Ignore grayscale pixels
			if math.Abs(float64(r8 - g8)) < 10 && math.Abs(float64(r8 - b8)) < 10 {
				continue
			}

			point := coloralgorithms.Point{
				float64(r8),
				float64(g8),
				float64(b8),
			}
			colors = append(colors, &point)
		}
	}

	return colors
}

func pointToColor(point coloralgorithms.Point) color.RGBA {
	color := color.RGBA{
		uint8(point[0]),
		uint8(point[1]),
		uint8(point[2]),
		0xFF,
	}
	return color
}

func logRankInfo(product *products.Product, rank int) {

	var file *os.File;
	var error error;

	if _, error := os.Stat(rankLogFile); error== nil {
		file, error = os.OpenFile(rankLogFile, os.O_APPEND|os.O_WRONLY, 0600)
	} else {
		file, error = os.Create(rankLogFile)
	}

	defer file.Close()
	if error != nil {
		log.Printf("Unable to persist rank. Could not stat %s. %v\n", rankLogFile, error)
	}
	_, error = file.WriteString(fmt.Sprintf("%v ranked at %d\n", product.Urls.Url, rank))
	if error != nil {
		log.Printf("Unable to persist rank. Could not write to %s. %v\n", rankLogFile, error)
	}
}
