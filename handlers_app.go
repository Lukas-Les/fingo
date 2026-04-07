package main

import (
	"net/http"

	"github.com/Lukas-Les/fingo/internal/auth"
	"github.com/Lukas-Les/fingo/internal/database"
	"github.com/Lukas-Les/fingo/templates"
	"github.com/gorilla/csrf"
)

func BuildDashboardHandler(db *database.Queries, jwtSecret string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetTokenFromCookie(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		userId, err := auth.ValidateJWT(token, jwtSecret)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		dbUser, err := db.GetUser(r.Context(), userId)
		if err != nil {
			http.Error(w, "no such user", http.StatusUnauthorized)
			return
		}
		balance, err := db.GetUserBalanceAsStr(r.Context(), userId)
		if err != nil {
			http.Error(w, "failed to fetch balance", http.StatusUnauthorized)
			return
		}
		userTransactions, err := db.GetUserTransactions(r.Context(), userId)
		if err != nil {
			http.Error(w, "failed to fetch transactions", http.StatusUnauthorized)
			return
		}
		templates.Dashboard(csrf.Token(r), dbUser.Email, balance, userTransactions).Render(r.Context(), w)
	}
}
