package main

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Lukas-Les/fingo/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
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

func TestDeleteTransaction(t *testing.T) {
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
	deleteHandler := BuildTransactionDeleteHandler(dbQueries, jwtSecret)

	email := "delete-transaction@example.com"
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

	user, err := dbQueries.GetUserByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("setup: failed to get user from DB: %v", err)
	}

	createTx := func(t *testing.T, userID uuid.UUID) database.Transaction {
		t.Helper()
		tx, err := dbQueries.CreateTransaction(context.Background(), database.CreateTransactionParams{
			UserID:          userID,
			Amount:          decimal.NewFromInt(10),
			TransactionType: database.TransactionTypeEnumIncome,
			TransactionDate: time.Now(),
		})
		if err != nil {
			t.Fatalf("setup: failed to create transaction: %v", err)
		}
		t.Cleanup(func() {
			db.Exec("DELETE FROM transactions WHERE id = $1", tx.ID)
		})
		return tx
	}

	doDelete := func(t *testing.T, id string, token string) *httptest.ResponseRecorder {
		t.Helper()
		f := url.Values{}
		f.Set("id", id)
		r := httptest.NewRequest("POST", "/api/v1/delete-transaction", strings.NewReader(f.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		deleteHandler(w, r)
		return w
	}

	t.Run("valid delete", func(t *testing.T) {
		tx := createTx(t, user.ID)

		w := doDelete(t, tx.ID.String(), bearerToken)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("expected 303, got %d: %s", w.Code, w.Body.String())
		}

		var deletedAt sql.NullTime
		err := db.QueryRow("SELECT deleted_at FROM transactions WHERE id = $1", tx.ID).Scan(&deletedAt)
		if err != nil {
			t.Fatalf("db query failed: %v", err)
		}
		if !deletedAt.Valid {
			t.Error("expected deleted_at to be set, but it was NULL")
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		w := doDelete(t, "not-a-uuid", bearerToken)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("non-existent transaction", func(t *testing.T) {
		w := doDelete(t, uuid.New().String(), bearerToken)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("no auth token", func(t *testing.T) {
		tx := createTx(t, user.ID)

		w := doDelete(t, tx.ID.String(), "")
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("another user's transaction", func(t *testing.T) {
		otherEmail := "delete-transaction-other@example.com"
		db.Exec("DELETE FROM users WHERE email = $1", otherEmail)
		t.Cleanup(func() {
			db.Exec("DELETE FROM users WHERE email = $1", otherEmail)
		})

		otherForm := url.Values{}
		otherForm.Set("email", otherEmail)
		otherForm.Set("password", password)
		otherReq := httptest.NewRequest("POST", "/api/v1/create-user", strings.NewReader(otherForm.Encode()))
		otherReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		createUserHandler(httptest.NewRecorder(), otherReq)

		otherUser, err := dbQueries.GetUserByEmail(context.Background(), otherEmail)
		if err != nil {
			t.Fatalf("setup: failed to get other user: %v", err)
		}

		otherTx := createTx(t, otherUser.ID)

		w := doDelete(t, otherTx.ID.String(), bearerToken)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}
