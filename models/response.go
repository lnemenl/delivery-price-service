package models

// PriceResponse is the final JSON payload we send to the client.
type PriceResponse struct {
	TotalPrice          int             `json:"total_price"`
	SmallOrderSurcharge int             `json:"small_order_surcharge"`
	CartValue           int             `json:"cart_value"`
	Delivery            DeliveryDetails `json:"delivery"`
}

// DeliveryDetails holds the breakdown of the delivery cost
type DeliveryDetails struct {
	Fee      int `json:"fee"`
	Distance int `json:"distance"`
}
