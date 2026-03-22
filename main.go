package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Lukas-Les/fingo/internal/database"
	"github.com/a-h/templ"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	port            = "8000"
	defaultErrorMsg = "Something bad happened"
)

type apiConfig struct {
	db        *database.Queries
	env       string
	jwtSecret string
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalln("failed to connect to the db")
	}
	filepathRoot := os.Getenv("FILEPATH_ROOT")
	if filepathRoot == "" {
		log.Fatal("FILEPATH_ROOT environment variable is not set")
	}

	dbQueries := database.New(db)
	cfg := apiConfig{db: dbQueries, env: os.Getenv("ENV"), jwtSecret: ""}

	mux := http.NewServeMux()
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", appHandler)
	mux.HandleFunc("GET /api/v1/live", handlerLive)
	mux.HandleFunc("GET /api/v1/ready", buildHandlerReady(cfg))
	mux.HandleFunc("POST /api/v1/create-user", BuildUserCreateHandler(cfg.db))
	mux.HandleFunc("POST /api/v1/login", BuildUserLoginHandler(cfg.db, cfg.jwtSecret))
	mux.Handle("GET /api/v1/templ", templ.Handler(hello("fingo")))
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	fmt.Printf("server started and serving on: http://localhost:%s/%s\n", port, strings.Split(filepathRoot, "./")[1])
	fmt.Printf("database url:                  %s\n", dbURL)
	log.Fatal(srv.ListenAndServe())

}

func handlerLive(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func buildHandlerReady(cfg apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}
}
