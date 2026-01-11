package handler

type DeliveryResponse struct {
	TotalPrice          int64 `json:"total_price"`
	SlammOrderSurcharge int64 `json:"small_order_surcharge"`
	DeliveryFee         int64 `json:"delivery_fee"`
	DeliveryDistance    int64 `json:"delivery_distance"`
	DeliveryPossible    bool  `json:"delivery_possible"`
}
