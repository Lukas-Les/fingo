package main

import (
	// "database/sql"
	"fmt"
	"log"
	"net/http"
	// "os"

	"github.com/Lukas-Les/fingo/internal/database"
)

const (
	port            = "8080"
	defaultErrorMsg = "Something bad happened"
)

type apiConfig struct {
	db        *database.Queries
	env       string
	jwtSecret string
}

func main() {

	// dbURL := os.Getenv("DB_URL")
	// db, err := sql.Open("postgres", dbURL)
	// if err != nil {
	// 	log.Fatalln("failed to connect to the db")
	// }

	// dbQueries := database.New(db)
	// cfg = apiConfig{db: dbQueries, env: os.Getenv("ENV"), jwtSecret: ""}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", handlerHealth)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	filepathRoot := http.Dir(".")
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())

	fmt.Println("Hello, World!")
}

func handlerHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
