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

		// TODO: move this to a validateCreds middleware
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
		// TODO: move this to a validateCreds middleware

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

func BuildUserLoginHandler(db database.Queries) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO: move this to a validateCreds middleware
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
		// TODO: move this to a validateCreds middleware

		// TODO: move this to a validatePassword milldeware
		user, err := db.GetUserByEmail(r.Context(), params.Email)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error(), err)
			return
		}
		isValid, err := auth.CheckPasswordHash(user.HashedPassword, hashedPassword)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error(), err)
			return
		}
		if !isValid {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		}
		// TODO: move this to a validatePassword milldeware

		// TODO: handle JWT
	}
}
