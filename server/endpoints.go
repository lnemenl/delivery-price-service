package server

import (
	"net/http"
)

type Router struct {
	mux      *http.ServeMux
	handlers *Handlers
}

func NewRouter(mux *http.ServeMux, handlers *Handlers) *Router {
	return &Router{
		mux:      mux,
		handlers: handlers,
	}
}

func (r *Router) Setup() {
	r.setupDeliveryRoutes()
}

func (r *Router) setupDeliveryRoutes() {
	r.mux.HandleFunc("/api/v1/delivery-order-price", r.handlers.Delivery.HandleRequest)
}
