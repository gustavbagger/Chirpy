package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

/*
atomic.Int32 is a thread-safe int32 which can be read and incremented
Across multiple go-routines/http requests
*/
type apiConfig struct {
	fileserverHits atomic.Int32
}

// Counts fileserver hits and increments every time it is called
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

/*
These are not files but code that run in response to a request
Examples include api usage, server status etc.
*/
func handlerHealthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	str := fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())
	w.Write([]byte(str))
}
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}

var apiCfg apiConfig

func main() {
	//Set up the defaults for serving files via urls like /app/foo
	mux := http.NewServeMux()
	file_server := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	// Includes middleware to count hits
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(file_server))

	mux.HandleFunc("/reset", apiCfg.handlerReset)
	mux.HandleFunc("/healthz", handlerHealthz)
	mux.HandleFunc("/metrics", apiCfg.handlerMetrics)

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
