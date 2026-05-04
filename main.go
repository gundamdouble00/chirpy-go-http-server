package main

import "net/http"

const port = ":8080"

func main() {
	serveMux := http.NewServeMux()
	server := new(http.Server)
	server.Handler = serveMux
	server.Addr = port
	server.ListenAndServe()
}
