package main

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
)

var purple color.RGBA = color.RGBA{0x80, 0x00, 0x80, 0xFF}
var rankThreshold int = 370
var percentageOfPixelsPerCluster float64 = 0.35
var config Config = Config {5, 1000000, Euclidean, GetPointFromLargestVarianceCluster}
var rankLogFile string = fmt.Sprintf("ranks-%s.txt", time.Now().Format("2006-01-02 15:04"))

func RankProductBasedOnAmountOfPurpleInImage(product *Product, c chan<- *RankedProduct) {
	imageFile, error := os.Open(product.Image)
	defer cleanUpFile(imageFile)

	if error != nil {
		log.Printf("Unable to find rank for %v. %v\n", product.Urls.Url, error)
		return
	}

	rank, error := findAmountOfPurpleInImage(imageFile)
	if error == nil {
		logRankInfo(product, rank)

		if rank >= rankThreshold {
			c <- &RankedProduct{product, rank}
		}
	} else if !strings.Contains(error.Error(), "few datapoints") {
		log.Printf("Unable to find rank for %v. %v\n", product.Urls.Url, error)
	}
}

func cleanUpFile(file *os.File) {
	if file == nil {
		log.Println("Tried to cleanup nil file")
	} else {

		err := os.Remove(file.Name())
		if err != nil {
			log.Printf("Unable to remove file %v: %v", file, err)
		}

		err = file.Close()
		if err != nil {
			log.Printf("Unable to close file %v: %v", file, err)
		}
	}
}

func findAmountOfPurpleInImage(imageFile *os.File) (int, error) {
	image, error := fileToImage(imageFile)
	if error != nil {
		return 0, error
	}

	points := pixelsToPoints(image)
	dominantColors, error := FindClusters(config, points)
	if error != nil {
		return 0, error
	}

	clusterQualities := CalculateClusterQuality(config, dominantColors)
	index, _ := LaMaximum(clusterQualities, int(float64(len(points)) * percentageOfPixelsPerCluster))
	if index < 0 {
		return -1, nil //errors.New("No cluster was large enough")
	}
	dominantColor := dominantColors[index]
	distanceToPurple := DistanceBetweenColors(pointToColor(*dominantColor.center), purple)

	// The distance should be as small as possible, but the rank should be as high as possible
	rank := int(MAX_DISTANCE - distanceToPurple)
	return rank, nil
}

func fileToImage(file *os.File) (image.Image, error) {
	if strings.HasSuffix(file.Name(), "jpg") || strings.HasSuffix(file.Name(), "jpeg") {
		return jpeg.Decode(file)
	} else if strings.HasSuffix(file.Name(), "png") {
		return png.Decode(file)
	}

	return nil, errors.New("I don't know the format of " + file.Name())
}

func pixelsToPoints(image image.Image) []*Point {
	var colors []*Point

	b := image.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r32,g32,b32,_ := image.At(x, y).RGBA()
			r8, g8, b8 := uint8(r32), uint8(g32), uint8(b32)

			// Ignore grayscale pixels
			if math.Abs(float64(r8 - g8)) < 10 && math.Abs(float64(r8 - b8)) < 10 {
				continue
			}

			point := Point{
				float64(r8),
				float64(g8),
				float64(b8),
			}
			colors = append(colors, &point)
		}
	}

	return colors
}

func pointToColor(point Point) color.RGBA {
	color := color.RGBA{
		uint8(point[0]),
		uint8(point[1]),
		uint8(point[2]),
		0xFF,
	}
	return color
}

func logRankInfo(product *Product, rank int) {

	var file *os.File;
	var err error;

	if _, err := os.Stat(rankLogFile); err == nil {
		file, err = os.OpenFile(rankLogFile, os.O_APPEND|os.O_WRONLY, 0600)
	} else {
		file, err = os.Create(rankLogFile)
	}

	defer file.Close()
	if err != nil {
		log.Printf("Unable to persist rank. %v\n", err)
	}
	_, err = file.WriteString(fmt.Sprintf("%v ranked at %d\n", product.Urls.Url, rank))
	if err != nil {
		log.Printf("Unable to persist rank. %v\n", err)
	}
}