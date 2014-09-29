package main

import (
	"net/url"
	"log"
)

func ToProductUrls(urls ...string) []*ProductUrls {
	var products []*ProductUrls;
	for _, rawUrl := range urls {
		url, err := url.Parse(rawUrl)
		if err == nil {
			products = append(products, &ProductUrls{url, nil})
		} else {
			log.Println(err)
		}
	}
	return products
}
