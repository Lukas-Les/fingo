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
	loginHandler := BuildUserLoginHandler(dbQueries, jwtSecret)
	transactionHandler := BuildTransactionCreateHandler(dbQueries, jwtSecret)

	email := "transaction@example.com"
	password := "pass"

	db.Exec("DELETE FROM users WHERE email = $1", email)
	t.Cleanup(func() {
		db.Exec("DELETE FROM users WHERE email = $1", email)
	})

	form := url.Values{}
	form.Set("email", email)
	form.Set("password", password)
	req := httptest.NewRequest("POST", "/api/v1/create-user", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	createUserHandler(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("setup: failed to create user, got %d: %s", rr.Code, rr.Body.String())
	}

	loginForm := url.Values{}
	loginForm.Set("email", email)
	loginForm.Set("password", password)
	loginReq := httptest.NewRequest("POST", "/api/v1/login", strings.NewReader(loginForm.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRecorder := httptest.NewRecorder()
	loginHandler(loginRecorder, loginReq)
	if loginRecorder.Code != http.StatusSeeOther {
		t.Fatalf("setup: failed to login, got %d: %s", loginRecorder.Code, loginRecorder.Body.String())
	}

	var bearerToken string
	for _, cookie := range loginRecorder.Result().Cookies() {
		if cookie.Name == "token" {
			bearerToken = cookie.Value
			break
		}
	}
	if bearerToken == "" {
		t.Fatal("setup: login did not return a token cookie")
	}

	cases := []struct {
		name         string
		amount       string
		date         string
		transType    string
		expectedCode int
	}{
		{"valid income", "10", "2025-01-01", "income", http.StatusSeeOther},
		{"valid expense", "50", "2025-01-02", "expense", http.StatusSeeOther},
		{"invalid amount", "abc", "2025-01-01", "income", http.StatusBadRequest},
		{"missing amount", "", "2025-01-01", "income", http.StatusBadRequest},
		{"invalid date", "10", "01-01-2025", "income", http.StatusBadRequest},
		{"missing date", "10", "", "income", http.StatusBadRequest},
		{"invalid type", "10", "2025-01-01", "bad", http.StatusInternalServerError},
		{"missing type", "10", "2025-01-01", "", http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				db.Exec("DELETE FROM transactions WHERE category = $1", "test")
			})

			f := url.Values{}
			f.Set("amount", tc.amount)
			f.Set("transaction_date", tc.date)
			f.Set("transaction_type", tc.transType)
			f.Set("category", "test")
			f.Set("description", "test")
			f.Set("party", "test")

			r := httptest.NewRequest("POST", "/api/v1/create-transaction", strings.NewReader(f.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r.Header.Set("Authorization", "Bearer "+bearerToken)
			w := httptest.NewRecorder()
			transactionHandler(w, r)

			if w.Code != tc.expectedCode {
				t.Errorf("expected %d, got %d: %s", tc.expectedCode, w.Code, w.Body.String())
			}

			if tc.expectedCode == http.StatusSeeOther {
				var count int
				err := db.QueryRow(
					"SELECT COUNT(*) FROM transactions WHERE amount = $1 AND category = $2",
					tc.amount, "test",
				).Scan(&count)
				if err != nil {
					t.Fatalf("db query failed: %v", err)
				}
				if count != 1 {
					t.Errorf("expected 1 transaction in db, got %d", count)
				}
			}
		})
	}
}
