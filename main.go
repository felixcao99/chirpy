package main

import (
	"net/http"
)

func main() {
	serverMux := http.NewServeMux()
	serverMux.Handle("/assets/", http.FileServer(http.Dir(".")))
	serverMux.Handle("/", http.FileServer(http.Dir(".")))

	var server http.Server
	server.Addr = ":8080"
	server.Handler = serverMux
	server.ListenAndServe()
}
