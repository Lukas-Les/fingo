package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/Lukas-Les/fingo/internal/auth"
	"github.com/Lukas-Les/fingo/internal/database"
	"github.com/shopspring/decimal"
)

type transactionQueries interface {
	CreateTransaction(ctx context.Context, arg database.CreateTransactionParams) (database.Transaction, error)
}

func BuildTransactionCreateHandler(db transactionQueries, jwtSecret string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		userId, err := auth.ValidateJWT(token, jwtSecret)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}

		amount, err := decimal.NewFromString(r.FormValue("amount"))
		if err != nil {
			http.Error(w, "invalid amount", http.StatusBadRequest)
			return
		}

		transactionDate, err := time.Parse("2006-01-02", r.FormValue("transaction_date"))
		if err != nil {
			http.Error(w, "invalid date, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		category := r.FormValue("category")
		description := r.FormValue("description")
		party := r.FormValue("party")
		transactionTypeRaw := r.FormValue("transaction_type")
		if transactionTypeRaw == "" || (transactionTypeRaw != "income" && transactionTypeRaw != "expense") {
			http.Error(w, "missing transaction_type[income | expense]", http.StatusInternalServerError)
			return
		}

		params := database.CreateTransactionParams{
			UserID:          userId,
			Amount:          amount,
			TransactionType: database.TransactionTypeEnum(r.FormValue("transaction_type")),
			Category:        sql.NullString{String: category, Valid: category != ""},
			Description:     sql.NullString{String: description, Valid: description != ""},
			Party:           sql.NullString{String: party, Valid: party != ""},
			TransactionDate: transactionDate,
		}

		_, err = db.CreateTransaction(r.Context(), params)
		if err != nil {
			http.Error(w, "failed to create transaction", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}
