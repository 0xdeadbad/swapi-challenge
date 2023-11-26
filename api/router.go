package api

import "net/http"

func routes() http.Handler {
	mux := http.DefaultServeMux

	mux.HandleFunc("/", homeEndpointHandler)

	return mux
}

func Router() http.Handler {

	handler := routes()

	return handler
}
