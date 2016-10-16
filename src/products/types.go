package products

import "net/url"

type ProductUrls struct {
	Url *url.URL
	ImageUrl *url.URL
}

type Product struct {
	Urls  *ProductUrls
	Image string
}

type RankedProduct struct {
	Product *Product
	Rank int
}
