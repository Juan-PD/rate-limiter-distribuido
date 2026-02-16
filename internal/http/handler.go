package http

import (
	"fmt"
	"net/http"
)

// PingHandler — Endpoint simple
func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// HelloHandler — Demo endpoint to show limited resource
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello, your ip: %s\n", r.RemoteAddr)
}
