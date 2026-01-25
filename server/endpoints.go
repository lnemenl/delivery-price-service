package server

import (
	"net/http"
)

type Router struct {
	mux     *http.ServeMux
	handler *PriceHandler
}

func NewRouter(mux *http.ServeMux, h *PriceHandler) *Router {
	return &Router{
		mux:     mux,
		handler: h,
	}
}

func (r *Router) Setup() {
	r.setupDeliveryRoutes()
}

func (r *Router) setupDeliveryRoutes() {
	r.mux.HandleFunc("/api/v1/delivery-order-price", r.handler.HandleRequest)
}
