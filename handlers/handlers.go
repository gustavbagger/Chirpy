package handlers

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

/*
atomic.Int32 is a thread-safe int32 which can be read and incremented
Across multiple go-routines/http requests
*/
type ApiConfig struct {
	fileserverHits atomic.Int32
}

// Counts fileserver hits and increments every time it is called
func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	cfg.fileserverHits.Add(1)
	return next
}

/*
These are not files but code that run in response to a request
Examples include api usage, server status etc.
*/
func HandlerHealthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
func (cfg *ApiConfig) HandlerMetrics(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	str := fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())
	w.Write([]byte(str))
}
func (cfg *ApiConfig) HandlerReset(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	str := fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())
	w.Write([]byte(str))
}
