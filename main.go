package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/gustavbagger/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/*
atomic.Int32 is a thread-safe int32 which can be read and incremented
Across multiple go-routines/http requests
*/
type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
}

// think tweet, sent to server
type chirp struct {
	Body string `json:"body"`
}

// json error
type jsonError struct {
	Err string `json:"error"`
}

// cleaned chirp
type chirpClean struct {
	CleanedBody string `json:"cleaned_body"`
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, err := json.Marshal(jsonError{Err: msg})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error: %v", err)
		return
	}
	w.Write(data)
}
func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error: %v", err)
		return
	}
	w.Write(data)
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
	if err := decoder.Decode(&chp); err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintln(err))
		return
	}
	if len(chp.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "chirp too long")
		return
	}
	words := strings.Split(chp.Body, " ")
	bad_words := []string{"kerfuffle", "sharbert", "fornax"}
	var cleaned_words []string
	for _, word := range words {
		if slices.Contains(bad_words, strings.ToLower(word)) {
			cleaned_words = append(cleaned_words, "****")
		} else {
			cleaned_words = append(cleaned_words, word)
		}
	}
	payload := chirpClean{CleanedBody: strings.Join(cleaned_words, " ")}

	respondWithJSON(w, http.StatusOK, payload)

}

var apiCfg apiConfig

func main() {
	//Load env file into environment variables
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	// open connection to database
	db, err1 := sql.Open("postgres", dbURL)
	if err1 != nil {
		log.Println(err1)
		return
	}
	apiCfg.dbQueries = database.New(db)

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
	err2 := server.ListenAndServe()
	if err2 != nil {
		fmt.Errorf("error: Could not serve server")
		return
	}
}
