package main

import (
	"fmt"
	"net/http"

	"github.com/Lukas-Les/fingo/internal/auth"
	"github.com/Lukas-Les/fingo/internal/database"
	"github.com/Lukas-Les/fingo/templates"
	"github.com/gorilla/csrf"

	decimal "github.com/shopspring/decimal"
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
		sumIncome := decimal.Zero
		sumExpense := decimal.Zero
		for _, trx := range userTransactions {
			if trx.TransactionType == database.TransactionTypeEnumIncome {
				sumIncome = sumIncome.Add(trx.Amount)
			}
			if trx.TransactionType == database.TransactionTypeEnumExpense {
				sumExpense = sumExpense.Add(trx.Amount)
			}
		}
		fmt.Printf("sum income %s; sum expense %s", sumIncome.String(), sumExpense.String())
		templates.Dashboard(csrf.Token(r), dbUser.Email, balance, sumIncome.String(), sumExpense.String(), userTransactions).Render(r.Context(), w)
	}
}
