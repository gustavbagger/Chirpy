package main

import (
	"fmt"
	"net/http"

	"github.com/gustavbagger/Chirpy/handlers"
)

func main() {
	//Set up the defaults for serving files via urls like /app/foo
	mux := http.NewServeMux()
	file_server := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	//Includes middleware to count hits
	var apiCfg handlers.ApiConfig
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(file_server))

	mux.HandleFunc("/healthz", handlers.HandlerHealthz)
	mux.HandleFunc("/metrics", apiCfg.HandlerMetrics)

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
