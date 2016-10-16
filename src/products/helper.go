package products

import (
	"net/url"
	"log"
)

func ToOneProductUrls(urlRaw, imageUrlRaw string) *ProductUrls {
	url, error := url.Parse(urlRaw)
	imageUrl, error2 := url.Parse(imageUrlRaw)

	if error == nil && error2 == nil {
		return &ProductUrls{url, imageUrl}
	} else {
		log.Printf("%v %v\n", error, error2)
		return nil
	}
}

func ToProductUrls(urls ...string) []*ProductUrls {
	var products []*ProductUrls;
	for _, rawUrl := range urls {
		url, error := url.Parse(rawUrl)
		if error == nil {
			products = append(products, &ProductUrls{url, nil})
		} else {
			log.Println(error)
		}
	}
	return products
}

func ToProducts(urls ...string) []*Product {
	var products []*Product;
	for _, rawUrl := range urls {
		url, error := url.Parse(rawUrl)
		if error == nil {
			urls := &ProductUrls{url, nil}
			products = append(products, &Product{urls, ""})
		} else {
			log.Println(error)
		}
	}
	return products
}

func ToRankedProducts(urls ...string) []*RankedProduct {
	var products []*RankedProduct;
	for _, rawUrl := range urls {
		url, error := url.Parse(rawUrl)
		if error == nil {
			urls := &ProductUrls{url, nil}
			prod := &Product{urls, ""}
			products = append(products, &RankedProduct{prod, 0})
			} else {
				log.Println(error)
			}
	}
	return products
}
