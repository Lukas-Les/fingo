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

	"github.com/Lukas-Les/fingo/internal/database"
	"github.com/joho/godotenv"
)

func TestUser(t *testing.T) {
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

	t.Run("Should create a new user", func(t *testing.T) {
		email := "create@example.com"

		t.Cleanup(func() {
			db.Exec("DELETE FROM users WHERE email = $1", email)
		})

		form := url.Values{}
		form.Set("email", email)
		form.Set("password", "pass")
		req := httptest.NewRequest("POST", "/api/v1/create-user", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		createUserHandler(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("expected status 303, got %d", rr.Code)
		}

		user, err := dbQueries.GetUserByEmail(context.Background(), email)
		if err != nil {
			t.Fatalf("User should have been saved to DB, but wasn't: %v", err)
		}

		if user.Email != email {
			t.Errorf("Expected email %s, got %s", email, user.Email)
		}
	})

	t.Run("Should log in", func(t *testing.T) {
		email := "login@example.com"
		password := "pass"

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

		loginForm := url.Values{}
		loginForm.Set("email", email)
		loginForm.Set("password", password)
		loginReq := httptest.NewRequest("POST", "/api/v1/login", strings.NewReader(loginForm.Encode()))
		loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		loginRr := httptest.NewRecorder()
		loginHandler(loginRr, loginReq)

		if loginRr.Code != http.StatusSeeOther {
			t.Errorf("expected status 303, got %d", loginRr.Code)
		}
		if loginRr.Header().Get("Location") != "/dashboard" {
			t.Errorf("wrong location header")
		}
	})
}
