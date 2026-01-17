package models

// PriceResponse represents the delivery price calculation response
type PriceResponse struct {
	TotalPrice          int             `json:"total_price"`
	SmallOrderSurcharge int             `json:"small_order_surcharge"`
	CartValue           int             `json:"cart_value"`
	Delivery            DeliveryDetails `json:"delivery"`
}

// DeliveryDetails contains fee and distance information
type DeliveryDetails struct {
	Fee      int `json:"fee"`
	Distance int `json:"distance"`
}
