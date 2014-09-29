package main

import (
	"net/url"
	"log"
)

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
