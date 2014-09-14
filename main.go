package main

/*
   TODO: Don't buy clothes
   TODO: Don't buy books
   TODO: Don't buy apps
   TODO: Don't buy sex toys
*/
import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"image"
	"image/jpeg"
	"image/png"
	"image/color"
	"errors"
)

type Product struct {
	Urls  *ProductUrls
	Image *os.File
}

func main() {
	// TODO: Make multiple parallel calls to the findRandom thingie
	// TODO: Limit products on price and availability
	productUrls := findProductsOnRandomSearchpage()
	if len(productUrls) == 0 {
		log.Fatal("No products found!")
	}

	products := downloadImages(productUrls)
	defer cleanUp(products)
	if len(products) == 0 {
		log.Fatal("No images downloaded!")
	}

	highestRankedProduct := findHighestRankedProduct(products, 350)
	if highestRankedProduct == nil {
		fmt.Println("Did not find a good enough product :(")
	} else {
		buyProduct(highestRankedProduct)
	}
}

func downloadImages(productUrls []*ProductUrls) []*Product {
	var products []*Product
	for _, urls := range productUrls {
		imageFile, error := downloadImage(urls.ImageUrl)
		if error == nil {
			products = append(products, &Product{urls, imageFile})
		} else {
			log.Println(error)
		}
	}

	return products
}

func downloadImage(url *url.URL) (*os.File, error) {
	res, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	file, err := os.Create("image_" + getImageName(url))
	if err != nil {
		return nil, err
	}

	_, err = file.Write(data)
	if err != nil {
		return nil, err
	}

	file.Sync()
	log.Printf("Downloaded %s to %s\n", url.String(), file.Name())
	return file, nil
}

func getImageName(url *url.URL) string {
	urlString := url.String()
	slashIndex := strings.LastIndex(urlString, "/")

	return urlString[slashIndex + 1:]
}

func findHighestRankedProduct(products []*Product, rankThreshold int) *Product {
	highestRank := rankThreshold
	var highestRankedProduct *Product = nil
	for _, product := range products {
		productRank := rankProduct(products, product)
		if productRank > highestRank {
			highestRank = productRank
			highestRankedProduct = product
		}
	}

	return highestRankedProduct
}

func rankProduct(products []*Product, product *Product) int {
	// Here is where the algorithm to choose product is implemented
	rank, error := findAmountOfPurpleInImage(product.Image)
	if error == nil {
		log.Printf("Rank: %d\n", rank)
		return rank
	} else {
		log.Printf("Unable to find rank. %v\n", error)
		return -1
	}
}

func findAmountOfPurpleInImage(imageFile *os.File) (int, error) {
	log.Println("Finding amount of purple in " + imageFile.Name())
	image, error := fileToImage(imageFile)
	if error != nil {
		return 0, error
	}

	purple := color.RGBA{0x66, 0x50, 0x88, 0xFF} // #665088
	distanceToPurple := distance(image, purple)
	log.Printf("Distance to %s: %d\n", purple, distanceToPurple)

	// The distance should be as small as possible, but the rank should be as high as possible
	return int(MAX_DISTANCE - distanceToPurple), nil
}

// TODO: Doesn't really work... :(
func fileToImage(file *os.File) (image.Image, error) {
	if strings.HasSuffix(file.Name(), "jpg") || strings.HasSuffix(file.Name(), "jpeg") {
		return jpeg.Decode(file)
	} else if strings.HasSuffix(file.Name(), "png") {
		return png.Decode(file)
	}

	return nil, errors.New("I don't know the format of " + file.Name())
}

func buyProduct(product *Product) {
	fmt.Printf("I AM GONNA BUY %s\n", product)
}

func cleanUp(products []*Product) {
	for _, product := range products {
		err := os.Remove(product.Image.Name())
		if err != nil {
			log.Println(err)
		}
	}
	log.Printf("Removed %d images\n", len(products))
}
