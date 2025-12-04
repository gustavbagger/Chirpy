package main

import (
	"encoding/json"
	"fmt"
	"log"
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

// think tweet, sent to server
type chirp struct {
	Body string `json:"body"`
}

// json error
type jsonError struct {
	Err string `json:"error"`
}

// json error
type chirpValid struct {
	Valid bool `json:"valid"`
}

/*
Counts fileserver hits and increments every time it is called
makes a http.Handler using a ordinary function
This function needs to take a responsewriter and request to match signatures
The function increments fileserverhits, then does what next normally does
*/
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
	hits := cfg.fileserverHits.Load()
	html := fmt.Sprintf(`
	<html>
  		<body>
    		<h1>Welcome, Chirpy Admin</h1>
    		<p>Chirpy has been visited %d times!</p>
  		</body>
	</html>
	`, hits)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}

func handlerValidate(w http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)
	var chp chirp
	w.Header().Set("Content-Type", "application/json")
	var errorstr string
	if err := decoder.Decode(&chp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorstr = fmt.Sprintf("error: %v", err)
	} else if len(chp.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		errorstr = "error: chirp too long"
	} else {
		w.WriteHeader(http.StatusOK)
	}
	if errorstr != "" {
		data, err := json.Marshal(jsonError{Err: errorstr})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("error: %v", err)
			return
		}
		w.Write(data)
	} else {
		data, err := json.Marshal(chirpValid{Valid: true})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("error: %v", err)
			return
		}
		w.Write(data)
	}
}

var apiCfg apiConfig

func main() {
	//Set up the defaults for serving files via urls like /app/foo
	mux := http.NewServeMux()
	file_server := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	// Includes middleware to count hits
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(file_server))

	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidate)

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
