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

func TestBuildUserCreateHandler(t *testing.T) {
	godotenv.Load(".env.test")
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	dbQueries := database.New(db)
	handler := BuildUserCreateHandler(dbQueries)

	t.Run("Should create a new user", func(t *testing.T) {
		email := "test@example.com"

		t.Cleanup(func() {
			db.Exec("DELETE FROM users WHERE email = $1", email)
		})

		form := url.Values{}
		form.Set("email", email)
		form.Set("password", "pass")
		req := httptest.NewRequest("POST", "/api/v1/create-user", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler(rr, req)

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
}
