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
)

var purple color.RGBA = color.RGBA{0x80, 0x00, 0x80, 0xFF}
var rankThreshold int = 370
var percentageOfPixelsPerCluster float64 = 0.35
var config Config = Config {5, 1000000, Euclidean, GetPointFromLargestVarianceCluster}

func RankProductBasedOnAmountOfPurpleInImage(products []*Product, product *Product, imageFile *os.File) int {
	rank, error := findAmountOfPurpleInImage(imageFile)
	if error == nil {
		//log.Printf("%v ranked at %d\n", product.Urls.Url, rank)
		return rank
	} else {
		log.Printf("Unable to find rank for %v. %v\n", product.Urls.Url, error)
		return -1
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
		return 0, errors.New("No cluster was large enough")
	}
	dominantColor := dominantColors[index]
	distanceToPurple := DistanceBetweenColors(pointToColor(*dominantColor.center), purple)

	// The distance should be as small as possible, but the rank should be as high as possible
	rank := int(MAX_DISTANCE - distanceToPurple)
	if rank < rankThreshold {
		return 0, errors.New(fmt.Sprintf("Image not purple enough, was %d needed %d", rank, rankThreshold))
	}
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
