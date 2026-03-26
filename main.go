package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Lukas-Les/fingo/internal/database"
	"github.com/Lukas-Les/fingo/templates"
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
	dbQueries := database.New(db)
	cfg := apiConfig{db: dbQueries, env: os.Getenv("ENV"), jwtSecret: ""}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	mux.Handle("GET /", templ.Handler(templates.Index()))
	mux.Handle("GET /login", templ.Handler(templates.Login()))
	mux.Handle("GET /signup", templ.Handler(templates.Signup()))
	mux.Handle("GET /dashboard", templ.Handler(templates.Dashboard()))
	mux.HandleFunc("GET /api/v1/live", handlerLive)
	mux.HandleFunc("GET /api/v1/ready", buildHandlerReady(cfg))
	mux.HandleFunc("POST /api/v1/create-user", BuildUserCreateHandler(cfg.db))
	mux.HandleFunc("POST /api/v1/login", BuildUserLoginHandler(cfg.db, cfg.jwtSecret))
	mux.HandleFunc("POST /api/v1/logout", UserLogoutHandler)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	fmt.Printf("server started and serving on: http://localhost:%s/\n", port)
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
