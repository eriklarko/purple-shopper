package main

import (
    "image"
    "image/color"
    "math"
    "log"
)

/*type RGB struct {
	r,g,b uint8
}
func (c *RGB) RGB() (uint8, uint8, uint8) {
	return c.r, c.g, c.b
}*/

var MAX_DISTANCE float64 = 255 * math.Sqrt(3)

func distance(image image.Image, targetColor color.RGBA) float64 {
	meanColor := findMeanColor(image)
	return distanceBetweenColors(meanColor, targetColor)
}

func findMeanColor(image image.Image) color.RGBA {
	var totalRed, totalGreen, totalBlue uint32
	b := image.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r,g,b,_ := image.At(x, y).RGBA()
			totalRed   += l(r)
			totalGreen += l(g)
			totalBlue  += l(b)
		}
	}
	numberOfPixels := uint32((b.Max.X - b.Min.X) * (b.Max.Y - b.Min.Y))
	color := color.RGBA {
		uint8(totalRed / numberOfPixels),
		uint8(totalGreen / numberOfPixels),
		uint8(totalBlue / numberOfPixels),
		0xFF,
	}

	log.Printf("Mean color  %s in %d pixels\n", color, numberOfPixels)
	return color
}

/** Used to convert two byte color values to one byte */
func l (c uint32) uint32 {
	return uint32(uint8(c))
}

func distanceBetweenColors(color1, color2 color.RGBA) float64 {
	r1, g1, b1, _ := color1.RGBA()
	r2, g2, b2, _ := color2.RGBA()

	rDiff := math.Pow(lolDiff(r1, r2), 2.0)
	gDiff := math.Pow(lolDiff(g1, g2), 2.0)
	bDiff := math.Pow(lolDiff(b1, b2), 2.0)
	sum :=  rDiff + gDiff + bDiff

	return math.Sqrt(sum);
}

func lolDiff(a, b uint32) float64 {
	return float64(l(a)) - float64(l(b))
}