package api

import (
	"fmt"
	"net/http"
)

func homeEndpointHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love Go! This is the home page.")
}
