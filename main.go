package main

import (
	"fmt"
	"net/http"
)

func main() {
	//Set up the defaults for serving files via urls like /app/foo
	mux := http.NewServeMux()
	file_server := http.FileServer(http.Dir("."))
	mux.Handle("/app/", http.StripPrefix("/app", file_server))

	/*
		A special handler which can be called via /healthz
		This is not a file but code that runs in response to a request
		Examples include api usage, server status etc.
	*/
	handler := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
	mux.HandleFunc("/healthz", handler)

	port := "8080"
	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Errorf("error: Could not serve server")
		return
	}
}
