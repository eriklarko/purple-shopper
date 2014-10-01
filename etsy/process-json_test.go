package etsy

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParseJson(t *testing.T) {
json := `{
	  "count": 48620,
	  "results": [
		{
		  "listing_id": 183382994,
		  "title": "baby boy birth announcement with map oh the places he&#39;ll go hello world photo, printable, digital file (item 410)",
		  "price": "13.00",
		  "url": "https:\/\/www.etsy.com\/listing\/183382994\/baby-boy-birth-announcement-with-map-oh?utm_source=purpleshopper&utm_medium=api&utm_campaign=api",
		  "MainImage": {
			"url_170x135": "https:\/\/img1.etsystatic.com\/023\/0\/6093699\/il_170x135.579210365_kriu.jpg"
		  },
		  "ShippingInfo": [
			{
			  "destination_country_name": "Everywhere Else",
			  "primary_cost": "10.00",
			  "secondary_cost": "20.00"
			}
		  ]
		}
	  ],
	  "params": {
		"limit": "1",
		"offset": 0,
		"page": null,
		"keywords": "hello",
		"sort_on": "created",
		"sort_order": "down",
		"min_price": "0.000000",
		"max_price": "100.000000",
		"color": null,
		"color_accuracy": 0,
		"tags": null,
		"category": null,
		"location": null,
		"lat": null,
		"lon": null,
		"region": null,
		"geo_level": "city",
		"accepts_gift_cards": "false",
		"translate_keywords": "false"
	  },
	  "type": "Listing",
	  "pagination": {
		"effective_limit": 1,
		"effective_offset": 0,
		"next_offset": 1,
		"effective_page": 1,
		"next_page": 2
	  }
	}`
	response := parseAsJson([]byte(json))

	assert.Equal(t, 48620, response.Count)
	assert.Equal(t, 1, len(response.Results))
	assert.Equal(t, "13.00", response.Results[0].Price)
	assert.Equal(t, "https://www.etsy.com/listing/183382994/baby-boy-birth-announcement-with-map-oh?utm_source=purpleshopper&utm_medium=api&utm_campaign=api", response.Results[0].Url)
	assert.Equal(t, "https://img1.etsystatic.com/023/0/6093699/il_170x135.579210365_kriu.jpg", response.Results[0].MainImage.Url_170x135)
	assert.Equal(t, "Everywhere Else", response.Results[0].ShippingInfo[0].Destination_country_name)
	assert.Equal(t, "10.00", response.Results[0].ShippingInfo[0].Primary_cost)
	assert.Equal(t, "20.00", response.Results[0].ShippingInfo[0].Secondary_cost)
	assert.Equal(t, 1, *response.Pagination.Next_offset)
}

func TestParseJsonNullNextOffset(t *testing.T) {
	json := `{
	  "count": 48620,
	  "results": [
		{
		  "listing_id": 183382994,
		  "title": "baby boy birth announcement with map oh the places he&#39;ll go hello world photo, printable, digital file (item 410)",
		  "price": "13.00",
		  "url": "https:\/\/www.etsy.com\/listing\/183382994\/baby-boy-birth-announcement-with-map-oh?utm_source=purpleshopper&utm_medium=api&utm_campaign=api",
		  "MainImage": {
			"url_170x135": "https:\/\/img1.etsystatic.com\/023\/0\/6093699\/il_170x135.579210365_kriu.jpg"
		  },
		  "ShippingInfo": [
			{
			  "destination_country_name": "Everywhere Else",
			  "primary_cost": "10.00",
			  "secondary_cost": "20.00"
			}
		  ]
		}
	  ],
	  "params": {
		"limit": "1",
		"offset": 0,
		"page": null,
		"keywords": "hello",
		"sort_on": "created",
		"sort_order": "down",
		"min_price": "0.000000",
		"max_price": "100.000000",
		"color": null,
		"color_accuracy": 0,
		"tags": null,
		"category": null,
		"location": null,
		"lat": null,
		"lon": null,
		"region": null,
		"geo_level": "city",
		"accepts_gift_cards": "false",
		"translate_keywords": "false"
	  },
	  "type": "Listing",
	  "pagination": {
		"effective_limit": 1,
		"effective_offset": 0,
		"next_offset": null,
		"effective_page": 1,
		"next_page": 2
	  }
	}`
	response := parseAsJson([]byte(json))
	assert.Nil(t, response.Pagination.Next_offset)
}
