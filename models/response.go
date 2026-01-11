package models

// PriceResponse is the JSON we send back to the client
type PriceResponse struct {
	TotalPrice          int            `json:"total_price"`
	SmallOrderSurcharge int            `json:"small_order_surcharge"`
	CartValue           int            `json:"cart_value"`
	Delivery            DeliveryDetail `json:"delivery"`
}

type DeliveryDetail struct {
	Fee      int `json:"fee"`
	Distance int `json:"distance"`
}
