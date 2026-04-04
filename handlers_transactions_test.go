package main

// import (
// 	"database/sql"
// 	"os"
// 	"testing"
//
// 	"github.com/Lukas-Les/fingo/internal/database"
// 	"github.com/joho/godotenv"
// )
//
// func TestBuildTransactionCreateHandler(t *testing.T) {
// 	godotenv.Load(".env.test")
// 	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
// 	if err != nil {
// 		t.Fatalf("Failed to connect to DB: %v", err)
// 	}
// 	defer db.Close()
//
// 	dbQueries := database.New(db)
// 	handler := BuildTransactionCreateHandler(db)
// }
