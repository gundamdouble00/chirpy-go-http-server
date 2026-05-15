package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/gundamdouble00/chirpy-go-http-server/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const port = "8080"

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerNumberOfRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	hits := strconv.Itoa(int(cfg.fileserverHits.Load()))
	w.Write([]byte(fmt.Sprintf(`
<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %s times!</p>
	</body>
</html>`, hits)))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = atomic.Int32{}
	// w.Write([]byte("Hits was reset"))
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func isProfaneWord(word string) bool {
	return (word == "fornax") || (word == "kerfuffle") || (word == "sharbert")
}

func cleanUpProfaneWords(msg string) string {
	cleanedWords := []string{}
	words := strings.Split(msg, " ")
	for _, word := range words {
		temp := strings.ToLower(word)
		if isProfaneWord(temp) {
			cleanedWords = append(cleanedWords, "****")
		} else {
			cleanedWords = append(cleanedWords, word)
		}
	}

	cleanedMsg := strings.Join(cleanedWords, " ")
	return cleanedMsg
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, 405, "method not allowed")
		return
	}

	type requestBody struct {
		Body string `json:"body"`
	}

	params := requestBody{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 400, "invalid JSON")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	// respondWithJSON(w, 200, map[string]bool{"valid": true})
	respondWithJSON(w, 200, map[string]string{"cleaned_body": cleanUpProfaneWords(params.Body)})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("couldn't load .env: %w", err)
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("couldn't connect database: %w", err)
	}
	dbQueries := database.New(db)

	hanlder := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	mux := http.NewServeMux()
	apiCfg := &apiConfig{
		dbQueries: dbQueries,
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(hanlder))

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerNumberOfRequests)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Printf("Serving on ports: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
