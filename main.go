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

	"github.com/gorilla/csrf"
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
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		templates.Login(csrf.Token(r)).Render(r.Context(), w)
	})

	mux.HandleFunc("GET /signup", func(w http.ResponseWriter, r *http.Request) {
		templates.Signup(csrf.Token(r)).Render(r.Context(), w)
	})
	mux.HandleFunc("GET /dashboard", func(w http.ResponseWriter, r *http.Request) {
		templates.Dashboard(csrf.Token(r)).Render(r.Context(), w)
	})
	mux.HandleFunc("GET /transaction/new", func(w http.ResponseWriter, r *http.Request) {
		templates.NewTransaction(csrf.Token(r)).Render(r.Context(), w)
	})

	mux.HandleFunc("GET /api/v1/live", handlerLive)
	mux.HandleFunc("GET /api/v1/ready", buildHandlerReady(cfg))

	mux.HandleFunc("POST /api/v1/create-user", BuildUserCreateHandler(cfg.db))
	mux.HandleFunc("POST /api/v1/login", BuildUserLoginHandler(cfg.db, cfg.jwtSecret))
	mux.HandleFunc("POST /api/v1/logout", UserLogoutHandler)

	mux.HandleFunc("POST /api/v1/create-transaction", BuildTransactionCreateHandler(cfg.db, cfg.jwtSecret))

	fmt.Printf("server started and serving on: http://localhost:%s/\n", port)
	fmt.Printf("database url:                  %s\n", dbURL)
	csrfMiddleware := csrf.Protect([]byte(os.Getenv("CSRF_SECRET")), csrf.Secure(false), csrf.TrustedOrigins([]string{"localhost:" + port}))
	log.Fatal(http.ListenAndServe(":"+port, csrfMiddleware(mux)))

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
