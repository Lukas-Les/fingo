package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/Lukas-Les/fingo/internal/database"
	"github.com/joho/godotenv"
)

const jwtSecret = "testing-token"

func TestTransaction(t *testing.T) {
	if err := godotenv.Load(".env.test"); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	if err = db.Ping(); err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	dbQueries := database.New(db)
	createUserHandler := BuildUserCreateHandler(dbQueries)
	loginHandler := BuildUserLoginHandler(dbQueries, "testing-token")
	transactionHandler := BuildTransactionCreateHandler(dbQueries, jwtSecret)

	// create user and login
	email := "transaction@example.com"
	password := "pass"

	db.Exec("DELETE FROM users WHERE email = $1", email)

	form := url.Values{}
	form.Set("email", email)
	form.Set("password", password)
	req := httptest.NewRequest("POST", "/api/v1/create-user", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	createUserHandler(rr, req)
	t.Cleanup(func() {
		db.Exec("DELETE FROM users WHERE email = $1", email)
	})

	loginForm := url.Values{}
	loginForm.Set("email", email)
	loginForm.Set("password", password)
	loginReq := httptest.NewRequest("POST", "/api/v1/login", strings.NewReader(loginForm.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRecorder := httptest.NewRecorder()
	loginHandler(loginRecorder, loginReq)
	var bearerToken string
	for _, cookie := range loginRecorder.Result().Cookies() {
		if cookie.Name == "token" {
			bearerToken = cookie.Value
			break
		}
	}

	t.Run("Should create a new transaction", func(t *testing.T) {
		createTransactionForm := url.Values{}
		createTransactionForm.Set("amount", "10")
		createTransactionForm.Set("transaction_date", "2025-01-01")
		createTransactionForm.Set("transaction_type", "income")
		createTransactionForm.Set("category", "test")
		createTransactionForm.Set("description", "test")
		createTransactionForm.Set("party", "test")

		createTransactionReq := httptest.NewRequest("POST", "/api/v1/create-transaction", strings.NewReader(createTransactionForm.Encode()))
		createTransactionReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		createTransactionReq.Header.Set("Authorization", "Bearer "+bearerToken)
		ctRecorder := httptest.NewRecorder()
		transactionHandler(ctRecorder, createTransactionReq)
		t.Cleanup(func() {
			db.Exec("DELETE FROM transactions WHERE category = $1", "test")
		})

		if ctRecorder.Code != http.StatusSeeOther {
			t.Errorf("expected 303, got %d: %s", ctRecorder.Code, ctRecorder.Body.String())
		}
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM transactions WHERE amount = $1 AND category = $2", "10", "test").Scan(&count)
		if err != nil {
			t.Fatalf("db query failed: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 transaction in db, got %d", count)
		}
	})
}
