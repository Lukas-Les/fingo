package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Lukas-Les/fingo/internal/auth"
	"github.com/Lukas-Les/fingo/internal/database"
)

type UserCreator interface {
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error)
}

func BuildUserCreateHandler(db UserCreator) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		type parameters struct {
			Password string `json:"password"`
			Email    string `json:"email"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
			return
		}

		if params.Password == "" || params.Email == "" {
			respondWithError(w, http.StatusBadRequest, "Email and password are required", nil)
			return
		}

		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
			return
		}

		user, err := db.CreateUser(r.Context(), database.CreateUserParams{
			Email:          params.Email,
			HashedPassword: hashedPassword,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
			return
		}

		respondWithJSON(w, http.StatusCreated, user)
	}
}

// func BuildUserLoginHandler(db UserCreator) func(http.ResponseWriter, *http.Request) {
//
// }
